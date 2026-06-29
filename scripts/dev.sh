#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     AI Price Parser — Режим разработки         ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# Проверка Ollama
echo -e "${YELLOW}📦 Проверка Ollama...${NC}"
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Ollama уже запущен${NC}"
else
    echo -e "${YELLOW}⏳ Запуск Ollama...${NC}"
    ollama serve > /dev/null 2>&1 &
    sleep 3
    echo -e "${GREEN}✅ Ollama запущен${NC}"
fi

# Остановка старых процессов
echo -e "${YELLOW}📦 Остановка старых процессов...${NC}"
pkill -f "go run cmd/server/main.go" 2>/dev/null || true
pkill -f "next dev" 2>/dev/null || true
sleep 1

# Запуск бэкенда
echo -e "${YELLOW}📦 Запуск бэкенда...${NC}"
cd backend
go run cmd/server/main.go > /tmp/backend.log 2>&1 &
BACKEND_PID=$!
sleep 3
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Бэкенд запущен (PID: $BACKEND_PID)${NC}"
else
    echo -e "${RED}❌ Ошибка запуска бэкенда${NC}"
    tail -10 /tmp/backend.log
    exit 1
fi
cd ..

# Запуск фронтенда
echo -e "${YELLOW}📦 Запуск фронтенда...${NC}"
cd frontend
npm run dev > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
sleep 5
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Фронтенд запущен (PID: $FRONTEND_PID)${NC}"
else
    echo -e "${RED}❌ Ошибка запуска фронтенда${NC}"
    tail -10 /tmp/frontend.log
    exit 1
fi
cd ..

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   🎉 ВСЕ СЕРВИСЫ ЗАПУЩЕНЫ!                     ║${NC}"
echo -e "${GREEN}║   🌐 Фронтенд: http://localhost:3000           ║${NC}"
echo -e "${GREEN}║   🔧 Бэкенд:  http://localhost:8080            ║${NC}"
echo -e "${GREEN}║   🤖 Ollama:  http://localhost:11434           ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}📝 Логи: tail -f /tmp/backend.log${NC}"
echo -e "${YELLOW}🛑 Остановка: make stop${NC}"
