package model

import (
        "time"
        "github.com/google/uuid"
)

type Product struct {
        ID              uuid.UUID  `json:"id"`
        TenantID        uuid.UUID  `json:"tenant_id"`
        FileID          uuid.UUID  `json:"file_id"`
        RowNumber       *int       `json:"row_number,omitempty"`
        SKU             *string    `json:"sku,omitempty"`
        Name            string     `json:"name"`
        Price           *float64   `json:"price,omitempty"`
        Currency        string     `json:"currency"`
        Unit            *string    `json:"unit,omitempty"`
        Description     *string    `json:"description,omitempty"`
        Category        *string    `json:"category,omitempty"`
        RawData         map[string]any `json:"raw_data,omitempty"`
        ConfidenceScore *float64   `json:"confidence_score,omitempty"`
        CreatedAt       time.Time  `json:"created_at"`
        UpdatedAt       time.Time  `json:"updated_at"`
}
