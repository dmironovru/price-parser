#!/bin/bash

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${GREEN}📦 Сборка проекта...${NC}"

# Сборка бэкенда
echo -e "${YELLOW}📦 Сборка бэкенда...${NC}"
cd backend
go build -o bin/price-parser cmd/server/main.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Бэкенд собран${NC}"
else
    echo -e "${RED}❌ Ошибка сборки бэкенда${NC}"
    exit 1
fi
cd ..

# Сборка фронтенда
echo -e "${YELLOW}📦 Сборка фронтенда...${NC}"
cd frontend
npm run build
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Фронтенд собран${NC}"
else
    echo -e "${RED}❌ Ошибка сборки фронтенда${NC}"
    exit 1
fi
cd ..

echo -e "${GREEN}✅ Сборка завершена!${NC}"
