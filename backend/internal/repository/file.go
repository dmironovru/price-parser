package repository

import (
        "context"

        "github.com/google/uuid"
        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/mds/ai-parser/internal/model"
)

type FileRepository struct {
        db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
        return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
        query := `
                INSERT INTO files (id, tenant_id, user_id, filename, original_name, file_hash, file_size_bytes, mime_type, status, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        `
        _, err := r.db.Exec(ctx, query,
                file.ID, file.TenantID, file.UserID,
                file.Filename, file.OriginalName,
                file.FileHash, file.FileSizeBytes, file.MimeType,
                file.Status,
                file.CreatedAt, file.UpdatedAt,
        )
        return err
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
        query := `
                SELECT id, tenant_id, user_id, filename, original_name, file_hash, file_size_bytes, mime_type, status, error_message, processing_started_at, processing_completed_at, created_at, updated_at
                FROM files WHERE id = $1
        `
        var f model.File
        err := r.db.QueryRow(ctx, query, id).Scan(
                &f.ID, &f.TenantID, &f.UserID,
                &f.Filename, &f.OriginalName,
                &f.FileHash, &f.FileSizeBytes, &f.MimeType,
                &f.Status,
                &f.ErrorMessage, &f.ProcessingStartedAt, &f.ProcessingCompletedAt,
                &f.CreatedAt, &f.UpdatedAt,
        )
        if err != nil {
                return nil, err
        }
        return &f, nil
}

func (r *FileRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus, errMsg *string) error {
        query := `UPDATE files SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`
        _, err := r.db.Exec(ctx, query, status, errMsg, id)
        return err
}
