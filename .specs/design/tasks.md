# Tasks — wedding-mc

**Design:** `.specs/design/architecture.md`
**Status:** Approved

---

## Plano de Execução

```
Phase 1 (Sequential — Fundação):
  T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9 → T10

Phase 2 (Parallel — Middleware):
  T10 completo, então:
    ├── T11 [P]
    ├── T12 [P]
    └── T13 [P]
  T11+T12+T13 completos → T14

Phase 3 (Sequential — Auth):
  T14 → T15 → T16 → T17 → T18

Phase 4 (Sequential — Wedding):
  T18 → T19 → T20 → T21 → T22 → T23

Phase 5 (Parallel — Guests + Gifts):
  T23 completo, então em paralelo:
    ├── T24 → T25 → T26 [P]   (Guests)
    └── T27 → T28 → T29 [P]   (Gifts)

Phase 6 (Sequential — Página Pública + Main):
  T26 + T29 completos → T30 → T31

Phase 7 (Sequential — Swagger + OpenAPI):
  T31 → T32 → T33

Phase 8 (Parallel — Testes):
  T31 completo, então em paralelo:
    ├── T34 [P]   unit: AuthService
    ├── T35 [P]   unit: WeddingService
    ├── T36 [P]   unit: GuestService
    └── T37 [P]   unit: GiftService
  T34+T35+T36+T37 completos → T38 → T39 → T40 → T41 → T42
```

---

## Task Breakdown

---

### T1: Inicializar módulo Go e instalar dependências

**What:** Criar `go.mod` com todas as dependências do projeto
**Where:** `/go.mod`, `/go.sum`
**Depends on:** Nenhuma
**Requirement:** Fundação

**Dependências a instalar:**
```
github.com/go-chi/chi/v5
github.com/go-chi/cors
github.com/jmoiron/sqlx
github.com/lib/pq
github.com/golang-migrate/migrate/v4
github.com/golang-jwt/jwt/v5
github.com/go-playground/validator/v10
github.com/rs/zerolog
github.com/joho/godotenv
github.com/aws/aws-sdk-go-v2
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/service/s3
github.com/google/uuid
golang.org/x/crypto
github.com/swaggo/swag
github.com/swaggo/http-swagger
```

**Dependências de teste:**
```
github.com/stretchr/testify
github.com/testcontainers/testcontainers-go
```

**Done when:**
- [ ] `go.mod` criado com module path `github.com/[user]/wedding-mc`
- [ ] `go mod tidy` executa sem erros
- [ ] `go build ./...` compila (mesmo que vazio)

---

### T2: Criar estrutura de pastas do projeto

**What:** Criar todos os diretórios e arquivos `.gitkeep` conforme o design
**Where:** Toda a estrutura raiz do projeto
**Depends on:** T1

**Estrutura:**
```
cmd/api/
internal/config/
internal/domain/
internal/handler/
internal/service/
internal/repository/
internal/middleware/
migrations/
docs/
```

**Done when:**
- [ ] Todas as pastas existem
- [ ] `go build ./...` ainda compila

---

### T3: Criar `.env.example` e `docker-compose.yml`

**What:** Arquivo de exemplo de variáveis de ambiente + compose para dev local
**Where:** `/.env.example`, `/docker-compose.yml`
**Depends on:** T1

**`.env.example` deve conter:**
```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/weddingmc?sslmode=disable
JWT_SECRET=change-me-in-production
JWT_EXPIRY=1h
REFRESH_EXPIRY=168h
STORAGE_DRIVER=local
LOCAL_STORAGE_PATH=./uploads
S3_BUCKET=
S3_REGION=
PORT=8080
```

**`docker-compose.yml` deve conter:**
- Serviço `postgres` (postgres:16) com healthcheck
- Volume para persistência de dados

**Done when:**
- [ ] `.env.example` criado com todas as vars documentadas
- [ ] `docker-compose up -d` sobe Postgres sem erros
- [ ] Postgres acessível em `localhost:5432`

---

### T4: Implementar `internal/config/config.go`

**What:** Struct `Config` que lê todas as env vars com validação e defaults
**Where:** `internal/config/config.go`
**Depends on:** T2

