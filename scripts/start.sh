#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Запуск продакшен версии...${NC}"

# Проверка сборки
if [ ! -f "backend/bin/price-parser" ]; then
    echo -e "${YELLOW}⚠️ Бэкенд не собран, запускаю сборку...${NC}"
    ./scripts/build.sh
fi

# Запуск бэкенда
echo -e "${YELLOW}📦 Запуск бэкенда...${NC}"
cd backend
./bin/price-parser > /tmp/backend.log 2>&1 &
BACKEND_PID=$!
sleep 2
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
PORT=3000 npm start > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
sleep 3
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Фронтенд запущен (PID: $FRONTEND_PID)${NC}"
else
    echo -e "${RED}❌ Ошибка запуска фронтенда${NC}"
    tail -10 /tmp/frontend.log
    exit 1
fi
cd ..

echo -e "${GREEN}✅ Все сервисы запущены!${NC}"
