package queue

import (
        "encoding/json"

        "github.com/google/uuid"
        "github.com/hibiken/asynq"
)

const TypeFileProcessing = "file:processing"

type FileProcessingPayload struct {
        FileID   uuid.UUID `json:"file_id"`
        TenantID uuid.UUID `json:"tenant_id"`
        UserID   uuid.UUID `json:"user_id"`
        FilePath string    `json:"file_path"`
        MimeType string    `json:"mime_type"`
}

func NewFileProcessingTask(payload FileProcessingPayload) (*asynq.Task, error) {
        data, err := json.Marshal(payload)
        if err != nil {
                return nil, err
        }
        return asynq.NewTask(TypeFileProcessing, data), nil
}
