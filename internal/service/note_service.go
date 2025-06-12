package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// NoteService defines note-related operations
// GetNotesByLectureID retrieves notes for a given lecture with pagination
// returns empty slice if none exist
 type NoteService interface {
	GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error)
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