**Done when:**
- [ ] Struct `Config` com todos os campos do design
- [ ] `Load() (*Config, error)` lê `.env` via godotenv em dev, env vars diretas em prod
- [ ] Campos obrigatórios ausentes retornam erro descritivo
- [ ] Defaults aplicados para `PORT=8080`, `JWT_EXPIRY=1h`, `REFRESH_EXPIRY=168h`
- [ ] `go build ./internal/config/...` compila sem erros

---

### T5: Implementar `internal/domain/errors.go`

**What:** Erros de domínio tipados usados em todas as camadas
**Where:** `internal/domain/errors.go`
**Depends on:** T2
**Requirement:** AUTH-01, GUEST-01, GIFT-01, PUB-02

**Done when:**
- [ ] Vars exportadas: `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrForbidden`, `ErrValidation`
- [ ] Compatíveis com `errors.Is()`

---

### T6: Implementar structs de domínio

**What:** Structs Go para todas as entidades do sistema
**Where:** `internal/domain/user.go`, `wedding.go`, `guest.go`, `gift.go`
**Depends on:** T5

**Structs:**
- `user.go`: `User{ID, Name, Email, PasswordHash, CreatedAt, UpdatedAt}`
- `wedding.go`: `Wedding{...}`, `WeddingPhoto{ID, WeddingID, URL, StorageKey, CreatedAt}`, `WeddingLink{ID, WeddingID, Label, URL, Position, CreatedAt}`
- `guest.go`: `Guest{ID, WeddingID, Name, Status RSVPStatus, RSVPAt, CreatedAt, UpdatedAt}`, `RSVPStatus` enum (`pending`/`confirmed`/`declined`)
- `gift.go`: `Gift{ID, WeddingID, Name, Description, ImageURL, StoreURL, Price, Status GiftStatus, ReservedByName, ReservedAt, CreatedAt, UpdatedAt}`, `GiftStatus` enum (`available`/`reserved`)

**Done when:**
- [ ] Todas as structs definidas com tags `db:""` (para sqlx) e `json:""`
- [ ] Enums como `string` types com constantes
- [ ] `go build ./internal/domain/...` compila

---

### T7: Implementar interfaces de repositório no domínio

**What:** Interfaces Go para cada repositório — contratos entre service e repository
**Where:** `internal/domain/user.go`, `wedding.go`, `guest.go`, `gift.go` (junto às structs)
**Depends on:** T6
**Requirement:** AUTH-01, WED-01, GUEST-01, GIFT-01

**Interfaces:**
```go
// user.go
type UserRepository interface {
    Create(ctx, *User) error
    FindByEmail(ctx, email string) (*User, error)
    FindByID(ctx, id string) (*User, error)
}

// wedding.go
type WeddingRepository interface {
    Create(ctx, *Wedding) error
    FindByUserID(ctx, userID string) (*Wedding, error)
    FindBySlug(ctx, slug string) (*Wedding, error)
    Update(ctx, *Wedding) error
    AddPhoto(ctx, *WeddingPhoto) error
    DeletePhoto(ctx, photoID string) (*WeddingPhoto, error)
    FindPhotoByID(ctx, photoID string) (*WeddingPhoto, error)
    ReplaceLinks(ctx, weddingID string, links []WeddingLink) error
}

// guest.go
type GuestRepository interface {
    Create(ctx, *Guest) error
    FindAll(ctx, weddingID string, status *RSVPStatus) ([]Guest, error)
    FindByID(ctx, id string) (*Guest, error)
    Update(ctx, *Guest) error
    Delete(ctx, id string) error
    CountByStatus(ctx, weddingID string) (map[RSVPStatus]int, error)
}

// gift.go
type GiftRepository interface {
    Create(ctx, *Gift) error
    FindAll(ctx, weddingID string, status *GiftStatus) ([]Gift, error)
    FindByID(ctx, id string) (*Gift, error)
    Update(ctx, *Gift) error
    Delete(ctx, id string) error
    Reserve(ctx, giftID, guestName string) error  // SELECT FOR UPDATE em transação
    CancelReserve(ctx, giftID string) error
    CountByStatus(ctx, weddingID string) (map[GiftStatus]int, error)
}

// refresh_token.go (novo arquivo)
type RefreshTokenRepository interface {
    Create(ctx, *RefreshToken) error
    FindByHash(ctx, hash string) (*RefreshToken, error)
    RevokeByUserID(ctx, userID string) error
}
```

