.PHONY: help run build test test-verbose test-cover lint fmt vet \
        docker-up docker-down docker-logs docker-reset \
        uploads-dir tidy

# Detecta o shell para compatibilidade
SHELL := /bin/bash

# Variáveis
BINARY     := wedding-mc
CMD        := ./cmd/api/main.go
COVER_OUT  := coverage.out
COVER_MIN  := 80

# Cor para o help
CYAN  := \033[36m
RESET := \033[0m

# ─── Help ────────────────────────────────────────────────────────────────────

help: ## Mostra este menu de ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'

# ─── Desenvolvimento ─────────────────────────────────────────────────────────

run: uploads-dir ## Sobe a API localmente (aplica migrations automaticamente)
	go run $(CMD)

build: ## Compila o binário em ./bin/$(BINARY)
	@mkdir -p bin
	go build -o bin/$(BINARY) $(CMD)
	@echo "Binário gerado em bin/$(BINARY)"

uploads-dir: ## Cria a pasta de uploads local (se não existir)
	@mkdir -p uploads

tidy: ## Atualiza go.mod e go.sum
	go mod tidy

fmt: ## Formata todo o código Go
	go fmt ./...

vet: ## Executa go vet em todos os pacotes
	go vet ./...

# ─── Testes ──────────────────────────────────────────────────────────────────

test: ## Roda todos os testes
	go test ./...

test-verbose: ## Roda todos os testes com output detalhado
	go test -v ./...

test-cover: ## Roda os testes e exibe a cobertura por pacote
	go test ./... -cover

test-cover-report: ## Gera relatório de cobertura e abre no browser
	go test ./... -coverprofile=$(COVER_OUT)
	go tool cover -html=$(COVER_OUT)

test-cover-func: ## Exibe cobertura detalhada por função
	go test ./... -coverprofile=$(COVER_OUT)
	go tool cover -func=$(COVER_OUT)

test-check: ## Falha se a cobertura de qualquer pacote estiver abaixo de $(COVER_MIN)%
	@echo "Verificando cobertura mínima de $(COVER_MIN)%..."
	@go test ./... -coverprofile=$(COVER_OUT) > /dev/null
	@go tool cover -func=$(COVER_OUT) | awk \
		'/^total:/ { \
			pct = $$3+0; \
			if (pct < $(COVER_MIN)) { \
				printf "FALHOU: cobertura total %.1f%% < $(COVER_MIN)%%\n", pct; exit 1 \
			} else { \
				printf "OK: cobertura total %.1f%%\n", pct \
			} \
		}'

# ─── Docker / Banco ──────────────────────────────────────────────────────────

docker-up: ## Sobe o Postgres em background
	docker compose up -d
	@echo "Aguardando Postgres ficar saudável..."
	@until docker compose exec postgres pg_isready -U postgres -d weddingmc > /dev/null 2>&1; do sleep 1; done
	@echo "Postgres pronto."

docker-down: ## Para e remove os containers (preserva os dados)
	docker compose down

docker-logs: ## Exibe os logs do Postgres em tempo real
	docker compose logs -f postgres

docker-reset: ## Remove containers E volume (apaga todos os dados)
	@echo "ATENÇÃO: isso apagará todos os dados do banco local."
	@read -p "Confirmar? [s/N] " ans && [ "$$ans" = "s" ] || exit 0
	docker compose down -v
	@echo "Banco resetado."
