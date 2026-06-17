package repository

import (
        "context"
        "fmt"
        "time"

        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/mds/ai-parser/internal/model"
)

type ProductRepository struct {
        db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
        return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateBatch(ctx context.Context, products []model.Product) error {
        if len(products) == 0 {
                return nil
        }

        query := `
                INSERT INTO parsed_products (id, tenant_id, file_id, row_number, sku, name, price, currency, unit, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        `

        for i, p := range products {
                rowNum := i + 1
                _, err := r.db.Exec(ctx, query,
                        p.ID, p.TenantID, p.FileID, rowNum,
                        p.SKU, p.Name, p.Price, p.Currency, p.Unit,
                        time.Now(), time.Now(),
                )
                if err != nil {
                        return fmt.Errorf("insert product %d: %w", i, err)
                }
        }

        return nil
}
