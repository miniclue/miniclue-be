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
	"github.com/rs/zerolog"
)

// ErrUploadLimitExceeded is returned when a user has reached their upload limit.
var ErrUploadLimitExceeded = repository.ErrUploadLimitExceeded

// LectureService defines lecture-related operations
// GetLecturesByCourseID retrieves lectures for a given course with pagination
type LectureService interface {
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLectureWithPDF(ctx context.Context, courseID, userID, title string, file multipart.File, header *multipart.FileHeader) (*model.Lecture, error)
	GetPresignedURL(ctx context.Context, storagePath string) (string, error)
	CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error)
	// CheckAndRecordUpload atomically enforces and records an upload event.
	CheckAndRecordUpload(ctx context.Context, userID string, start, end time.Time, maxUploads int) error
}

// lectureService is the implementation of LectureService
type lectureService struct {
	repo           repository.LectureRepository
	userRepo       repository.UserRepository
	usageRepo      repository.UsageRepository
	s3Client       *s3.Client
	bucketName     string
	publisher      pubsub.Publisher
	ingestionTopic string
	lectureLogger  zerolog.Logger
}

// NewLectureService creates a new LectureService
func NewLectureService(
	repo repository.LectureRepository,
	userRepo repository.UserRepository,
	usageRepo repository.UsageRepository,
	s3Client *s3.Client,
	bucketName string,
	publisher pubsub.Publisher,
	ingestionTopic string,
	logger zerolog.Logger,
) LectureService {
	return &lectureService{
		repo:           repo,
		userRepo:       userRepo,
		usageRepo:      usageRepo,
		s3Client:       s3Client,
		bucketName:     bucketName,
		publisher:      publisher,
		ingestionTopic: ingestionTopic,
		lectureLogger:  logger.With().Str("service", "LectureService").Logger(),
	}
}

// GetLecturesByCourseID retrieves lectures for a given course with pagination
func (s *lectureService) GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error) {
	lectures, err := s.repo.GetLecturesByCourseID(ctx, courseID, limit, offset)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("course_id", courseID).Msg("Failed to get lectures by course ID")
		return nil, err
	}
	return lectures, nil
}

// GetLectureByID retrieves a lecture by ID
func (s *lectureService) GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error) {
	lecture, err := s.repo.GetLectureByID(ctx, lectureID)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to get lecture by ID")
		return nil, err
	}
	return lecture, nil
}

// DeleteLecture removes a lecture by ID and cleans up external resources.
func (s *lectureService) DeleteLecture(ctx context.Context, lectureID string) error {
	// Retrieve lecture metadata for cleanup
	lecture, err := s.repo.GetLectureByID(ctx, lectureID)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to get lecture for deletion")
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
			s.lectureLogger.Error().Err(err).Str("prefix", prefix).Msg("Failed to list S3 objects for deletion")
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
			s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to delete objects from S3")
			// This is not a fatal error, we can still proceed to delete the DB record.
		}
	}

	// Delete lecture and cascade database cleanup
	if err := s.repo.DeleteLecture(ctx, lectureID); err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to delete lecture from database")
		return err
	}
	return nil
}

// UpdateLecture applies title and accessed_at changes to a lecture
func (s *lectureService) UpdateLecture(ctx context.Context, l *model.Lecture) error {
	if err := s.repo.UpdateLecture(ctx, l); err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", l.ID).Msg("Failed to update lecture")
		return err
	}
	return nil
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
		s.lectureLogger.Error().Err(err).Msg("Failed to create lecture record")
		return nil, fmt.Errorf("failed to create lecture record: %w", err)
	}

	// 2. Upload PDF to S3
	storagePath := fmt.Sprintf("lectures/%s/original.pdf", createdLecture.ID)
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		// Cleanup on failure
		_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
		s.lectureLogger.Error().Err(err).Str("lecture_id", createdLecture.ID).Msg("Failed to read file into buffer")
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
		s.lectureLogger.Error().Err(err).Str("storage_path", storagePath).Msg("Failed to upload PDF to S3")
		return nil, fmt.Errorf("failed to upload pdf to s3: %w", err)
	}

	// 3. Store storage path instead of full URL
	createdLecture.StoragePath = storagePath
	// After upload, mark as pending further processing
	createdLecture.Status = "pending_processing"
	if err := s.repo.UpdateLecture(ctx, createdLecture); err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", createdLecture.ID).Msg("Failed to update lecture with storage path and status")
		return nil, fmt.Errorf("failed to update lecture with pdf url and status: %w", err)
	}

	// 4. Publish ingestion job to Pub/Sub, enriched with user info for tracking
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch user for ingestion")
	}
	// Default customer identifier is the user ID
	customerIdentifier := userID
	name := ""
	email := ""
	if user != nil {
		name = user.Name
		email = user.Email
	}
	payload := struct {
		LectureID          string `json:"lecture_id"`
		StoragePath        string `json:"storage_path"`
		CustomerIdentifier string `json:"customer_identifier"`
		Name               string `json:"name"`
		Email              string `json:"email"`
	}{
		LectureID:          createdLecture.ID,
		StoragePath:        storagePath,
		CustomerIdentifier: customerIdentifier,
		Name:               name,
		Email:              email,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", createdLecture.ID).Msg("Failed to marshal ingestion payload")
		// Don't return an error here, the lecture is created, but ingestion won't start automatically.
	} else {
		if _, err := s.publisher.Publish(ctx, s.ingestionTopic, data); err != nil {
			s.lectureLogger.Error().Err(err).Str("topic", s.ingestionTopic).Msg("Failed to publish ingestion job")
			// Don't return an error here either.
		}
	}

	// Record upload event for usage tracking
	if err := s.usageRepo.RecordUploadEvent(ctx, userID); err != nil {
		s.lectureLogger.Error().Err(err).Msg("Failed to record upload event")
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
		s.lectureLogger.Error().Err(err).Str("storage_path", storagePath).Msg("Failed to generate presigned URL")
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return resp.URL, nil
}

// CheckAndRecordUpload enforces the upload limit and logs the event
func (s *lectureService) CheckAndRecordUpload(ctx context.Context, userID string, start, end time.Time, maxUploads int) error {
	return s.usageRepo.CheckAndRecordUpload(ctx, userID, start, end, maxUploads)
}

// CountLecturesByUserInTimeRange delegates to repository
func (s *lectureService) CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error) {
	return s.usageRepo.CountUploadEvents(ctx, userID, start, end)
}
