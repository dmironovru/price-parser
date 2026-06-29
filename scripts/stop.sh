#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${YELLOW}🛑 Остановка сервисов...${NC}"

pkill -f "go run cmd/server/main.go" 2>/dev/null && echo -e "${GREEN}✅ Бэкенд (dev) остановлен${NC}" || true
pkill -f "next dev" 2>/dev/null && echo -e "${GREEN}✅ Фронтенд (dev) остановлен${NC}" || true
pkill -f "backend/bin/price-parser" 2>/dev/null && echo -e "${GREEN}✅ Бэкенд (prod) остановлен${NC}" || true
pkill -f "next start" 2>/dev/null && echo -e "${GREEN}✅ Фронтенд (prod) остановлен${NC}" || true

echo -e "${GREEN}✅ Все сервисы остановлены${NC}"