**Done when:**
- [ ] Todas as interfaces definidas e exportadas
- [ ] `go build ./internal/domain/...` compila

---

### T8: Implementar `internal/handler/response.go`

**What:** Helpers para serializar respostas HTTP no envelope padrão
**Where:** `internal/handler/response.go`
**Depends on:** T5

**Funções:**
```go
func JSON(w http.ResponseWriter, status int, data any)
func Error(w http.ResponseWriter, status int, code, message string)
func ValidationError(w http.ResponseWriter, errs validator.ValidationErrors)
func NoContent(w http.ResponseWriter)
```

**Done when:**
- [ ] `JSON` escreve `{"data": ...}` com status e Content-Type corretos
- [ ] `Error` escreve `{"error": "...", "message": "..."}` 
- [ ] `ValidationError` escreve `{"error": "validation_error", "details": [...]}`
- [ ] `go build ./internal/handler/...` compila

---

### T9: Criar migrations SQL

**What:** Arquivos `.up.sql` e `.down.sql` para todas as 7 tabelas
**Where:** `migrations/`
**Depends on:** T2
**Requirement:** Todas

**Arquivos:**
```
000001_create_users.up.sql / .down.sql
000002_create_weddings.up.sql / .down.sql
000003_create_wedding_photos.up.sql / .down.sql
000004_create_wedding_links.up.sql / .down.sql
000005_create_guests.up.sql / .down.sql
000006_create_gifts.up.sql / .down.sql
000007_create_refresh_tokens.up.sql / .down.sql
```

**SQL exatamente conforme `.specs/design/architecture.md` — Schema do Banco de Dados.**

**Done when:**
- [ ] Todos os 14 arquivos criados
- [ ] `migrate -database $DATABASE_URL -path migrations up` aplica sem erros
- [ ] `migrate -database $DATABASE_URL -path migrations down` reverte sem erros
- [ ] Schema final bate com o design (verificar via `\d` no psql)

---

### T10: Implementar conexão com banco e auto-migrate no startup

**What:** Função que conecta ao Postgres via sqlx e aplica migrations na inicialização
**Where:** `internal/config/database.go`
**Depends on:** T4, T9

**Done when:**
- [ ] `NewDB(cfg *Config) (*sqlx.DB, error)` conecta e faz ping
- [ ] Migrations são aplicadas automaticamente ao subir a aplicação
- [ ] Erro de conexão encerra a aplicação com log descritivo
- [ ] `go build ./internal/config/...` compila

---

### T11: Middleware de logging [P]

**What:** Middleware Chi que loga cada requisição com zerolog
**Where:** `internal/middleware/logger.go`
**Depends on:** T10

**Log deve incluir:** método, path, status, duração, request ID

**Done when:**
- [ ] Cada request logado em JSON via zerolog
- [ ] Request ID gerado por request e propagado no contexto
- [ ] Não loga body (segurança)

---

### T12: Middleware de recover (panic handler) [P]

**What:** Middleware que captura panics e retorna 500
**Where:** `internal/middleware/recover.go`
**Depends on:** T10

**Done when:**
- [ ] Panic capturado → response `500` com `{"error": "internal_error"}`
- [ ] Stack trace logado via zerolog (não exposto no response)

---

### T13: Middleware de CORS [P]

**What:** Configuração de CORS para permitir acesso do frontend externo
**Where:** `internal/middleware/cors.go`
**Depends on:** T10

**Done when:**
- [ ] Origins configurados via env var `ALLOWED_ORIGINS` (default: `*` em dev)
- [ ] Métodos e headers padrão REST permitidos
- [ ] Preflight `OPTIONS` respondido corretamente

---

### T14: Middleware de autenticação JWT

**What:** Middleware que valida Bearer token e injeta `userID` no contexto da requisição
**Where:** `internal/middleware/auth.go`
**Depends on:** T11, T12, T13
**Requirement:** AUTH-03

**Done when:**
- [ ] Header `Authorization: Bearer <token>` extraído e validado
- [ ] `userID` injetado no `context.Context` via chave tipada
- [ ] Request sem token ou com token inválido → `401` imediato
- [ ] Helper `UserIDFromContext(ctx) (string, bool)` exportado para uso nos handlers

---

### T15: Implementar `internal/repository/user.go`

