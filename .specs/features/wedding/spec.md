# Casamento — Especificação

## Problem Statement

Cada casal precisa configurar o perfil do seu casamento — informações do evento, fotos e links — que servirão tanto para o painel de gestão quanto para a página pública acessada pelos convidados.

## Goals

- [ ] Casal consegue criar e editar todas as informações do seu casamento
- [ ] Upload de fotos com armazenamento configurável (S3 ou local)
- [ ] Perfil do casamento associado de forma única à conta do casal

## Out of Scope

| Feature                        | Razão                                       |
|-------------------------------|---------------------------------------------|
| Múltiplos casamentos por conta | Um casal = um casamento no v1              |
| Temas e personalização visual  | Responsabilidade do frontend               |
| Vídeos                        | Apenas fotos em v1                         |

---

## User Stories

### P1: Criar perfil do casamento ⭐ MVP

**User Story:** Como casal, quero criar o perfil do meu casamento com nome, data, local e uma descrição para que os convidados saibam os detalhes do evento.

**Why P1:** É o ponto de partida de toda a plataforma.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `POST /v1/wedding` com dados válidos THEN sistema SHALL criar o casamento e retornar `201 Created`
2. WHEN casal já possui um casamento cadastrado THEN sistema SHALL retornar `409 Conflict`
3. WHEN campos obrigatórios (`bride_name`, `groom_name`, `date`, `location`) estão ausentes THEN sistema SHALL retornar `422 Unprocessable Entity`
4. WHEN casamento criado THEN sistema SHALL associá-lo ao casal autenticado (1:1)

**Campos do perfil:**
- `bride_name` (string, obrigatório)
- `groom_name` (string, obrigatório)
- `date` (date, obrigatório) — data do casamento
- `time` (time, opcional) — horário da cerimônia
- `location` (string, obrigatório) — nome do local/endereço
- `city` (string, opcional)
- `state` (string, opcional)
- `description` (text, opcional) — história do casal ou mensagem
- `links` (array de objetos, opcional) — links externos com `label` e `url`

**Independent Test:** `POST /v1/wedding` autenticado → `201`; segundo `POST` → `409`.

---

### P1: Editar perfil do casamento ⭐ MVP

**User Story:** Como casal, quero editar as informações do meu casamento a qualquer momento.

**Why P1:** Informações mudam durante o planejamento.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `PATCH /v1/wedding` com campos válidos THEN sistema SHALL atualizar e retornar `200 OK` com dados atualizados
2. WHEN casal não possui casamento cadastrado THEN sistema SHALL retornar `404 Not Found`
3. WHEN campos enviados violam validação THEN sistema SHALL retornar `422 Unprocessable Entity`
4. WHEN outro casal tenta editar o casamento alheio THEN sistema SHALL retornar `403 Forbidden`

**Independent Test:** `PATCH /v1/wedding` com novo `location` → `200` com valor atualizado.

---

### P1: Consultar perfil do casamento ⭐ MVP

**User Story:** Como casal, quero visualizar os dados do meu casamento para conferir o que está configurado.

**Why P1:** Base do painel de administração.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `GET /v1/wedding` THEN sistema SHALL retornar `200 OK` com todos os dados do casamento incluindo URLs das fotos
2. WHEN casal não possui casamento cadastrado THEN sistema SHALL retornar `404 Not Found`

**Independent Test:** `GET /v1/wedding` autenticado → `200` com perfil completo.

---

### P1: Upload de fotos ⭐ MVP

**User Story:** Como casal, quero fazer upload de fotos para personalizar a página do nosso casamento.

**Why P1:** A landing page sem fotos não tem apelo.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `POST /v1/wedding/photos` com arquivo de imagem válido (JPEG, PNG, WebP) THEN sistema SHALL armazenar o arquivo (S3 ou local) e retornar `201 Created` com a URL pública da foto
2. WHEN arquivo não é uma imagem ou excede o limite de tamanho (ex: 10MB) THEN sistema SHALL retornar `422 Unprocessable Entity`
3. WHEN foto adicionada THEN sistema SHALL disponibilizá-la na listagem de fotos do casamento
4. WHEN configurado para S3 THEN sistema SHALL armazenar no bucket configurado; WHEN configurado para local THEN sistema SHALL armazenar no sistema de arquivos do servidor

**Independent Test:** Upload de JPEG → `201` com URL; GET no perfil → foto aparece na lista.

---

### P2: Remover foto

**User Story:** Como casal, quero remover fotos que não quero mais exibir.

**Why P2:** Gestão básica do conteúdo.

**Acceptance Criteria:**

1. WHEN casal autenticado envia `DELETE /v1/wedding/photos/:photo_id` THEN sistema SHALL remover o arquivo do storage e retornar `204 No Content`
2. WHEN foto não existe ou não pertence ao casamento THEN sistema SHALL retornar `404 Not Found`

**Independent Test:** Upload → `DELETE` → GET no perfil → foto não aparece mais.

---

### P2: Gerenciar links externos

**User Story:** Como casal, quero adicionar links externos (ex: link para o buffet, Instagram do fotógrafo) na nossa página.

**Why P2:** Enriquece a landing page sem complexidade adicional no backend.

**Acceptance Criteria:**

1. WHEN casal atualiza `links` via `PATCH /v1/wedding` com array de `{label, url}` THEN sistema SHALL persistir e retornar os links atualizados
2. WHEN `url` inválida é enviada THEN sistema SHALL retornar `422 Unprocessable Entity`

**Independent Test:** `PATCH` com `links: [{label: "Buffet", url: "https://..."}]` → GET retorna os links.

---

## Edge Cases

- WHEN data do casamento é no passado THEN sistema SHALL aceitar (casal pode estar cadastrando após o evento) mas logar um warning
- WHEN `links` array vazio é enviado THEN sistema SHALL limpar os links existentes
- WHEN upload simultâneo de múltiplas fotos THEN sistema SHALL processar cada uma independentemente

---

## Requirement Traceability

| Requirement ID | Story                      | Status  |
|----------------|----------------------------|---------|
| WED-01         | P1: Criar perfil           | Pending |
| WED-02         | P1: Editar perfil          | Pending |
| WED-03         | P1: Consultar perfil       | Pending |
| WED-04         | P1: Upload de fotos        | Pending |
| WED-05         | P2: Remover foto           | Pending |
| WED-06         | P2: Gerenciar links        | Pending |

---

## Success Criteria

- [ ] Casal consegue configurar o perfil completo em < 10 minutos
- [ ] Fotos disponíveis publicamente via URL após upload
- [ ] Nenhum dado de outro casal é acessível
