# wedding-mc

API REST para gestão de casamentos. Oferece ao casal uma landing page personalizada e ferramentas para gerenciar convidados e lista de presentes, com acesso simplificado para convidados sem necessidade de autenticação.

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.26 |
| Framework HTTP | Chi |
| Banco de dados | PostgreSQL 16 |
| Queries | sqlx + golang-migrate |
| Autenticação | JWT (golang-jwt/jwt) |
| Validação | go-playground/validator |
| Storage de fotos | Local (dev) / AWS S3 (prod) |
| Logging | zerolog |

## Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) e Docker Compose

## Setup local

**1. Clone o repositório e instale as dependências:**

```bash
git clone https://github.com/ropehapi/wedding-mc.git
cd wedding-mc
go mod download
```

**2. Configure as variáveis de ambiente:**

```bash
cp .env.example .env
```

O `.env.example` já vem com valores funcionais para desenvolvimento local. Ajuste se necessário:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/weddingmc?sslmode=disable
JWT_SECRET=change-me-in-production
JWT_EXPIRY=1h
REFRESH_EXPIRY=168h
STORAGE_DRIVER=local
LOCAL_STORAGE_PATH=./uploads
PORT=8080
ALLOWED_ORIGINS=*
```

**3. Suba o banco de dados:**

```bash
docker compose up -d
```

Aguarde o healthcheck do Postgres passar (alguns segundos). Para verificar:

```bash
docker compose ps
```

**4. Rode a API:**

```bash
go run ./cmd/api/main.go
```

As migrations são aplicadas automaticamente na inicialização. A API sobe na porta `8080`.

```
INF server starting addr=:8080
```

## Executar os testes

**Todos os testes:**

```bash
go test ./...
```

**Com cobertura:**

```bash
go test ./... -cover
```

**Pacote específico:**

```bash
go test ./internal/service/... -v
go test ./internal/handler/... -v
```

**Cobertura detalhada por função:**

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

A cobertura mínima esperada é de **80%** por pacote.

## Testar manualmente os endpoints

Os arquivos `.http` na pasta `http/` permitem executar requisições diretamente do editor.

**VS Code** — instale a extensão [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) e clique em `Send Request` acima de cada bloco.

**IntelliJ / GoLand** — suporte nativo. Selecione o ambiente `dev` no seletor da IDE.

### Fluxo básico de teste

**Passo 1 — Registrar e autenticar (`http/auth.http`):**

```
POST /v1/auth/register   → cria a conta
POST /v1/auth/login      → retorna access_token e refresh_token
```

Copie o `access_token` do response do login e cole na variável `@token` no topo dos arquivos `.http`.

**Passo 2 — Criar e gerenciar o casamento (`http/wedding.http`):**

```
POST   /v1/wedding                    → cria o casamento (slug gerado automaticamente)
GET    /v1/wedding                    → consulta o perfil completo
PATCH  /v1/wedding                    → atualiza campos (partial update)
POST   /v1/wedding/photos             → upload de foto (multipart/form-data)
DELETE /v1/wedding/photos/{photoID}   → remove uma foto
```

### Upload de foto via curl

O VS Code REST Client requer um arquivo local na pasta `http/` para o upload. Alternativamente, use curl:

```bash
curl -X POST http://localhost:8080/v1/wedding/photos \
  -H "Authorization: Bearer SEU_TOKEN" \
  -F "photo=@/caminho/para/foto.jpg"
```

Formatos aceitos: `.jpg`, `.jpeg`, `.png`, `.webp` — máximo de **10MB**.

As fotos são servidas estaticamente em:

```
http://localhost:8080/uploads/weddings/{weddingID}/{fotoID}.jpg
```

### Endpoints disponíveis

| Método | Endpoint | Auth | Descrição |
|---|---|---|---|
| `POST` | `/v1/auth/register` | — | Registrar casal |
| `POST` | `/v1/auth/login` | — | Login |
| `POST` | `/v1/auth/refresh` | — | Renovar access token |
| `POST` | `/v1/auth/logout` | JWT | Logout |
| `GET` | `/v1/wedding` | JWT | Consultar casamento |
| `POST` | `/v1/wedding` | JWT | Criar casamento |
| `PATCH` | `/v1/wedding` | JWT | Editar casamento |
| `POST` | `/v1/wedding/photos` | JWT | Upload de foto |
| `DELETE` | `/v1/wedding/photos/{photoID}` | JWT | Remover foto |

### Formato das respostas

Sucesso:
```json
{
  "data": { ... }
}
```

Erro:
```json
{
  "error": "not_found",
  "message": "wedding not found"
}
```

Erro de validação:
```json
{
  "error": "validation_error",
  "details": [
    { "field": "BrideName", "rule": "required" }
  ]
}
```

## Estrutura do projeto

```
cmd/
  api/
    main.go              ← entrypoint, DI manual, roteamento
internal/
  config/                ← leitura de env vars, conexão com banco
  domain/                ← structs, interfaces, erros de domínio
  handler/               ← recebe HTTP, valida, chama service, responde
  middleware/            ← auth JWT, logger, recover, CORS
  repository/            ← queries SQL via sqlx
  service/               ← regras de negócio
migrations/              ← arquivos .up.sql / .down.sql (golang-migrate)
http/                    ← arquivos .http para testes manuais
```

## Módulos MVP

| Módulo | Status |
|---|---|
| Auth | ✅ Implementado |
| Casamento | ✅ Implementado |
| Convidados | 🔲 Pendente |
| Presentes | 🔲 Pendente |
| Página Pública | 🔲 Pendente |