**What:** Implementação concreta de `UserRepository` usando sqlx
**Where:** `internal/repository/user.go`
**Depends on:** T7, T10
**Requirement:** AUTH-01

**Queries:**
- `Create`: INSERT com retorno do ID gerado
- `FindByEmail`: SELECT por email
- `FindByID`: SELECT por ID

**Done when:**
- [ ] Implementa interface `domain.UserRepository`
- [ ] `go build ./internal/repository/...` compila

---

### T16: Implementar `internal/repository/refresh_token.go`

**What:** Implementação de `RefreshTokenRepository` para controle de sessões
**Where:** `internal/repository/refresh_token.go`
**Depends on:** T7, T10
**Requirement:** AUTH-04, AUTH-05

**Done when:**
- [ ] Implementa interface `domain.RefreshTokenRepository`
- [ ] `Create`, `FindByHash`, `RevokeByUserID` funcionando

---

### T17: Implementar `internal/service/auth.go`

**What:** Lógica de negócio de autenticação — registro, login, tokens
**Where:** `internal/service/auth.go`
**Depends on:** T15, T16
**Requirement:** AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05

**Responsabilidades:**
- `Register`: valida, hash bcrypt (custo 12), cria user
- `Login`: busca user, compara hash, gera access JWT + refresh token (armazena hash no banco)
- `RefreshToken`: valida refresh token, revoga o antigo, gera novo par
- `Logout`: revoga todos refresh tokens do usuário
- `GenerateAccessToken` / `ValidateToken`: usando `golang-jwt/jwt/v5`

**Done when:**
- [ ] Senha nunca armazenada em plain text
- [ ] JWT contém: `sub` (userID), `exp`, `iat`
- [ ] Refresh token: valor aleatório seguro, armazenado como SHA-256 hash no banco
- [ ] `go build ./internal/service/...` compila
- [ ] Unit tests cobrindo: registro duplicado, senha errada, token expirado (T34)

---

### T18: Implementar `internal/handler/auth.go`

**What:** Handlers HTTP para os 4 endpoints de autenticação
**Where:** `internal/handler/auth.go`
**Depends on:** T17, T8, T14
**Requirement:** AUTH-01, AUTH-02, AUTH-04, AUTH-05

**Endpoints:**
- `POST /v1/auth/register` → `Register`
- `POST /v1/auth/login` → `Login`
- `POST /v1/auth/refresh` → `Refresh`
- `POST /v1/auth/logout` (autenticado) → `Logout`

**Done when:**
- [ ] Validação via `go-playground/validator` com mensagens descritivas
- [ ] Senha nunca retornada em nenhum response
- [ ] Todos os status codes conforme spec (201, 200, 401, 409, 422)
- [ ] `go build ./internal/handler/...` compila

---

### T19: Implementar `internal/service/storage.go`

**What:** Interface `StorageService` + implementações `LocalStorage` e `S3Storage`
**Where:** `internal/service/storage.go`
**Depends on:** T4
**Requirement:** WED-04

**Interface:**
```go
type StorageService interface {
    Upload(ctx, key string, r io.Reader, contentType string) (publicURL string, err error)
    Delete(ctx, key string) error
}
```

**LocalStorage:** salva em `LOCAL_STORAGE_PATH/{key}`, retorna URL relativa
**S3Storage:** usa AWS SDK v2, retorna URL pública do S3

**Done when:**
- [ ] Seleção de implementação via `STORAGE_DRIVER` env var no `main.go`
- [ ] `LocalStorage.Upload` cria diretórios intermediários se não existirem
- [ ] `S3Storage.Upload` usa content-type correto
- [ ] `go build ./internal/service/...` compila

---

### T20: Implementar `internal/repository/wedding.go`

**What:** Implementação de `WeddingRepository` usando sqlx
**Where:** `internal/repository/wedding.go`
**Depends on:** T7, T10
**Requirement:** WED-01, WED-02, WED-03, WED-04, WED-05, WED-06

**Queries:**
- `Create`: INSERT wedding, retorna com ID e slug
- `FindByUserID`: SELECT com JOIN em photos e links
- `FindBySlug`: SELECT com JOIN em photos e links (para página pública)
- `Update`: UPDATE campos editáveis
- `AddPhoto` / `DeletePhoto` / `FindPhotoByID`
- `ReplaceLinks`: DELETE + INSERT em transação

