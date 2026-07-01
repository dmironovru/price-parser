#!/bin/bash

echo "🚀 Быстрый старт AI Price Parser"
echo "================================="

# 1. Проверка и установка зависимостей
echo "📦 Проверка зависимостей..."

# Проверка zstd
if ! command -v zstd &> /dev/null; then
    echo "📦 Установка zstd..."
    sudo apt update && sudo apt install zstd -y
fi

# Проверка Ollama
if ! command -v ollama &> /dev/null; then
    echo "📦 Установка Ollama..."
    curl -fsSL https://ollama.com/install.sh | sh
fi

# Проверка модели
if ! ollama list | grep -q "llama3.2:3b"; then
    echo "📦 Загрузка модели llama3.2:3b (~2GB)..."
    ollama pull llama3.2:3b
fi

# 2. Запуск Ollama (если не запущен)
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "🚀 Запуск Ollama..."
    ollama serve > /dev/null 2>&1 &
    sleep 3
fi

# 3. Запуск проекта
echo "🚀 Запуск проекта..."
make dev
