# Página Pública — Especificação

## Problem Statement

Cada casamento precisa de uma URL pública, acessível por qualquer pessoa sem autenticação, que consolide as informações do evento, a lista de convidados para confirmação e a lista de presentes para reserva.

## Goals

- [ ] Cada casamento tem uma URL única e amigável (slug)
- [ ] Convidado acessa todas as informações e realiza as ações necessárias em uma única página
- [ ] Nenhuma autenticação necessária para qualquer ação pública

## Out of Scope

| Feature                         | Razão                                      |
|---------------------------------|--------------------------------------------|
| Renderização HTML (SSR)         | API REST pura — frontend é responsabilidade externa |
| Senha de acesso à página        | Fora do escopo v1                         |
| Analytics de visitas            | Planejado para v2                         |
| Mural de mensagens              | Fora do escopo v1                         |

---

## User Stories

### P1: Slug único por casamento ⭐ MVP

**User Story:** Como casal, quero que meu casamento tenha uma URL amigável e única para compartilhar com os convidados.

**Why P1:** Sem slug, não há como identificar o casamento publicamente.

**Acceptance Criteria:**

1. WHEN casal cria seu casamento THEN sistema SHALL gerar automaticamente um `slug` baseado nos nomes do casal (ex: `ana-e-joao`)
2. WHEN slug gerado já existe THEN sistema SHALL acrescentar um sufixo numérico (ex: `ana-e-joao-2`)
3. WHEN casal envia `PATCH /v1/wedding` com `slug` customizado THEN sistema SHALL validar que é único e contém apenas letras minúsculas, números e hífens
4. WHEN slug inválido é enviado THEN sistema SHALL retornar `422 Unprocessable Entity`
5. WHEN slug duplicado é enviado THEN sistema SHALL retornar `409 Conflict`

**Independent Test:** Criar casamento → slug gerado automaticamente → acessar `GET /v1/public/:slug` → retorna dados do casamento.

---

### P1: Página pública do casamento ⭐ MVP

**User Story:** Como convidado, quero acessar uma página com todas as informações do casamento, sem precisar criar conta.

**Why P1:** É o ponto de entrada de todos os convidados.

**Acceptance Criteria:**

1. WHEN qualquer pessoa envia `GET /v1/public/:wedding_slug` THEN sistema SHALL retornar `200 OK` com dados públicos do casamento
2. WHEN `wedding_slug` não existe THEN sistema SHALL retornar `404 Not Found`
3. WHEN dados retornados THEN sistema SHALL incluir: nomes do casal, data, horário, local, cidade, estado, descrição, fotos (URLs), links externos
4. WHEN dados retornados THEN sistema SHALL NÃO incluir: informações de conta do casal, emails, dados sensíveis

**Payload público do casamento:**
```json
{
  "slug": "ana-e-joao",
  "bride_name": "Ana",
  "groom_name": "João",
  "date": "2026-11-15",
  "time": "17:00",
  "location": "Espaço Villa Rica",
  "city": "São Paulo",
  "state": "SP",
  "description": "...",
  "photos": ["https://..."],
  "links": [{"label": "Buffet", "url": "https://..."}]
}
```

**Independent Test:** Casal configura perfil → `GET /v1/public/ana-e-joao` (sem auth) → retorna dados corretos.

---

### P1: Lista pública de convidados ⭐ MVP

**User Story:** Como convidado, quero ver a lista de nomes disponíveis para selecionar o meu e confirmar presença.

**Why P1:** É o mecanismo de identificação do convidado no RSVP.

**Acceptance Criteria:**

1. WHEN qualquer pessoa envia `GET /v1/public/:wedding_slug/guests` THEN sistema SHALL retornar `200 OK` com lista de `{id, name, status}` de todos os convidados
2. WHEN casamento não existe THEN sistema SHALL retornar `404 Not Found`
3. WHEN lista está vazia THEN sistema SHALL retornar `200 OK` com array vazio

**Nota:** O frontend usará esta lista para montar o `<select>` de seleção de nome.

**Independent Test:** Casal cadastra 5 convidados → `GET /v1/public/:slug/guests` (sem auth) → retorna 5 nomes com IDs.

---

### P1: Lista pública de presentes ⭐ MVP

**User Story:** Como convidado, quero ver a lista de presentes com status de disponibilidade para escolher um para reservar.

**Why P1:** É a vitrine do módulo de presentes para os convidados.

**Acceptance Criteria:**

1. WHEN qualquer pessoa envia `GET /v1/public/:wedding_slug/gifts` THEN sistema SHALL retornar `200 OK` com todos os presentes
2. WHEN presente está `reserved` THEN sistema SHALL incluir `reserved: true` mas NÃO incluir `reserved_by_name` (privacidade)
3. WHEN presente está `available` THEN sistema SHALL incluir `reserved: false`
4. WHEN casamento não existe THEN sistema SHALL retornar `404 Not Found`

**Payload público de presente:**
```json
{
  "id": "uuid",
  "name": "Jogo de panelas",
  "description": "...",
  "image_url": "https://...",
  "store_url": "https://...",
  "price": 350.00,
  "reserved": false
}
```

**Independent Test:** Casal cadastra 5 presentes, convidado reserva 1 → `GET /v1/public/:slug/gifts` → 1 aparece com `reserved: true`, sem nome do reservante.

---

## Edge Cases

- WHEN slug contém caracteres especiais na URL THEN sistema SHALL aceitar apenas slugs com `[a-z0-9-]`
- WHEN casamento existe mas nenhuma foto foi cadastrada THEN sistema SHALL retornar `photos: []`
- WHEN casamento existe mas lista de presentes está vazia THEN sistema SHALL retornar array vazio (sem erro)

---

## Requirement Traceability

| Requirement ID | Story                             | Status  |
|----------------|-----------------------------------|---------|
| PUB-01         | P1: Slug único                    | Pending |
| PUB-02         | P1: Página pública do casamento   | Pending |
| PUB-03         | P1: Lista pública de convidados   | Pending |
| PUB-04         | P1: Lista pública de presentes    | Pending |

---

## Success Criteria

- [ ] URL pública funciona sem qualquer autenticação
- [ ] Dados sensíveis do casal nunca expostos nos endpoints públicos
- [ ] Slug único garantido por constraint no banco de dados