**Done when:**
- [ ] Implementa interface `domain.WeddingRepository`
- [ ] `FindBySlug` e `FindByUserID` populam `Photos []WeddingPhoto` e `Links []WeddingLink`

---

### T21: Implementar `internal/service/wedding.go`

**What:** Lógica de negócio do módulo de casamento
**Where:** `internal/service/wedding.go`
**Depends on:** T20, T19
**Requirement:** WED-01, WED-02, WED-03, WED-04, WED-05, WED-06, PUB-01

**Responsabilidades:**
- `CreateWedding`: gera slug (`ana-e-joao`), verifica unicidade, cria
- `generateSlug(bride, groom string) string`: slugifica nomes, adiciona sufixo `-2`, `-3` se conflito
- `UpdateWedding`: valida propriedade, atualiza
- `UploadPhoto`: valida tipo/tamanho, gera chave `weddings/{id}/{uuid}.{ext}`, chama StorageService
- `DeletePhoto`: busca photo, deleta do storage, deleta do banco

**Done when:**
- [ ] Slug gerado é sempre `[a-z0-9-]`, único no banco
- [ ] Upload rejeita arquivos > 10MB e tipos não-imagem
- [ ] `go build ./internal/service/...` compila
- [ ] Unit tests cobrindo geração de slug com colisões (T35)

---

### T22: Implementar `internal/handler/wedding.go`

**What:** Handlers HTTP do módulo de casamento
**Where:** `internal/handler/wedding.go`
**Depends on:** T21, T8, T14
**Requirement:** WED-01, WED-02, WED-03, WED-04, WED-05, WED-06

**Endpoints:**
- `GET /v1/wedding`
- `POST /v1/wedding`
- `PATCH /v1/wedding`
- `POST /v1/wedding/photos` (multipart/form-data)
- `DELETE /v1/wedding/photos/{photoID}`

**Done when:**
- [ ] Todos protegidos pelo middleware de auth (exceto nenhum aqui)
- [ ] Upload parse correto de `multipart/form-data`
- [ ] Status codes conforme spec

---

### T23: Implementar `internal/repository/guest.go`

**What:** Implementação de `GuestRepository`
**Where:** `internal/repository/guest.go`
**Depends on:** T7, T10
**Requirement:** GUEST-01, GUEST-02, GUEST-03, GUEST-04, GUEST-05, GUEST-06

**Done when:**
- [ ] Implementa interface `domain.GuestRepository`
- [ ] `FindAll` suporta filtro opcional por status
- [ ] `CountByStatus` retorna map com os 3 status

---

### T24: Implementar `internal/service/guest.go` [P]

**What:** Lógica de negócio do módulo de convidados
**Where:** `internal/service/guest.go`
**Depends on:** T23
**Requirement:** GUEST-01, GUEST-02, GUEST-03, GUEST-04, GUEST-05, GUEST-06

**Responsabilidades:**
- `CreateGuest`: verifica que wedding pertence ao user, cria guest
- `ListGuests`: lista com filtro de status
- `UpdateGuest`: verifica propriedade do guest
- `DeleteGuest`: verifica propriedade
- `RSVP`: atualiza status + `rsvp_at` via slug (endpoint público)
- `GetSummary`: retorna contagens por status

**Done when:**
- [ ] Guest de outro casamento nunca modificável
- [ ] `RSVP` aceita `confirmed` e `declined`, rejeita outros valores
- [ ] Unit tests cobrindo RSVP com status inválido (T36)

---

### T25: Implementar `internal/handler/guest.go` [P]

**What:** Handlers HTTP do módulo de convidados
**Where:** `internal/handler/guest.go`
**Depends on:** T24, T8, T14
**Requirement:** GUEST-01..GUEST-06

**Endpoints:**
- `POST /v1/guests`
- `GET /v1/guests`
- `GET /v1/guests/summary`
- `PATCH /v1/guests/{guestID}`
- `DELETE /v1/guests/{guestID}`

**Done when:**
- [ ] Todos protegidos pelo middleware de auth
- [ ] `?status=` query param funciona corretamente

---

### T26: Implementar `internal/repository/gift.go` [P]

**What:** Implementação de `GiftRepository` com reserva atômica
**Where:** `internal/repository/gift.go`
**Depends on:** T7, T10
**Requirement:** GIFT-01..GIFT-07

