# Design de Arquitetura — wedding-mc

**Specs:** `.specs/features/*/spec.md`
**Status:** Draft

---

## Visão Geral

API REST em Go seguindo arquitetura em camadas (handler → service → repository). Cada camada tem responsabilidade única e se comunica apenas com a camada adjacente. Injeção de dependência manual no `main.go`.

```
HTTP Request
    │
    ▼
┌─────────────┐
│  Middleware  │  ← auth JWT, logger, recover, CORS
└──────┬──────┘
       │
    ▼
┌─────────────┐
│   Handler   │  ← parse request, valida input, chama service, serializa response
└──────┬──────┘
       │
    ▼
┌─────────────┐
│   Service   │  ← regras de negócio, orquestra repositórios e storages
└──────┬──────┘
       │
    ▼
┌─────────────┐
│ Repository  │  ← queries SQL via sqlx, sem lógica de negócio
└──────┬──────┘
       │
    ▼
 PostgreSQL
```

---

## Estrutura de Pastas

```
wedding-mc/
├── cmd/
│   └── api/
│       └── main.go              ← entrypoint: carrega config, conecta DB, monta rotas, DI manual
├── internal/
│   ├── config/
│   │   └── config.go            ← lê env vars via godotenv, expõe struct Config
│   ├── domain/
│   │   ├── user.go              ← struct User, interface UserRepository
│   │   ├── wedding.go           ← struct Wedding, Photo, Link, interface WeddingRepository
│   │   ├── guest.go             ← struct Guest, RSVPStatus enum, interface GuestRepository
│   │   ├── gift.go              ← struct Gift, GiftStatus enum, interface GiftRepository
│   │   └── errors.go            ← erros de domínio (ErrNotFound, ErrConflict, etc.)
│   ├── handler/
│   │   ├── auth.go              ← POST /v1/auth/register, /login, /refresh, /logout
│   │   ├── wedding.go           ← GET/POST/PATCH /v1/wedding + /photos
│   │   ├── guest.go             ← CRUD /v1/guests + /summary
│   │   ├── gift.go              ← CRUD /v1/gifts + /summary + /reserve
│   │   ├── public.go            ← GET /v1/public/:slug, /guests, /gifts + RSVP + reserva
│   │   └── response.go          ← helpers: JSON(), Error(), envelope padrão
│   ├── service/
│   │   ├── auth.go              ← registro, login, geração/validação de JWT
│   │   ├── wedding.go           ← criação/edição do perfil, geração de slug
│   │   ├── guest.go             ← CRUD convidados, lógica de RSVP
│   │   ├── gift.go              ← CRUD presentes, lógica de reserva (atomicidade)
│   │   └── storage.go           ← interface StorageService + implementações S3 e local
│   ├── repository/
│   │   ├── user.go              ← queries: FindByEmail, Create
│   │   ├── wedding.go           ← queries: FindByUserID, FindBySlug, Create, Update
│   │   ├── guest.go             ← queries: FindAll, FindByID, Create, Update, Delete
│   │   └── gift.go              ← queries: FindAll, FindByID, Create, Update, Delete, Reserve
│   └── middleware/
│       ├── auth.go              ← valida JWT, injeta user no contexto
│       ├── logger.go            ← log de cada request (zerolog)
│       ├── recover.go           ← captura panics, retorna 500
│       └── cors.go              ← headers CORS para o frontend externo
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_weddings.up.sql
│   ├── 000002_create_weddings.down.sql
│   ├── 000003_create_wedding_photos.up.sql
│   ├── 000003_create_wedding_photos.down.sql
│   ├── 000004_create_wedding_links.up.sql
│   ├── 000004_create_wedding_links.down.sql
│   ├── 000005_create_guests.up.sql
│   ├── 000005_create_guests.down.sql
│   ├── 000006_create_gifts.up.sql
│   ├── 000006_create_gifts.down.sql
│   ├── 000007_create_refresh_tokens.up.sql
│   └── 000007_create_refresh_tokens.down.sql
├── docs/
│   ├── swagger.json             ← gerado pelo swaggo/swag
│   └── openapi.yaml             ← export para Bruno
├── .env.example
├── go.mod
└── go.sum
```

---

## Schema do Banco de Dados

### Tabela: `users`
```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255)        NOT NULL,
    email         VARCHAR(255)        NOT NULL UNIQUE,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);
```

### Tabela: `weddings`
```sql
CREATE TABLE weddings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    slug        VARCHAR(255)  NOT NULL UNIQUE,
    bride_name  VARCHAR(255)  NOT NULL,
    groom_name  VARCHAR(255)  NOT NULL,
    date        DATE          NOT NULL,
    time        TIME,
    location    VARCHAR(500)  NOT NULL,
    city        VARCHAR(255),
    state       VARCHAR(2),
    description TEXT,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_weddings_slug ON weddings(slug);
CREATE INDEX idx_weddings_user_id ON weddings(user_id);
```

