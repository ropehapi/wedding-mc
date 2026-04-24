# Getting Started — wedding-mc

## Pré-requisitos

- Go 1.21+
- Docker e Docker Compose

---

## 1. Configurar o ambiente

```bash
cp .env.example .env
```

Os valores padrão do `.env.example` funcionam para desenvolvimento local.

---

## 2. Subir o banco de dados

```bash
docker compose up -d
```

Verifique se o container está saudável:

```bash
docker compose ps
```

---

## 3. Subir a API

```bash
go run ./cmd/api
```

As migrations são aplicadas automaticamente na inicialização. Você verá:

```
server starting addr=:8080
```

---

## 4. Verificar que está funcionando

```bash
curl http://localhost:8080/health
# {"data":{"status":"ok"}}
```

---

## 5. Explorar via Swagger UI

Abra no browser:

```
http://localhost:8080/swagger/index.html
```

Todos os endpoints estão documentados com exemplos de request/response. É possível testar direto pela interface.

---

## 6. Fluxo completo de teste manual

### Registrar o casal

```bash
curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana e João","email":"ana@email.com","password":"senha123"}' | jq
```

### Login

```bash
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ana@email.com","password":"senha123"}' | jq
```

Guarde o `access_token` retornado:

```bash
TOKEN="cole_aqui_o_access_token"
```

### Criar casamento

```bash
curl -s -X POST http://localhost:8080/v1/wedding \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bride_name": "Ana",
    "groom_name": "João",
    "date": "2026-12-20",
    "location": "Buffet Royal",
    "city": "São Paulo",
    "state": "SP"
  }' | jq
```

O slug é gerado automaticamente (`ana-e-joao`).

### Adicionar convidado

```bash
curl -s -X POST http://localhost:8080/v1/guests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Carlos"}' | jq
```

### Adicionar presente

```bash
curl -s -X POST http://localhost:8080/v1/gifts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Panela Le Creuset",
    "store_url": "https://amazon.com.br",
    "price": 599.90
  }' | jq
```

### Página pública (sem autenticação)

```bash
curl -s http://localhost:8080/v1/public/ana-e-joao | jq
```

### RSVP de um convidado

Use o `id` retornado na criação do convidado:

```bash
GUEST_ID="id_do_convidado"

curl -s -X POST http://localhost:8080/v1/public/ana-e-joao/guests/$GUEST_ID/rsvp \
  -H "Content-Type: application/json" \
  -d '{"status":"confirmed"}' | jq
```

Status aceitos: `confirmed` ou `declined`.

### Reservar presente

Use o `id` retornado na criação do presente:

```bash
GIFT_ID="id_do_presente"

curl -s -X POST http://localhost:8080/v1/public/ana-e-joao/gifts/$GIFT_ID/reserve \
  -H "Content-Type: application/json" \
  -d '{"guest_name":"Carlos"}' | jq
```

---

## 7. Importar no Bruno

No Bruno: **Import Collection → OpenAPI v2** → selecione `docs/openapi.yaml`.