**Detalhe crítico — `Reserve`:**
```sql
BEGIN;
SELECT id FROM gifts WHERE id = $1 AND wedding_id = $2 AND status = 'available' FOR UPDATE;
-- se não encontrar: ROLLBACK + retorna ErrConflict
UPDATE gifts SET status = 'reserved', reserved_by_name = $3, reserved_at = NOW() WHERE id = $1;
COMMIT;
```

**Done when:**
- [ ] Implementa interface `domain.GiftRepository`
- [ ] `Reserve` usa transação com `SELECT FOR UPDATE`
- [ ] Duas goroutines tentando reservar o mesmo gift: apenas uma tem sucesso, outra recebe `ErrConflict`

---

### T27: Implementar `internal/service/gift.go` [P]

**What:** Lógica de negócio do módulo de presentes
**Where:** `internal/service/gift.go`
**Depends on:** T26
**Requirement:** GIFT-01..GIFT-07

**Done when:**
- [ ] `ReserveGift` verifica que gift pertence ao wedding_slug correto antes de reservar
- [ ] `CancelReserve` verifica que casal autenticado é dono do casamento
- [ ] Unit tests cobrindo reserva dupla (T37)

---

### T28: Implementar `internal/handler/gift.go` [P]

**What:** Handlers HTTP do módulo de presentes
**Where:** `internal/handler/gift.go`
**Depends on:** T27, T8, T14
**Requirement:** GIFT-01..GIFT-07

**Endpoints:**
- `POST /v1/gifts`
- `GET /v1/gifts`
- `GET /v1/gifts/summary`
- `PATCH /v1/gifts/{giftID}`
- `DELETE /v1/gifts/{giftID}`
- `DELETE /v1/gifts/{giftID}/reserve`

**Done when:**
- [ ] `409` retornado quando presente já reservado
- [ ] `reserved_by_name` visível apenas nas rotas autenticadas do casal

---

### T29: Implementar `internal/handler/public.go`

**What:** Handlers dos endpoints públicos (sem autenticação)
**Where:** `internal/handler/public.go`
**Depends on:** T25, T28
**Requirement:** PUB-01, PUB-02, PUB-03, PUB-04, GUEST-05, GIFT-05

**Endpoints:**
- `GET /v1/public/{slug}` — página do casamento
- `GET /v1/public/{slug}/guests` — lista de convidados para o select
- `POST /v1/public/{slug}/guests/{guestID}/rsvp` — confirmar presença
- `GET /v1/public/{slug}/gifts` — lista pública de presentes
- `POST /v1/public/{slug}/gifts/{giftID}/reserve` — reservar presente

**Done when:**
- [ ] `reserved_by_name` NUNCA exposto na lista pública de gifts
- [ ] Response de gifts públicos usa `reserved: bool` (não o enum interno)
- [ ] Nenhum dado sensível do casal exposto (email, password_hash)
- [ ] `404` quando slug não existe

---

### T30: Implementar `cmd/api/main.go`

**What:** Entrypoint da aplicação — DI manual, roteamento, inicialização
**Where:** `cmd/api/main.go`
**Depends on:** T29
**Requirement:** Todos

**Responsabilidades:**
- Carregar config
- Conectar ao banco e aplicar migrations
- Instanciar repositórios, services e handlers (DI manual)
- Montar router Chi com todos os middlewares e rotas
- Iniciar servidor HTTP

**Estrutura de rotas:**
```
/v1/auth/*          → AuthHandler (público)
/v1/wedding/*       → WeddingHandler (autenticado)
/v1/guests/*        → GuestHandler (autenticado)
/v1/gifts/*         → GiftHandler (autenticado)
/v1/public/*        → PublicHandler (público)
```

**Done when:**
- [ ] `go run ./cmd/api` sobe sem erros com Postgres rodando
- [ ] Todos os endpoints respondem (mesmo que com 401/404)
- [ ] `GET /health` retorna `200 {"status": "ok"}`
- [ ] Graceful shutdown ao receber SIGTERM/SIGINT

---

### T31: Adicionar anotações Swagger em todos os handlers

**What:** Comentários `swaggo` em todos os handlers para geração automática da documentação
**Where:** `internal/handler/*.go` + `cmd/api/main.go`
**Depends on:** T30
**Requirement:** Documentação

