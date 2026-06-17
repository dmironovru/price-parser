package model

import (
        "time"

        "github.com/google/uuid"
)

type FileStatus string

const (
        FileStatusPending    FileStatus = "pending"
        FileStatusProcessing FileStatus = "processing"
        FileStatusCompleted  FileStatus = "completed"
        FileStatusFailed     FileStatus = "failed"
)

type File struct {
        ID                    uuid.UUID  `json:"id"`
        TenantID              uuid.UUID  `json:"tenant_id"`
        UserID                uuid.UUID  `json:"user_id"`
        Filename              string     `json:"filename"`
        OriginalName          string     `json:"original_name"`
        FileHash              string     `json:"file_hash"`
        FileSizeBytes         int64      `json:"file_size_bytes"`
        MimeType              string     `json:"mime_type"`
        Status                FileStatus `json:"status"`
        ErrorMessage          *string    `json:"error_message,omitempty"`
        ProcessingStartedAt   *time.Time `json:"processing_started_at,omitempty"`
        ProcessingCompletedAt *time.Time `json:"processing_completed_at,omitempty"`
        CreatedAt             time.Time  `json:"created_at"`
        UpdatedAt             time.Time  `json:"updated_at"`
}
