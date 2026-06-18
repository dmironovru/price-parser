#!/bin/bash

echo "🚀 Starting AI Price Parser..."
cd ~/ai-parser

# Запускаем Docker
echo "📦 Starting Docker..."
docker-compose up -d --remove-orphans
sleep 3

# Запускаем бэкенд в фоне
echo "🔧 Starting Backend..."
cd backend
export $(grep -v '^#' ../.env | xargs)
nohup go run cmd/api/main.go > /tmp/ai-parser-backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > /tmp/ai-parser-backend.pid
cd ..

# Запускаем фронтенд в фоне
echo "🎨 Starting Frontend..."
cd frontend
nohup npm run dev -- --webpack -p 3001 > /tmp/ai-parser-frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > /tmp/ai-parser-frontend.pid
cd ..

sleep 5

# Открываем браузер
echo "🌐 Opening browser..."
xdg-open http://localhost:3001 2>/dev/null || sensible-browser http://localhost:3001 2>/dev/null || echo "Open http://localhost:3001 in your browser"

echo ""
echo "✅ All services started!"
echo "📌 Backend:  http://localhost:3000  (log: /tmp/ai-parser-backend.log)"
echo "📌 Frontend: http://localhost:3001  (log: /tmp/ai-parser-frontend.log)"
echo ""
echo "🛑 To stop, run: ./stop.sh"
