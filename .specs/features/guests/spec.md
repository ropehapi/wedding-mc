# Convidados — Especificação

## Problem Statement

O casal precisa gerenciar quem está convidado para o casamento e acompanhar as confirmações de presença. Os convidados precisam confirmar presença de forma simples, sem autenticação, apenas selecionando seu nome em uma lista.

## Goals

- [ ] Casal consegue cadastrar, editar e remover convidados
- [ ] Casal acompanha em tempo real quem confirmou e quem não confirmou
- [ ] Convidado confirma presença em < 3 cliques, sem criar conta

## Out of Scope

| Feature                          | Razão                                      |
|----------------------------------|--------------------------------------------|
| Opções de cardápio               | Não solicitado                             |
| Acompanhantes (plus-one)         | Fora do escopo v1                         |
| Restrições alimentares           | Fora do escopo v1                         |
| Importação de lista via CSV      | Planejado para v2                         |
| Envio automático de convites     | O convite é feito pessoalmente pelo casal |
| Notificação quando convidado confirma | Planejado para v2                    |

---

## User Stories

### P1: Adicionar convidado ⭐ MVP

**User Story:** Como casal, quero adicionar convidados à minha lista para que possam confirmar presença.

**Why P1:** Base do módulo de gestão de convidados.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `POST /v1/guests` com `name` válido THEN sistema SHALL criar o convidado com status `pending` e retornar `201 Created`
2. WHEN `name` está vazio ou ausente THEN sistema SHALL retornar `422 Unprocessable Entity`
3. WHEN convidado criado THEN sistema SHALL associá-lo ao casamento do casal autenticado
4. WHEN mesmo nome é cadastrado duas vezes THEN sistema SHALL permitir (pode haver homônimos) mas retornar aviso no response

**Campos do convidado:**
- `name` (string, obrigatório) — nome completo
- `status` (enum, gerenciado pelo sistema): `pending` | `confirmed` | `declined`

**Independent Test:** `POST /v1/guests` com `name` → `201`; `GET /v1/guests` → convidado aparece na lista.

---

### P1: Listar convidados ⭐ MVP

**User Story:** Como casal, quero ver todos os meus convidados e seus status de confirmação.

**Why P1:** Painel de controle central do módulo.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `GET /v1/guests` THEN sistema SHALL retornar `200 OK` com lista de todos os convidados do casamento com `id`, `name` e `status`
2. WHEN filtro `?status=confirmed` é enviado THEN sistema SHALL retornar apenas convidados com aquele status
3. WHEN lista está vazia THEN sistema SHALL retornar `200 OK` com array vazio
4. WHEN casal não possui casamento cadastrado THEN sistema SHALL retornar `404 Not Found`

**Status possíveis:** `pending`, `confirmed`, `declined`

**Independent Test:** Adicionar 3 convidados → `GET /v1/guests` → retorna os 3 com status `pending`.

---

### P1: Editar convidado ⭐ MVP

**User Story:** Como casal, quero corrigir o nome de um convidado.

**Why P1:** Erros de digitação são inevitáveis.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `PATCH /v1/guests/:guest_id` com `name` válido THEN sistema SHALL atualizar e retornar `200 OK`
2. WHEN convidado não existe ou não pertence ao casamento THEN sistema SHALL retornar `404 Not Found`

**Independent Test:** Adicionar convidado → `PATCH` nome → `GET` → nome atualizado.

---

### P1: Remover convidado ⭐ MVP

**User Story:** Como casal, quero remover um convidado da lista.

**Why P1:** A lista muda durante o planejamento.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `DELETE /v1/guests/:guest_id` THEN sistema SHALL remover e retornar `204 No Content`
2. WHEN convidado não existe THEN sistema SHALL retornar `404 Not Found`
3. WHEN convidado já havia confirmado presença THEN sistema SHALL remover mesmo assim (sem bloqueio)

**Independent Test:** Adicionar → `DELETE` → `GET /v1/guests` → convidado não aparece.

---

### P1: Confirmação de presença pelo convidado ⭐ MVP

**User Story:** Como convidado, quero confirmar ou recusar minha presença no casamento sem precisar criar uma conta.

**Why P1:** É a funcionalidade central do módulo para os convidados.

**Acceptance Criteria:**

1. WHEN convidado envia `GET /v1/public/:wedding_slug/guests` THEN sistema SHALL retornar lista pública com `id` e `name` de todos os convidados com status `pending` ou `confirmed` (para o select)
2. WHEN convidado envia `POST /v1/public/:wedding_slug/guests/:guest_id/rsvp` com `status: "confirmed"` ou `status: "declined"` THEN sistema SHALL atualizar o status e retornar `200 OK`
3. WHEN convidado tenta confirmar por um guest_id que não pertence àquele casamento THEN sistema SHALL retornar `404 Not Found`
4. WHEN convidado já havia confirmado e envia novamente THEN sistema SHALL aceitar a atualização (permite mudar de "confirmed" para "declined" e vice-versa)
5. WHEN `status` enviado é inválido (nem `confirmed` nem `declined`) THEN sistema SHALL retornar `422 Unprocessable Entity`

**Independent Test:** Casal cadastra convidado → `GET /public/:slug/guests` → nome aparece → `POST /rsvp` com `confirmed` → status atualizado no painel do casal.

---

### P2: Resumo de confirmações

**User Story:** Como casal, quero ver um resumo rápido de quantos confirmaram, recusaram e ainda não responderam.

**Why P2:** Facilita o acompanhamento sem precisar contar manualmente.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `GET /v1/guests/summary` THEN sistema SHALL retornar `{ total, confirmed, declined, pending }`

**Independent Test:** 10 convidados, 3 confirmados, 2 recusados → summary retorna `{total: 10, confirmed: 3, declined: 2, pending: 5}`.

---

## Edge Cases

- WHEN convidado tenta confirmar presença por nome (sem ID) THEN sistema SHALL rejeitar — a confirmação é sempre por ID para evitar conflitos de homônimos
- WHEN lista de convidados está vazia e convidado acessa a página pública THEN sistema SHALL retornar array vazio (sem erro)
- WHEN `wedding_slug` não existe THEN sistema SHALL retornar `404 Not Found`

---

## Requirement Traceability

| Requirement ID | Story                          | Status  |
|----------------|-------------------------------|---------|
| GUEST-01       | P1: Adicionar convidado       | Pending |
| GUEST-02       | P1: Listar convidados         | Pending |
| GUEST-03       | P1: Editar convidado          | Pending |
| GUEST-04       | P1: Remover convidado         | Pending |
| GUEST-05       | P1: RSVP pelo convidado       | Pending |
| GUEST-06       | P2: Resumo de confirmações    | Pending |

---

## Success Criteria

- [ ] Convidado consegue confirmar presença em menos de 3 cliques
- [ ] Casal vê status atualizado em tempo real após confirmação
- [ ] Nenhuma autenticação exigida para o fluxo público de RSVP
