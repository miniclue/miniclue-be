package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"app/internal/model"
	"app/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// LectureService defines lecture-related operations
// GetLecturesByCourseID retrieves lectures for a given course with pagination
type LectureService interface {
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLectureWithPDF(ctx context.Context, courseID, userID, title string, file multipart.File, header *multipart.FileHeader) (*model.Lecture, error)
}

// lectureService is the implementation of LectureService
type lectureService struct {
	repo       repository.LectureRepository
	s3Client   *s3.Client
	bucketName string
}

// NewLectureService creates a new LectureService
func NewLectureService(repo repository.LectureRepository, s3Client *s3.Client, bucketName string) LectureService {
	return &lectureService{
		repo:       repo,
		s3Client:   s3Client,
		bucketName: bucketName,
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

// DeleteLecture removes a lecture by ID
func (s *lectureService) DeleteLecture(ctx context.Context, lectureID string) error {
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
		Status:   "uploaded",
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

	// 3. Update lecture with PDF URL and status
	pdfURL := fmt.Sprintf("%s/%s/%s", aws.ToString(s.s3Client.Options().BaseEndpoint), s.bucketName, storagePath)
	createdLecture.PDFURL = pdfURL
	createdLecture.Status = "uploaded"
	if err := s.repo.UpdateLecture(ctx, createdLecture); err != nil {
		return nil, fmt.Errorf("failed to update lecture with pdf url: %w", err)
	}

	// 4. Enqueue ingestion job
	if err := s.repo.EnqueueIngestionJob(ctx, createdLecture.ID, storagePath); err != nil {
		// Log, but don't fail the request
		fmt.Printf("failed to enqueue ingestion job: %v", err)
	}

	return createdLecture, nil
}
