-- Включение pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Основная таблица задач
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    metadata JSONB,
    processed_rows INTEGER DEFAULT 0,
    total_rows INTEGER DEFAULT 0
);

-- Таблица извлеченных данных с векторными эмбеддингами
CREATE TABLE IF NOT EXISTS extracted_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    row_index INTEGER NOT NULL,
    product_name VARCHAR(500) NOT NULL,
    product_code VARCHAR(100),
    price DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'RUB',
    quantity INTEGER,
    unit VARCHAR(50),
    description TEXT,
    confidence_score FLOAT,
    embedding vector(768),
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для поиска
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_extracted_task_id ON extracted_data(task_id);
CREATE INDEX idx_extracted_embedding ON extracted_data USING ivfflat (embedding vector_cosine_ops);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tasks_updated_at 
    BEFORE UPDATE ON tasks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_extracted_updated_at 
    BEFORE UPDATE ON extracted_data 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();