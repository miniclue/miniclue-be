package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// NoteService defines note-related operations
type NoteService interface {
	// GetNotesByLectureID retrieves notes for a given lecture with pagination
	GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error)
	// UpdateNoteByLectureID updates a note's content by lecture and returns the updated note
	UpdateNoteByLectureID(ctx context.Context, lectureID string, content string) (*model.Note, error)
	// CreateNoteByLectureID creates a note for the given lecture and returns the created note
	CreateNoteByLectureID(ctx context.Context, userID string, lectureID string, content string) (*model.Note, error)
	// GetNoteByLectureIDAndUserID retrieves a note for a given lecture and user
	GetNoteByLectureIDAndUserID(ctx context.Context, lectureID string, userID string) (*model.Note, error)
}

// noteService is the implementation of NoteService
type noteService struct {
	repo repository.NoteRepository
}

// NewNoteService creates a new NoteService
func NewNoteService(repo repository.NoteRepository) NoteService {
	return &noteService{repo: repo}
}

// GetNotesByLectureID retrieves notes for a given lecture with pagination
func (s *noteService) GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error) {
	return s.repo.GetNotesByLectureID(ctx, lectureID, limit, offset)
}

// UpdateNoteByLectureID updates a note's content by lecture and returns updated note
func (s *noteService) UpdateNoteByLectureID(ctx context.Context, lectureID string, content string) (*model.Note, error) {
	return s.repo.UpdateNoteByLectureID(ctx, lectureID, content)
}

// CreateNoteByLectureID creates a note for the given lecture and returns the created note
func (s *noteService) CreateNoteByLectureID(ctx context.Context, userID string, lectureID string, content string) (*model.Note, error) {
	return s.repo.CreateNoteByLectureID(ctx, userID, lectureID, content)
}

// GetNoteByLectureIDAndUserID retrieves a note for a given lecture and user
func (s *noteService) GetNoteByLectureIDAndUserID(ctx context.Context, lectureID string, userID string) (*model.Note, error) {
	return s.repo.GetNoteByLectureIDAndUserID(ctx, lectureID, userID)
}