### Tabela: `wedding_photos`
```sql
CREATE TABLE wedding_photos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID          NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    url         VARCHAR(1000) NOT NULL,   -- URL pública (S3 ou local)
    storage_key VARCHAR(1000) NOT NULL,   -- chave para deletar do storage
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wedding_photos_wedding_id ON wedding_photos(wedding_id);
```

### Tabela: `wedding_links`
```sql
CREATE TABLE wedding_links (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID          NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    label       VARCHAR(255)  NOT NULL,
    url         VARCHAR(1000) NOT NULL,
    position    INTEGER       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wedding_links_wedding_id ON wedding_links(wedding_id);
```

### Tabela: `guests`
```sql
CREATE TYPE rsvp_status AS ENUM ('pending', 'confirmed', 'declined');

CREATE TABLE guests (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID        NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    status      rsvp_status NOT NULL DEFAULT 'pending',
    rsvp_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guests_wedding_id ON guests(wedding_id);
CREATE INDEX idx_guests_status ON guests(wedding_id, status);
```

### Tabela: `gifts`
```sql
CREATE TYPE gift_status AS ENUM ('available', 'reserved');

CREATE TABLE gifts (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id       UUID         NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    image_url        VARCHAR(1000),
    store_url        VARCHAR(1000),
    price            NUMERIC(10,2),
    status           gift_status  NOT NULL DEFAULT 'available',
    reserved_by_name VARCHAR(255),
    reserved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX idx_gifts_status ON gifts(wedding_id, status);
```

### Tabela: `refresh_tokens`
```sql
CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

---

## Contratos da API

### Envelope padrão de resposta

**Sucesso:**
```json
{
  "data": { ... },
  "message": "opcional"
}
```

**Erro:**
```json
{
  "error": "error_code",
  "message": "Mensagem legível para o frontend"
}
```

**Erro de validação (422):**
```json
{
  "error": "validation_error",
  "message": "Validation failed",
  "details": [
    { "field": "email", "message": "must be a valid email" },
    { "field": "password", "message": "minimum 8 characters" }
  ]
}
```

---

### Auth

#### `POST /v1/auth/register`
```json
// Request
{
  "name": "Ana e João",
  "email": "ana@email.com",
  "password": "minimo8chars"
}

// Response 201
{
  "data": {
    "id": "uuid",
    "name": "Ana e João",
    "email": "ana@email.com",
    "created_at": "2026-11-15T10:00:00Z"
  }
}
```

#### `POST /v1/auth/login`
```json
// Request
{
  "email": "ana@email.com",
  "password": "minimo8chars"
}

// Response 200
{
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_at": "2026-11-15T11:00:00Z"
  }
}
```

#### `POST /v1/auth/refresh`
```json
// Request
{ "refresh_token": "eyJ..." }

// Response 200
{
  "data": {
    "access_token": "eyJ...",
    "expires_at": "2026-11-15T12:00:00Z"
  }
}
```

#### `POST /v1/auth/logout`
```
// Header: Authorization: Bearer <token>
// Response 204 No Content
```

---

### Casamento (autenticado)

#### `POST /v1/wedding`
```json
// Request
{
  "bride_name": "Ana",
  "groom_name": "João",
  "date": "2026-11-15",
  "time": "17:00",
  "location": "Espaço Villa Rica",
  "city": "São Paulo",
  "state": "SP",
  "description": "Nossa história...",
  "links": [
    { "label": "Buffet", "url": "https://buffet.com", "position": 1 }
  ]
}

// Response 201
{
  "data": {
    "id": "uuid",
    "slug": "ana-e-joao",
    "bride_name": "Ana",
    "groom_name": "João",
    "date": "2026-11-15",
    "time": "17:00",
    "location": "Espaço Villa Rica",
    "city": "São Paulo",
    "state": "SP",
    "description": "Nossa história...",
    "photos": [],
    "links": [{ "id": "uuid", "label": "Buffet", "url": "https://buffet.com", "position": 1 }],
    "created_at": "2026-04-13T10:00:00Z"
  }
}
```

#### `GET /v1/wedding`
```json
// Response 200 — mesmo shape do POST, com fotos e links populados
```

#### `PATCH /v1/wedding`
```json
// Request — todos os campos são opcionais
{
  "date": "2026-12-01",
  "description": "Nova descrição"
}
// Response 200 — wedding atualizado completo
```

#### `POST /v1/wedding/photos`
```
// Content-Type: multipart/form-data
// Campo: file (JPEG | PNG | WebP, máx 10MB)

