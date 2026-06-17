# AI Parser — B2B Micro-SaaS

**AI-парсер прайс-листов** с локальной LLM (Ollama) и векторным поиском.

## 🚀 Технологии
- **Backend:** Go + Fiber
- **База данных:** PostgreSQL + pgvector
- **Очередь:** Redis + asynq
- **AI:** Ollama (llama3.2:3b)
- **Деплой:** Docker Compose

## 📦 Функционал
- Загрузка файлов (CSV, TXT, Excel, PDF)
- Парсинг через AI (локально, данные не покидают сервер)
- Сохранение товаров в БД
- Асинхронная обработка через очередь

## ▶️ Запуск
```bash
docker compose up -d
cd backend
export $(grep -v '^#' ../.env | xargs)
go run cmd/api/main.go
📊 API
POST /upload — загрузка файла

GET /health — проверка статуса

📈 Результат
Файл → AI парсинг → сохранение в БД

Векторный поиск похожих товаров (pgvector)
