cat > quickstart.sh << 'EOF'
#!/bin/bash

set -e

echo "🚀 Быстрый старт AI Price Parser"
echo "================================="
echo ""

# 1. Проверка zstd
echo "📦 Проверка zstd..."
if ! command -v zstd &> /dev/null; then
    echo "📦 Установка zstd..."
    sudo apt update && sudo apt install zstd -y
else
    echo "✅ zstd уже установлен"
fi

# 2. Проверка Ollama
echo "📦 Проверка Ollama..."
if ! command -v ollama &> /dev/null; then
    echo "📦 Установка Ollama..."
    curl -fsSL https://ollama.com/install.sh | sh
else
    echo "✅ Ollama уже установлен"
fi

# 3. Проверка модели
echo "📦 Проверка модели llama3.2:3b..."
if ! ollama list | grep -q "llama3.2:3b"; then
    echo "📦 Загрузка модели llama3.2:3b (~2GB)..."
    echo "⏳ Это может занять 5-10 минут..."
    ollama pull llama3.2:3b
else
    echo "✅ Модель уже загружена"
fi

# 4. Запуск Ollama
echo "📦 Проверка запуска Ollama..."
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "🚀 Запуск Ollama..."
    ollama serve > /dev/null 2>&1 &
    sleep 3
else
    echo "✅ Ollama уже запущен"
fi

# 5. Установка зависимостей проекта
echo "📦 Установка зависимостей проекта..."

echo "📦 Установка Go зависимостей..."
cd backend
go mod download
cd ..

echo "📦 Установка Node.js зависимостей..."
cd frontend
rm -rf node_modules package-lock.json  # Очистка для чистой установки
npm install
cd ..

# 6. Запуск проекта
echo "🚀 Запуск проекта..."
make dev

echo ""
echo "✅ Готово! Открой http://localhost:3000"
EOF

chmod +x quickstart.sh