// Response 201
{
  "data": {
    "id": "uuid",
    "url": "https://bucket.s3.amazonaws.com/weddings/uuid/foto.jpg",
    "created_at": "..."
  }
}
```

#### `DELETE /v1/wedding/photos/:photo_id`
```
// Response 204 No Content
```

---

### Convidados (autenticado)

#### `POST /v1/guests`
```json
// Request
{ "name": "Maria Silva" }

// Response 201
{
  "data": {
    "id": "uuid",
    "name": "Maria Silva",
    "status": "pending",
    "rsvp_at": null,
    "created_at": "..."
  }
}
```

#### `GET /v1/guests?status=pending`
```json
// Response 200
{
  "data": [
    { "id": "uuid", "name": "Maria Silva", "status": "pending", "rsvp_at": null },
    { "id": "uuid", "name": "José Santos", "status": "confirmed", "rsvp_at": "..." }
  ]
}
```

#### `PATCH /v1/guests/:guest_id`
```json
// Request
{ "name": "Maria da Silva" }
// Response 200 — guest atualizado
```

#### `DELETE /v1/guests/:guest_id`
```
// Response 204 No Content
```

#### `GET /v1/guests/summary`
```json
// Response 200
{
  "data": {
    "total": 50,
    "confirmed": 32,
    "declined": 5,
    "pending": 13
  }
}
```

---

### Presentes (autenticado)

#### `POST /v1/gifts`
```json
// Request
{
  "name": "Jogo de panelas",
  "description": "Tramontina 7 peças",
  "image_url": "https://...",
  "store_url": "https://amazon.com.br/...",
  "price": 350.00
}

// Response 201
{
  "data": {
    "id": "uuid",
    "name": "Jogo de panelas",
    "description": "Tramontina 7 peças",
    "image_url": "https://...",
    "store_url": "https://amazon.com.br/...",
    "price": 350.00,
    "status": "available",
    "reserved_by_name": null,
    "reserved_at": null,
    "created_at": "..."
  }
}
```

#### `GET /v1/gifts?status=available`
```json
// Response 200
{
  "data": [
    {
      "id": "uuid",
      "name": "Jogo de panelas",
      "status": "available",
      "reserved_by_name": null,
      ...
    }
  ]
}
```

#### `PATCH /v1/gifts/:gift_id`
```json
// Request — campos opcionais
{ "price": 380.00 }
// Response 200 — gift atualizado
```

#### `DELETE /v1/gifts/:gift_id`
```
// Response 204 No Content
```

#### `GET /v1/gifts/summary`
```json
// Response 200
{
  "data": {
    "total": 20,
    "available": 14,
    "reserved": 6
  }
}
```

#### `DELETE /v1/gifts/:gift_id/reserve` (cancela reserva — só casal)
```
// Response 200
{
  "data": { "id": "uuid", "status": "available", ... }
}
```

---

### Página Pública (sem autenticação)

#### `GET /v1/public/:slug`
```json
// Response 200
{
  "data": {
    "slug": "ana-e-joao",
    "bride_name": "Ana",
    "groom_name": "João",
    "date": "2026-11-15",
    "time": "17:00",
    "location": "Espaço Villa Rica",
    "city": "São Paulo",
    "state": "SP",
    "description": "Nossa história...",
    "photos": ["https://..."],
    "links": [{ "label": "Buffet", "url": "https://buffet.com" }]
  }
}
```

#### `GET /v1/public/:slug/guests`
```json
// Response 200
{
  "data": [
    { "id": "uuid", "name": "Maria Silva", "status": "pending" },
    { "id": "uuid", "name": "José Santos", "status": "confirmed" }
  ]
}
```

#### `POST /v1/public/:slug/guests/:guest_id/rsvp`
```json
// Request
{ "status": "confirmed" }  // ou "declined"

// Response 200
{
  "data": {
    "id": "uuid",
    "name": "Maria Silva",
    "status": "confirmed",
    "rsvp_at": "2026-04-13T10:00:00Z"
  }
}
```

#### `GET /v1/public/:slug/gifts`
```json
// Response 200
{
  "data": [
    {
      "id": "uuid",
      "name": "Jogo de panelas",
      "description": "Tramontina 7 peças",
      "image_url": "https://...",
      "store_url": "https://amazon.com.br/...",
      "price": 350.00,
      "reserved": false
    },
    {
      "id": "uuid",
      "name": "Air Fryer",
      "reserved": true
      // reserved_by_name NÃO é exposto publicamente
    }
  ]
}
```

#### `POST /v1/public/:slug/gifts/:gift_id/reserve`
```json
// Request
{ "guest_name": "Maria Silva" }

