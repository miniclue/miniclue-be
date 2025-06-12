package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"
)

// NoteRepository defines note-related DB operations
type NoteRepository interface {
	// GetNotesByLectureID retrieves note records for a given lecture with pagination
	GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error)
	// UpdateNoteByLectureID updates a note's content for the given lecture and returns the updated note
	UpdateNoteByLectureID(ctx context.Context, lectureID string, content string) (*model.Note, error)
	// CreateNoteByLectureID creates a note for the given lecture and returns the created note
	CreateNoteByLectureID(ctx context.Context, userID string, lectureID string, content string) (*model.Note, error)
	// GetNoteByLectureIDAndUserID retrieves a note for a given lecture and user
	GetNoteByLectureIDAndUserID(ctx context.Context, lectureID string, userID string) (*model.Note, error)
}

// noteRepository is the DB implementation of NoteRepository
type noteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new NoteRepository
func NewNoteRepository(db *sql.DB) NoteRepository {
	return &noteRepository{db: db}
}

// GetNotesByLectureID retrieves note records for a given lecture with pagination
func (r *noteRepository) GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error) {
	query := `SELECT id, lecture_id, content, created_at, updated_at FROM notes WHERE lecture_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, lectureID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	notes := []model.Note{}
	for rows.Next() {
		var n model.Note
		if err := rows.Scan(&n.ID, &n.LectureID, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return notes, nil
}

// UpdateNoteByLectureID updates a note's content and returns updated record
func (r *noteRepository) UpdateNoteByLectureID(ctx context.Context, lectureID string, content string) (*model.Note, error) {
	query := `UPDATE notes SET content = $1, updated_at = NOW() WHERE lecture_id = $2 RETURNING id, user_id, lecture_id, content, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query, content, lectureID)
	var n model.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.LectureID, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}
	return &n, nil
}

// CreateNoteByLectureID creates a note record for the given lecture and returns the created note
func (r *noteRepository) CreateNoteByLectureID(ctx context.Context, userID string, lectureID string, content string) (*model.Note, error) {
	query := `INSERT INTO notes (user_id, lecture_id, content) VALUES ($1, $2, $3) RETURNING id, user_id, lecture_id, content, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query, userID, lectureID, content)
	var n model.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.LectureID, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}
	return &n, nil
}

// GetNoteByLectureIDAndUserID retrieves a note for a given lecture and user
func (r *noteRepository) GetNoteByLectureIDAndUserID(ctx context.Context, lectureID string, userID string) (*model.Note, error) {
	query := `SELECT id, user_id, lecture_id, content, created_at, updated_at FROM notes WHERE lecture_id = $1 AND user_id = $2`
	row := r.db.QueryRowContext(ctx, query, lectureID, userID)
	var n model.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.LectureID, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query existing note: %w", err)
	}
	return &n, nil
}
