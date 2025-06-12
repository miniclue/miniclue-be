package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"
)

// NoteRepository defines note-related DB operations
// GetNotesByLectureID retrieves note records for a given lecture with pagination
// returns empty slice if none exist
 type NoteRepository interface {
	GetNotesByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Note, error)
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
