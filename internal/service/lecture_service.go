package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"app/internal/model"
	"app/internal/pubsub"
	"app/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// LectureService defines lecture-related operations
// GetLecturesByCourseID retrieves lectures for a given course with pagination
type LectureService interface {
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLectureWithPDF(ctx context.Context, courseID, userID, title string, file multipart.File, header *multipart.FileHeader) (*model.Lecture, error)
	GetPresignedURL(ctx context.Context, storagePath string) (string, error)
}

// lectureService is the implementation of LectureService
type lectureService struct {
	repo           repository.LectureRepository
	s3Client       *s3.Client
	bucketName     string
	publisher      pubsub.Publisher
	ingestionTopic string
}

// NewLectureService creates a new LectureService
func NewLectureService(repo repository.LectureRepository, s3Client *s3.Client, bucketName string, publisher pubsub.Publisher, ingestionTopic string) LectureService {
	return &lectureService{
		repo:           repo,
		s3Client:       s3Client,
		bucketName:     bucketName,
		publisher:      publisher,
		ingestionTopic: ingestionTopic,
	}
}

// GetLecturesByCourseID retrieves lectures for a given course with pagination
func (s *lectureService) GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error) {
	return s.repo.GetLecturesByCourseID(ctx, courseID, limit, offset)
}

// GetLectureByID retrieves a lecture by ID
func (s *lectureService) GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error) {
	return s.repo.GetLectureByID(ctx, lectureID)
}

// DeleteLecture removes a lecture by ID and cleans up external resources.
func (s *lectureService) DeleteLecture(ctx context.Context, lectureID string) error {
	// Retrieve lecture metadata for cleanup
	lecture, err := s.repo.GetLectureByID(ctx, lectureID)
	if err != nil {
		return fmt.Errorf("failed to get lecture: %w", err)
	}
	if lecture == nil {
		return fmt.Errorf("lecture not found")
	}

	// Delete all objects under the lecture's storage folder from S3
	prefix := fmt.Sprintf("lectures/%s/", lectureID)
	paginator := s3.NewListObjectsV2Paginator(s.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	var toDelete []types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("failed to list objects for deletion: %v\n", err)
			break
		}
		for _, obj := range page.Contents {
			toDelete = append(toDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}
	if len(toDelete) > 0 {
		if _, err := s.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucketName),
			Delete: &types.Delete{Objects: toDelete, Quiet: aws.Bool(true)},
		}); err != nil {
			fmt.Printf("failed to delete objects from storage: %v\n", err)
		}
	}

	// Clear any pending jobs from all related queues
	if err := s.repo.DeletePendingJobs(ctx, lectureID); err != nil {
		fmt.Printf("failed to delete pending jobs: %v\n", err)
	}

	// Delete lecture and cascade database cleanup
	return s.repo.DeleteLecture(ctx, lectureID)
}

// UpdateLecture applies title and accessed_at changes to a lecture
func (s *lectureService) UpdateLecture(ctx context.Context, l *model.Lecture) error {
	return s.repo.UpdateLecture(ctx, l)
}

func (s *lectureService) CreateLectureWithPDF(ctx context.Context, courseID, userID, title string, file multipart.File, header *multipart.FileHeader) (*model.Lecture, error) {
	// 1. Create lecture record
	lecture := &model.Lecture{
		CourseID: courseID,
		UserID:   userID,
		Title:    title,
		Status:   "uploading",
	}
	createdLecture, err := s.repo.CreateLecture(ctx, lecture)
	if err != nil {
		return nil, fmt.Errorf("failed to create lecture record: %w", err)
	}

	// 2. Upload PDF to S3
	storagePath := fmt.Sprintf("lectures/%s/original.pdf", createdLecture.ID)
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		// Cleanup on failure
		_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
		return nil, fmt.Errorf("failed to read file into buffer: %w", err)
	}

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(storagePath),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		// Cleanup on failure
		_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
		return nil, fmt.Errorf("failed to upload pdf to s3: %w", err)
	}

	// 3. Store storage path instead of full URL
	createdLecture.StoragePath = storagePath
	// After upload, mark as pending further processing
	createdLecture.Status = "pending_processing"
	if err := s.repo.UpdateLecture(ctx, createdLecture); err != nil {
		return nil, fmt.Errorf("failed to update lecture with pdf url and status: %w", err)
	}

	// 4. Publish ingestion job to Pub/Sub
	payload := struct {
		LectureID   string `json:"lecture_id"`
		StoragePath string `json:"storage_path"`
	}{
		LectureID:   createdLecture.ID,
		StoragePath: storagePath,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("failed to marshal ingestion payload: %v\n", err)
	} else {
		if _, err := s.publisher.Publish(ctx, s.ingestionTopic, data); err != nil {
			fmt.Printf("failed to publish ingestion job: %v\n", err)
		}
	}

	return createdLecture, nil
}

// GetPresignedURL generates a signed URL for the given storage path
func (s *lectureService) GetPresignedURL(ctx context.Context, storagePath string) (string, error) {
	presigner := s3.NewPresignClient(s.s3Client)
	resp, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(storagePath),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return resp.URL, nil
}
