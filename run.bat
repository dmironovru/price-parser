@echo off
chcp 65001 >nul
title AI Парсер Прайс-листов

echo 🚀 Запуск AI Парсера Прайс-листов
echo.

:: Проверяем Go
where go >nul 2>nul
if errorlevel 1 (
    echo ❌ Go не установлен!
    echo Скачай: https://go.dev/dl/
    pause
    exit /b
)

:: Проверяем Node.js
where node >nul 2>nul
if errorlevel 1 (
    echo ❌ Node.js не установлен!
    echo Скачай: https://nodejs.org/
    pause
    exit /b
)

:: Проверяем Docker
where docker >nul 2>nul
if errorlevel 1 (
    echo ❌ Docker не установлен!
    echo Скачай: https://www.docker.com/products/docker-desktop/
    pause
    exit /b
)

echo ✅ Go, Node.js, Docker найдены
echo.

:: Поднимаем контейнеры
echo 📦 Поднимаем контейнеры...
docker-compose up -d

timeout /t 3 /nobreak >nul

:: Запускаем бэкенд
echo 🔧 Запуск бэкенда...
start /B cmd /c "cd backend && go run cmd/api/main.go"

timeout /t 3 /nobreak >nul

:: Запускаем фронтенд
echo 🎨 Запуск фронтенда...
start /B cmd /c "cd frontend && npm run dev"

timeout /t 5 /nobreak >nul

:: Открываем браузер
echo 🌐 Открываем браузер...
start http://localhost:3001

echo.
echo ✅ Все сервисы запущены!
echo 📌 Открыто: http://localhost:3001
echo.
echo ⚠️  Закройте это окно, чтобы остановить все сервисы
echo.

pause >nul

:: Остановка
echo 🛑 Остановка сервисов...
taskkill /F /IM go.exe 2>nul
taskkill /F /IM node.exe 2>nul
docker-compose down
echo ✅ Готово!