// Response 200
{
  "data": {
    "id": "uuid",
    "name": "Jogo de panelas",
    "reserved": true
  }
}

// Response 409 se já reservado
{
  "error": "already_reserved",
  "message": "Este presente já foi reservado"
}
```

---

## Componentes

### `internal/config`

- **Propósito:** Carregar e expor configuração da aplicação via env vars
- **Interfaces:**
  - `Load() (*Config, error)` — lê `.env` (dev) ou env vars diretas (prod)
- **Config struct:**
  ```go
  type Config struct {
      DatabaseURL     string
      JWTSecret       string
      JWTExpiry       time.Duration  // default: 1h
      RefreshExpiry   time.Duration  // default: 7d
      StorageDriver   string         // "s3" | "local"
      S3Bucket        string
      S3Region        string
      LocalStoragePath string
      Port            string         // default: "8080"
  }
  ```

### `internal/domain`

- **Propósito:** Contratos (interfaces) e tipos compartilhados entre camadas
- Nenhuma dependência de framework ou banco
- Interfaces de repositório definidas aqui — implementadas em `repository/`
- Erros de domínio tipados (evita strings mágicas):
  ```go
  var (
      ErrNotFound      = errors.New("not_found")
      ErrConflict      = errors.New("conflict")
      ErrUnauthorized  = errors.New("unauthorized")
      ErrForbidden     = errors.New("forbidden")
  )
  ```

### `internal/handler`

- **Propósito:** Receber e responder requisições HTTP
- **Responsabilidades:**
  - Parse de body/path params/query params
  - Validação de input (via `go-playground/validator`)
  - Chamar o service adequado
  - Serializar resposta no envelope padrão
- **NÃO contém:** lógica de negócio, queries SQL
- `response.go` centraliza `JSON(w, status, data)`, `Error(w, status, code, msg)`, `ValidationError(w, errs)`

### `internal/service`

- **Propósito:** Lógica de negócio pura
- **Responsabilidades:**
  - Orquestrar repositórios
  - Gerar slug (service/wedding.go)
  - Hash de senha e geração de JWT (service/auth.go)
  - Reserva atômica de presente via transação (service/gift.go)
  - Upload e deleção de arquivos via StorageService (service/wedding.go)
- **NÃO contém:** código HTTP, queries SQL diretas

### `internal/repository`

- **Propósito:** Acesso ao banco de dados via sqlx
- **Responsabilidades:** Queries SQL puras, mapeamento para structs de domínio
- **NÃO contém:** lógica de negócio
- Cada repositório implementa a interface definida em `domain/`

### `internal/service/storage.go`

- **Propósito:** Abstração do storage de arquivos
- **Interface:**
  ```go
  type StorageService interface {
      Upload(ctx context.Context, key string, file io.Reader, contentType string) (url string, err error)
      Delete(ctx context.Context, key string) error
  }
  ```
- **Implementações:**
  - `S3Storage` — usa AWS SDK v2
  - `LocalStorage` — salva em disco, serve via endpoint estático

---

## Estratégia de Erros

| Erro de domínio   | HTTP Status | `error` code          |
|-------------------|-------------|----------------------|
| `ErrNotFound`     | 404         | `not_found`          |
| `ErrConflict`     | 409         | `conflict` / custom  |
| `ErrUnauthorized` | 401         | `unauthorized`       |
| `ErrForbidden`    | 403         | `forbidden`          |
| Validação falha   | 422         | `validation_error`   |
| Erro inesperado   | 500         | `internal_error`     |

Handler mapeia erros de domínio para HTTP usando `errors.Is()`. Erros desconhecidos retornam 500 sem expor detalhes internos.

---

## Decisões Técnicas

| Decisão | Escolha | Razão |
|---|---|---|
| Reserva de presente | Transação SQL com `SELECT FOR UPDATE` | Garante atomicidade — evita reserva dupla em requisições concorrentes |
| Geração de slug | `bride_name + "-e-" + groom_name`, slugificado + sufixo numérico se conflito | Simples, legível, único |
| Hash de senha | bcrypt (custo 12) | Padrão de mercado, resistente a brute-force |
| Refresh token no banco | Tabela `refresh_tokens` com hash do token | Permite revogação individual sem blacklist em memória |
| Fotos: chave de storage | `weddings/{wedding_id}/{uuid}.{ext}` | Namespacing por casamento, sem colisão |
| Validação | `go-playground/validator` via struct tags | Padrão Go, integra bem com Chi |
| UUID | `gen_random_uuid()` no Postgres | Sem dependência de lib externa no Go para geração |
