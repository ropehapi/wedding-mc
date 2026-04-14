# Lista de Presentes — Especificação

## Problem Statement

O casal precisa divulgar uma lista de presentes com links externos para as lojas, e os convidados precisam reservar um presente de forma simples, sem autenticação, para evitar que o mesmo presente seja comprado duas vezes.

## Goals

- [ ] Casal gerencia lista de presentes com links externos
- [ ] Convidado reserva um presente com zero fricção (sem cadastro)
- [ ] Presentes reservados ficam visualmente indisponíveis para outros convidados

## Out of Scope

| Feature                          | Razão                                       |
|----------------------------------|---------------------------------------------|
| Pagamento na plataforma         | Planejado para v2                           |
| Contribuição financeira (vaquinha) | Planejado para v2                        |
| Integração com lojas (API)      | Links externos apenas em v1                |
| Múltiplos convidados por presente | Um presente = uma reserva                 |
| Notificação ao casal quando presente é reservado | Planejado para v2          |

---

## User Stories

### P1: Adicionar presente ⭐ MVP

**User Story:** Como casal, quero adicionar presentes à minha lista com nome, descrição e link externo para a loja.

**Why P1:** Base do módulo de presentes.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `POST /v1/gifts` com dados válidos THEN sistema SHALL criar o presente com status `available` e retornar `201 Created`
2. WHEN `name` está ausente THEN sistema SHALL retornar `422 Unprocessable Entity`
3. WHEN `url` é enviada e é inválida THEN sistema SHALL retornar `422 Unprocessable Entity`
4. WHEN presente criado THEN sistema SHALL associá-lo ao casamento do casal autenticado

**Campos do presente:**
- `name` (string, obrigatório) — nome do presente
- `description` (string, opcional) — descrição ou observação
- `image_url` (string, opcional) — URL de uma imagem do produto (externa)
- `store_url` (string, opcional) — link para a loja onde comprar
- `price` (decimal, opcional) — preço estimado (informativo)
- `status` (enum, gerenciado pelo sistema): `available` | `reserved`

**Independent Test:** `POST /v1/gifts` → `201`; `GET /v1/gifts` → presente aparece com status `available`.

---

### P1: Listar presentes (painel do casal) ⭐ MVP

**User Story:** Como casal, quero ver todos os presentes da lista com seus status.

**Why P1:** Painel de controle do módulo.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `GET /v1/gifts` THEN sistema SHALL retornar `200 OK` com todos os presentes incluindo `reserved_by_name` quando reservado
2. WHEN filtro `?status=available` ou `?status=reserved` é enviado THEN sistema SHALL filtrar
3. WHEN lista está vazia THEN sistema SHALL retornar `200 OK` com array vazio

**Independent Test:** Adicionar 3 presentes, reservar 1 → `GET /v1/gifts` retorna 3, com o reservado mostrando `reserved_by_name`.

---

### P1: Editar presente ⭐ MVP

**User Story:** Como casal, quero editar as informações de um presente.

**Why P1:** Preços e links mudam durante o planejamento.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `PATCH /v1/gifts/:gift_id` com campos válidos THEN sistema SHALL atualizar e retornar `200 OK`
2. WHEN presente não existe ou não pertence ao casamento THEN sistema SHALL retornar `404 Not Found`
3. WHEN presente está reservado e casal tenta editar THEN sistema SHALL permitir a edição (apenas o casal pode editar)

**Independent Test:** Adicionar → `PATCH` `store_url` → `GET` → URL atualizada.

---

### P1: Remover presente ⭐ MVP

**User Story:** Como casal, quero remover um presente da lista.

**Why P1:** A lista evolui durante o planejamento.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `DELETE /v1/gifts/:gift_id` THEN sistema SHALL remover e retornar `204 No Content`
2. WHEN presente não existe THEN sistema SHALL retornar `404 Not Found`
3. WHEN presente está reservado THEN sistema SHALL remover mesmo assim (sem bloqueio)

**Independent Test:** Adicionar → `DELETE` → `GET` → presente não aparece.

---

### P1: Reservar presente (convidado) ⭐ MVP

**User Story:** Como convidado, quero reservar um presente para garantir que ninguém mais compre o mesmo item.

**Why P1:** É a funcionalidade central do módulo para os convidados.

**Acceptance Criteria:**

1. WHEN convidado envia `POST /v1/public/:wedding_slug/gifts/:gift_id/reserve` com `guest_name` THEN sistema SHALL marcar o presente como `reserved`, armazenar `reserved_by_name` e retornar `200 OK`
2. WHEN presente já está `reserved` THEN sistema SHALL retornar `409 Conflict` com a mensagem "Este presente já foi reservado"
3. WHEN `guest_name` está vazio THEN sistema SHALL retornar `422 Unprocessable Entity`
4. WHEN `gift_id` não pertence àquele `wedding_slug` THEN sistema SHALL retornar `404 Not Found`

**Independent Test:** Convidado reserva presente → status muda para `reserved`; segundo convidado tenta reservar o mesmo → `409`.

---

### P2: Cancelar reserva (casal)

**User Story:** Como casal, quero cancelar a reserva de um presente (ex: convidado pediu para trocar).

**Why P2:** Casos excepcionais de gestão.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `DELETE /v1/gifts/:gift_id/reserve` THEN sistema SHALL limpar `reserved_by_name`, mudar status para `available` e retornar `200 OK`
2. WHEN presente não está reservado THEN sistema SHALL retornar `422 Unprocessable Entity`

**Independent Test:** Reservar → `DELETE /reserve` pelo casal → status volta para `available`.

---

### P2: Resumo da lista de presentes

**User Story:** Como casal, quero ver um resumo rápido de quantos presentes foram reservados.

**Why P2:** Visibilidade rápida do engajamento.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `GET /v1/gifts/summary` THEN sistema SHALL retornar `{ total, available, reserved }`

**Independent Test:** 10 presentes, 4 reservados → `{total: 10, available: 6, reserved: 4}`.

---

## Edge Cases

- WHEN convidado tenta reservar presente de outro casamento via `gift_id` válido mas `wedding_slug` errado THEN sistema SHALL retornar `404 Not Found`
- WHEN `price` negativo é enviado THEN sistema SHALL retornar `422 Unprocessable Entity`
- WHEN `image_url` ou `store_url` são URLs inválidas THEN sistema SHALL rejeitar na validação

---

## Requirement Traceability

| Requirement ID | Story                           | Status  |
|----------------|---------------------------------|---------|
| GIFT-01        | P1: Adicionar presente          | Pending |
| GIFT-02        | P1: Listar presentes (casal)    | Pending |
| GIFT-03        | P1: Editar presente             | Pending |
| GIFT-04        | P1: Remover presente            | Pending |
| GIFT-05        | P1: Reservar presente (público) | Pending |
| GIFT-06        | P2: Cancelar reserva (casal)    | Pending |
| GIFT-07        | P2: Resumo da lista             | Pending |

---

## Success Criteria

- [ ] Convidado reserva presente sem criar conta, com 2 ações (selecionar + confirmar)
- [ ] Conflito de reserva dupla nunca ocorre (409 garante atomicidade)
- [ ] Casal vê quem reservou cada presente
