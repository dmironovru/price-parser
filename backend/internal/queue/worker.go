package queue

import (
        "context"
        "encoding/json"
        "log"
        "time"

        "github.com/google/uuid"
        "github.com/hibiken/asynq"
        "github.com/jackc/pgx/v5/pgxpool"
        "github.com/mds/ai-parser/internal/model"
        "github.com/mds/ai-parser/internal/parser"
        "github.com/mds/ai-parser/internal/repository"
)

type Worker struct {
        server      *asynq.Server
        productRepo *repository.ProductRepository
        fileRepo    *repository.FileRepository
}

func NewWorker(redisAddr string, pool *pgxpool.Pool) *Worker {
        srv := asynq.NewServer(
                asynq.RedisClientOpt{Addr: redisAddr},
                asynq.Config{
                        Concurrency: 5,
                        Queues: map[string]int{
                                "default": 1,
                        },
                },
        )
        return &Worker{
                server:      srv,
                productRepo: repository.NewProductRepository(pool),
                fileRepo:    repository.NewFileRepository(pool),
        }
}

func (w *Worker) Start() error {
        mux := asynq.NewServeMux()
        mux.HandleFunc(TypeFileProcessing, w.handleFileProcessing)
        return w.server.Start(mux)
}

func (w *Worker) Stop() {
        w.server.Stop()
}

func (w *Worker) handleFileProcessing(ctx context.Context, t *asynq.Task) error {
        var payload FileProcessingPayload
        if err := json.Unmarshal(t.Payload(), &payload); err != nil {
                return err
        }

        log.Printf("Processing file: %s (ID: %s)", payload.FilePath, payload.FileID)

        result, err := parser.ParseFile(payload.FilePath)
        if err != nil {
                log.Printf("Parse error: %v", err)
                return err
        }

        allProducts := result.Products
        log.Printf("📦 Найдено товаров: %d", len(allProducts))

        if len(allProducts) > 0 {
                var dbProducts []model.Product
                for _, p := range allProducts {
                        price := p.Price
                        sku := p.SKU
                        unit := p.Unit
                        dbProducts = append(dbProducts, model.Product{
                                ID:        uuid.New(),
                                TenantID:  payload.TenantID,
                                FileID:    payload.FileID,
                                SKU:       &sku,
                                Name:      p.Name,
                                Price:     &price,
                                Currency:  "RUB",
                                Unit:      &unit,
                                CreatedAt: time.Now(),
                                UpdatedAt: time.Now(),
                        })
                }

                if err := w.productRepo.CreateBatch(ctx, dbProducts); err != nil {
                        log.Printf("❌ Failed to save products: %v", err)
                        return err
                }

                log.Printf("✅ Saved %d products to database", len(dbProducts))

                if err := w.fileRepo.UpdateStatus(ctx, payload.FileID, model.FileStatusCompleted, nil); err != nil {
                        log.Printf("Failed to update file status: %v", err)
                }
        } else {
                errMsg := "No products found"
                log.Printf("⚠️ %s", errMsg)
                if err := w.fileRepo.UpdateStatus(ctx, payload.FileID, model.FileStatusFailed, &errMsg); err != nil {
                        log.Printf("Failed to update file status: %v", err)
                }
        }

        log.Printf("File %s processed!", payload.FileID)
        return nil
}
