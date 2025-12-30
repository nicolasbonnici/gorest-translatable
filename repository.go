package translatable

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, t *Translatable) error
	GetByID(ctx context.Context, id uuid.UUID) (*Translatable, error)
	Query(ctx context.Context, params QueryParams) ([]*Translatable, error)
	Update(ctx context.Context, id uuid.UUID, content string, userID *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, t *Translatable) error {
	query := `
		INSERT INTO translatable (id, user_id, translatable_id, translatable, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		t.ID,
		t.UserID,
		t.TranslatableID,
		t.Translatable,
		t.Content,
		t.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create translatable: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Translatable, error) {
	query := `
		SELECT id, user_id, translatable_id, translatable, content, updated_at, created_at
		FROM translatable
		WHERE id = $1
	`

	var t Translatable
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.UserID,
		&t.TranslatableID,
		&t.Translatable,
		&t.Content,
		&t.UpdatedAt,
		&t.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("translatable not found")
		}
		return nil, fmt.Errorf("failed to get translatable: %w", err)
	}

	return &t, nil
}

func (r *repository) Query(ctx context.Context, params QueryParams) ([]*Translatable, error) {
	query := `
		SELECT id, user_id, translatable_id, translatable, content, updated_at, created_at
		FROM translatable
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if params.TranslatableID != nil {
		query += fmt.Sprintf(" AND translatable_id = $%d", argCount)
		args = append(args, *params.TranslatableID)
		argCount++
	}

	if params.Translatable != nil {
		query += fmt.Sprintf(" AND translatable = $%d", argCount)
		args = append(args, *params.Translatable)
		argCount++
	}

	if params.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *params.UserID)
		argCount++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query translatable: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []*Translatable
	for rows.Next() {
		var t Translatable
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.TranslatableID,
			&t.Translatable,
			&t.Content,
			&t.UpdatedAt,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan translatable: %w", err)
		}
		results = append(results, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating translatable rows: %w", err)
	}

	return results, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, content string, userID *uuid.UUID) error {
	query := `
		UPDATE translatable
		SET content = $1, updated_at = $2
		WHERE id = $3
	`

	args := []interface{}{content, time.Now(), id}
	if userID != nil {
		query += " AND user_id = $4"
		args = append(args, *userID)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update translatable: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		if userID != nil {
			return errors.New("translatable not found or you don't have permission to update it")
		}
		return errors.New("translatable not found")
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	query := `DELETE FROM translatable WHERE id = $1`

	args := []interface{}{id}
	if userID != nil {
		query += " AND user_id = $2"
		args = append(args, *userID)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete translatable: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		if userID != nil {
			return errors.New("translatable not found or you don't have permission to delete it")
		}
		return errors.New("translatable not found")
	}

	return nil
}
