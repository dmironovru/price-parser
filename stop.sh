#!/bin/bash

echo "🛑 Stopping AI Price Parser..."

# Останавливаем бэкенд
if [ -f /tmp/ai-parser-backend.pid ]; then
    BACKEND_PID=$(cat /tmp/ai-parser-backend.pid)
    echo "🔧 Stopping Backend (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null
    rm /tmp/ai-parser-backend.pid
fi

# Останавливаем фронтенд
if [ -f /tmp/ai-parser-frontend.pid ]; then
    FRONTEND_PID=$(cat /tmp/ai-parser-frontend.pid)
    echo "🎨 Stopping Frontend (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null
    rm /tmp/ai-parser-frontend.pid
fi

# Добиваем процессы на всякий случай
echo "🧹 Cleaning up remaining processes..."
pkill -f "go run cmd/api/main.go" 2>/dev/null
pkill -f "next dev" 2>/dev/null

# Останавливаем Docker
echo "📦 Stopping Docker containers..."
cd ~/ai-parser
docker-compose down

# Освобождаем порты
echo "🔓 Freeing ports..."
fuser -k 3000/tcp 2>/dev/null
fuser -k 3001/tcp 2>/dev/null

echo "✅ All services stopped!"
