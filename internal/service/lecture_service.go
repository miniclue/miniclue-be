package service

import (
	"context"
	"encoding/json"
	"fmt"
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

	GetPresignedURL(ctx context.Context, storagePath string) (string, error)

	// RecordUploadEvent records a new upload event without checking quota limits.
	RecordUploadEvent(ctx context.Context, userID string, lectureID string) error
	// CountLecturesByUserInTimeRange counts lecture uploads in the given period.
	CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error)
	InitiateUpload(ctx context.Context, courseID, userID, filename string) (*model.Lecture, string, error)
	InitiateBatchUpload(ctx context.Context, courseID, userID string, filenames []string) ([]*model.Lecture, []string, error)
	CompleteUpload(ctx context.Context, lectureID, userID string) (*model.Lecture, error)
}

// lectureService is the implementation of LectureService
type lectureService struct {
	repo           repository.LectureRepository
	userRepo       repository.UserRepository
	usageRepo      repository.UsageRepository
	s3Client       *s3.Client
	presignClient  *s3.PresignClient
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
		presignClient:  s3.NewPresignClient(s3Client),
		bucketName:     bucketName,
		publisher:      publisher,
		ingestionTopic: ingestionTopic,
		lectureLogger:  logger.With().Str("service", "LectureService").Logger(),
	}
}

// InitiateUpload creates a lecture record and returns a presigned URL for upload.
func (s *lectureService) InitiateUpload(ctx context.Context, courseID, userID, filename string) (*model.Lecture, string, error) {
	// 1. Create lecture record with 'uploading' status
	lecture := &model.Lecture{
		CourseID: courseID,
		UserID:   userID,
		Title:    filename, // Use filename as the initial title
		Status:   "uploading",
	}
	createdLecture, err := s.repo.CreateLecture(ctx, lecture)
	if err != nil {
		s.lectureLogger.Error().Err(err).Msg("Failed to create lecture record for upload")
		return nil, "", fmt.Errorf("failed to create lecture record: %w", err)
	}

	// 2. Generate presigned URL for direct S3 upload
	storagePath := fmt.Sprintf("lectures/%s/original.pdf", createdLecture.ID)
	presignedURL, err := s.getPresignedPutURL(ctx, storagePath)
	if err != nil {
		// Attempt to clean up the created lecture record on failure
		_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
		s.lectureLogger.Error().Err(err).Str("lecture_id", createdLecture.ID).Msg("Failed to generate presigned PUT URL")
		return nil, "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// 3. Update lecture with the storage path (but still 'uploading')
	createdLecture.StoragePath = storagePath
	if err := s.repo.UpdateLecture(ctx, createdLecture); err != nil {
		_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
		s.lectureLogger.Error().Err(err).Str("lecture_id", createdLecture.ID).Msg("Failed to update lecture with storage path")
		return nil, "", fmt.Errorf("failed to update lecture with storage path: %w", err)
	}

	return createdLecture, presignedURL, nil
}

// InitiateBatchUpload creates multiple lecture records and returns presigned URLs for batch upload.
func (s *lectureService) InitiateBatchUpload(ctx context.Context, courseID, userID string, filenames []string) ([]*model.Lecture, []string, error) {
	if len(filenames) == 0 {
		return nil, nil, fmt.Errorf("no filenames provided")
	}
	if len(filenames) > 10 {
		return nil, nil, fmt.Errorf("too many files: maximum 10 allowed")
	}

	var lectures []*model.Lecture
	var presignedURLs []string

	// Process each filename
	for _, filename := range filenames {
		lecture, presignedURL, err := s.InitiateUpload(ctx, courseID, userID, filename)
		if err != nil {
			// If any file fails, clean up any successfully created lectures
			for _, createdLecture := range lectures {
				_ = s.repo.DeleteLecture(ctx, createdLecture.ID)
			}
			s.lectureLogger.Error().Err(err).Str("filename", filename).Msg("Failed to initiate upload for file in batch")
			return nil, nil, fmt.Errorf("failed to initiate upload for %s: %w", filename, err)
		}

		lectures = append(lectures, lecture)
		presignedURLs = append(presignedURLs, presignedURL)
	}

	return lectures, presignedURLs, nil
}

// CompleteUpload finalizes the upload, updates the lecture status, and triggers the processing pipeline.
func (s *lectureService) CompleteUpload(ctx context.Context, lectureID, userID string) (*model.Lecture, error) {
	// 1. Retrieve the lecture
	lecture, err := s.repo.GetLectureByID(ctx, lectureID)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to get lecture for completion")
		return nil, fmt.Errorf("failed to retrieve lecture: %w", err)
	}
	if lecture == nil {
		return nil, fmt.Errorf("lecture not found")
	}
	if lecture.UserID != userID {
		return nil, fmt.Errorf("user does not own this lecture")
	}

	// Optional: Verify the object exists in S3 before proceeding
	_, err = s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(lecture.StoragePath),
	})
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("storage_path", lecture.StoragePath).Msg("File not found in S3 at expected path")
		lecture.Status = "failed"
		_ = s.repo.UpdateLecture(ctx, lecture) // Mark as failed
		return nil, fmt.Errorf("file not found in storage: %w", err)
	}

	// 2. Update status to 'pending_processing'
	lecture.Status = "pending_processing"
	if err := s.repo.UpdateLecture(ctx, lecture); err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to update lecture status to pending")
		return nil, fmt.Errorf("failed to update lecture status: %w", err)
	}

	// 3. Publish ingestion job
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.lectureLogger.Warn().Err(err).Str("user_id", userID).Msg("Could not fetch user details for ingestion job enrichment")
	}
	name, email := "", ""
	if user != nil {
		name, email = user.Name, user.Email
	}

	payload := struct {
		LectureID          string `json:"lecture_id"`
		StoragePath        string `json:"storage_path"`
		CustomerIdentifier string `json:"customer_identifier"`
		Name               string `json:"name"`
		Email              string `json:"email"`
	}{
		LectureID:          lectureID,
		StoragePath:        lecture.StoragePath,
		CustomerIdentifier: userID,
		Name:               name,
		Email:              email,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to marshal ingestion payload")
		// Don't return error, but log it. The lecture is uploaded, but processing needs manual trigger.
	} else {
		if _, err := s.publisher.Publish(ctx, s.ingestionTopic, data); err != nil {
			s.lectureLogger.Error().Err(err).Str("topic", s.ingestionTopic).Msg("Failed to publish ingestion job")
			// Don't return an error here either.
		}
	}

	return lecture, nil
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

// GetPresignedURL generates a signed URL for the given storage path
func (s *lectureService) GetPresignedURL(ctx context.Context, storagePath string) (string, error) {
	resp, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(storagePath),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("storage_path", storagePath).Msg("Failed to generate presigned URL")
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return resp.URL, nil
}

// getPresignedPutURL generates a presigned URL for uploading an object.
func (s *lectureService) getPresignedPutURL(ctx context.Context, objectKey string) (string, error) {
	request, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		s.lectureLogger.Error().Err(err).Str("object_key", objectKey).Msg("Failed to generate presigned PUT URL")
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return request.URL, nil
}

// CountLecturesByUserInTimeRange delegates to repository
func (s *lectureService) CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error) {
	return s.usageRepo.CountUploadEventsInTimeRange(ctx, userID, start, end)
}

// RecordUploadEvent delegates to repository
func (s *lectureService) RecordUploadEvent(ctx context.Context, userID string, lectureID string) error {
	return s.usageRepo.RecordUploadEvent(ctx, userID, lectureID)
}