**Done when:**
- [ ] Todos os endpoints documentados com `@Summary`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`
- [ ] Schemas de request/response definidos via structs com tags `swaggertype`
- [ ] Auth JWT documentado com `@Security BearerAuth`

---

### T32: Gerar documentação Swagger

**What:** Executar `swag init` para gerar os arquivos em `docs/`
**Where:** `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`
**Depends on:** T31

**Done when:**
- [ ] `swag init -g cmd/api/main.go` executa sem erros
- [ ] `GET /swagger/index.html` (em dev) renderiza a documentação
- [ ] Todos os endpoints aparecem documentados

---

### T33: Exportar OpenAPI para Bruno

**What:** Gerar `docs/openapi.yaml` formatado para importação no Bruno
**Where:** `docs/openapi.yaml`
**Depends on:** T32

**Done when:**
- [ ] `docs/openapi.yaml` válido no formato OpenAPI 3.0
- [ ] Importado no Bruno sem erros
- [ ] Todos os endpoints com exemplos de request/response

---

### T34: Testes unitários — `AuthService` [P]

**What:** Testes de unidade para `internal/service/auth.go`
**Where:** `internal/service/auth_test.go`
**Depends on:** T30
**Requirement:** AUTH-01, AUTH-02, AUTH-04, AUTH-05

**Casos de teste:**
- Registro com email duplicado → `ErrConflict`
- Login com senha errada → `ErrUnauthorized`
- Login com email inexistente → `ErrUnauthorized`
- Token JWT gerado tem `sub` e `exp` corretos
- Refresh token inválido → `ErrUnauthorized`
- Refresh token revogado → `ErrUnauthorized`

**Done when:**
- [ ] `go test ./internal/service/ -run TestAuth -v` passa
- [ ] Cobertura do AuthService ≥ 80%

---

### T35: Testes unitários — `WeddingService` [P]

**What:** Testes de unidade para `internal/service/wedding.go`
**Where:** `internal/service/wedding_test.go`
**Depends on:** T30
**Requirement:** WED-01, WED-04, PUB-01

**Casos de teste:**
- `generateSlug("Ana", "João")` → `"ana-e-joao"`
- `generateSlug` com slug existente → `"ana-e-joao-2"`
- Upload de arquivo não-imagem → erro
- Upload de arquivo > 10MB → erro
- Criar segundo casamento para mesmo user → `ErrConflict`

**Done when:**
- [ ] `go test ./internal/service/ -run TestWedding -v` passa
- [ ] Cobertura do WeddingService ≥ 80%

---

### T36: Testes unitários — `GuestService` [P]

**What:** Testes de unidade para `internal/service/guest.go`
**Where:** `internal/service/guest_test.go`
**Depends on:** T30
**Requirement:** GUEST-05, GUEST-06

**Casos de teste:**
- RSVP com status inválido → erro de validação
- RSVP de guest de outro casamento → `ErrNotFound`
- `GetSummary` com mix de status → contagens corretas

**Done when:**
- [ ] `go test ./internal/service/ -run TestGuest -v` passa
- [ ] Cobertura do GuestService ≥ 80%

---

### T37: Testes unitários — `GiftService` [P]

**What:** Testes de unidade para `internal/service/gift.go`
**Where:** `internal/service/gift_test.go`
**Depends on:** T30
**Requirement:** GIFT-05, GIFT-06, GIFT-07

**Casos de teste:**
- Reservar presente disponível → sucesso
- Reservar presente já reservado → `ErrConflict`
- Cancelar reserva de presente não-reservado → erro
- `GetSummary` com mix de status → contagens corretas

**Done when:**
- [ ] `go test ./internal/service/ -run TestGift -v` passa
- [ ] Cobertura do GiftService ≥ 80%

---

### T38: Testes e2e — endpoints de Auth

**What:** Testes de integração usando testcontainers (Postgres real)
**Where:** `internal/handler/auth_test.go`
**Depends on:** T34
**Requirement:** AUTH-01, AUTH-02, AUTH-04, AUTH-05

**Cenários:**
- `POST /register` com dados válidos → 201
- `POST /register` com email duplicado → 409
- `POST /register` com email inválido → 422
- `POST /login` com credenciais corretas → 200 com tokens
- `POST /login` com senha errada → 401
- `POST /refresh` com token válido → 200 com novo access token
- `POST /logout` → 204; reutilizar token → 401

**Done when:**
- [ ] `go test ./internal/handler/ -run TestAuthE2E -v` passa
- [ ] Banco limpo entre cada teste (rollback ou truncate)

---

### T39: Testes e2e — endpoints de Wedding

**What:** Testes de integração do módulo de casamento
**Where:** `internal/handler/wedding_test.go`
**Depends on:** T35, T38
**Requirement:** WED-01..WED-06

**Cenários:**
- Criar casamento → 201 com slug gerado
- Criar segundo casamento → 409
- Editar casamento → 200 com campos atualizados
- Upload de foto válida → 201 com URL
- Upload de arquivo não-imagem → 422
- Deletar foto → 204

**Done when:**
- [ ] `go test ./internal/handler/ -run TestWeddingE2E -v` passa

---

### T40: Testes e2e — endpoints de Guests

**What:** Testes de integração do módulo de convidados
**Where:** `internal/handler/guest_test.go`
**Depends on:** T36, T38
**Requirement:** GUEST-01..GUEST-06

**Cenários:**
- CRUD completo de convidados
- Filtro por status funciona
- Summary retorna contagens corretas
- RSVP público atualiza status

**Done when:**
- [ ] `go test ./internal/handler/ -run TestGuestE2E -v` passa

---

### T41: Testes e2e — endpoints de Gifts

**What:** Testes de integração do módulo de presentes
**Where:** `internal/handler/gift_test.go`
**Depends on:** T37, T38
**Requirement:** GIFT-01..GIFT-07

**Cenários:**
- CRUD completo de presentes
- Reserva pública → 200
- Reserva dupla → 409
- Cancelar reserva → 200 com status `available`
- Summary correto

**Done when:**
- [ ] `go test ./internal/handler/ -run TestGiftE2E -v` passa

---

### T42: Testes e2e — endpoints Públicos

**What:** Testes de integração da página pública
**Where:** `internal/handler/public_test.go`
**Depends on:** T38, T39, T40, T41
**Requirement:** PUB-01..PUB-04

**Cenários:**
- `GET /public/:slug` sem auth → 200 com dados do casamento
- `GET /public/:slug` slug inexistente → 404
- `GET /public/:slug/gifts` → `reserved_by_name` ausente no response
- `GET /public/:slug` → email do casal ausente no response
- Fluxo completo: criar casamento → acessar página pública → convidado confirma → casal vê status

**Done when:**
- [ ] `go test ./internal/handler/ -run TestPublicE2E -v` passa
- [ ] Cobertura geral do projeto ≥ 80% (`go test ./... -cover`)

---

## Mapa de Execução Paralela

```
Phase 1 (Sequential):
  T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9 → T10

