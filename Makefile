.PHONY: help dev build start stop clean deploy

# Colors
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
BLUE   := \033[0;34m
NC     := \033[0m

help:
	@echo "$(GREEN)╔══════════════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║     AI Price Parser — Команды для управления    ║$(NC)"
	@echo "$(GREEN)╚══════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)📦 Основные команды:$(NC)"
	@echo "  $(BLUE)make dev$(NC)       — Запустить в режиме разработки"
	@echo "  $(BLUE)make build$(NC)      — Собрать продакшен версию"
	@echo "  $(BLUE)make start$(NC)      — Запустить собранную версию"
	@echo "  $(BLUE)make stop$(NC)       — Остановить все процессы"
	@echo "  $(BLUE)make clean$(NC)      — Очистить сборки и кэш"
	@echo "  $(BLUE)make deploy$(NC)     — Полный деплой с нуля"
	@echo "  $(BLUE)make install$(NC)    — Установить все зависимости"
	@echo ""

dev:
	@echo "$(GREEN)🚀 Запуск в режиме разработки...$(NC)"
	@./scripts/dev.sh || ./dev.sh

build:
	@echo "$(GREEN)📦 Сборка проекта...$(NC)"
	@./scripts/build.sh || ./build.sh

start:
	@echo "$(GREEN)🚀 Запуск продакшен версии...$(NC)"
	@./scripts/start.sh || ./start.sh

stop:
	@echo "$(YELLOW)🛑 Остановка сервисов...$(NC)"
	@./scripts/stop.sh || ./stop.sh

clean:
	@echo "$(YELLOW)🧹 Очистка...$(NC)"
	@rm -rf backend/bin/
	@rm -rf frontend/.next/
	@rm -rf frontend/out/
	@rm -f backend/price-parser
	@rm -rf /tmp/price-parser-*
	@echo "$(GREEN)✅ Очистка завершена$(NC)"

install:
	@echo "$(YELLOW)📦 Установка зависимостей...$(NC)"
	@cd backend && go mod download
	@cd frontend && npm install
	@echo "$(GREEN)✅ Зависимости установлены$(NC)"

deploy: install build
	@echo "$(GREEN)🚀 Полный деплой...$(NC)"
	@make stop
	@make start
	@echo "$(GREEN)✅ Деплой завершен!$(NC)"

.PHONY: help dev build start stop clean deploy install