Phase 2 (Parallel):
  T10 ──┬── T11 ──┐
        ├── T12 ──┼── T14
        └── T13 ──┘

Phase 3 (Sequential — Auth):
  T14 → T15 → T16 → T17 → T18

Phase 4 (Sequential — Wedding):
  T18 → T19 → T20 → T21 → T22

Phase 5 (Parallel — Guests + Gifts):
  T22 ──┬── T23 → T24 → T25
        └── T26 → T27 → T28

Phase 6 (Sequential):
  T25 + T28 → T29 → T30

Phase 7 (Sequential — Docs):
  T30 → T31 → T32 → T33

Phase 8 (Parallel — Unit Tests):
  T30 ──┬── T34
        ├── T35
        ├── T36
        └── T37

Phase 9 (Sequential — E2E Tests):
  T34+T35+T36+T37 → T38 → T39 → T40 → T41 → T42
```

---

## Granularity Check

| Task | Escopo | Status |
|------|--------|--------|
| T1: go.mod | 1 arquivo de config | ✅ |
| T4: config.go | 1 arquivo, 1 struct | ✅ |
| T5: errors.go | 1 arquivo, vars tipadas | ✅ |
| T6: structs domínio | 4 arquivos coesos | ✅ |
| T9: migrations | 14 arquivos SQL | ✅ |
| T17: AuthService | 1 service, 4 métodos | ✅ |
| T18: AuthHandler | 1 handler, 4 endpoints | ✅ |
| T26: GiftRepository | 1 repo com lógica crítica | ✅ |
| T30: main.go | 1 arquivo de composição | ✅ |
