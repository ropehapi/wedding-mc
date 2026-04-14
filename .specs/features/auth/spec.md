# Auth — Especificação

## Problem Statement

O casal precisa de uma conta para acessar o painel de gestão do casamento. Apenas o casal autenticado pode criar e editar conteúdo. Nenhuma outra entidade (convidados) precisa de autenticação.

## Goals

- [ ] Casal consegue se registrar e fazer login de forma segura
- [ ] Rotas protegidas rejeitam requisições sem token válido
- [ ] Token com expiração configurável

## Out of Scope

| Feature                        | Razão                                         |
|-------------------------------|-----------------------------------------------|
| Login social (Google, Facebook) | Simplicidade — pode ser adicionado depois    |
| Recuperação de senha           | Planejado para v2                            |
| Multi-fator (MFA)             | Fora do escopo v1                            |
| Múltiplos usuários por casamento | Um casal, uma conta                        |

---

## User Stories

### P1: Registro do casal ⭐ MVP

**User Story:** Como casal, quero criar uma conta com email e senha para acessar o painel do meu casamento.

**Why P1:** Sem autenticação, não há como proteger os dados do casamento.

**Acceptance Criteria:**

1. WHEN casal envia `POST /v1/auth/register` com `name`, `email` e `password` válidos THEN sistema SHALL criar a conta e retornar `201 Created` com os dados do usuário (sem senha)
2. WHEN email já está cadastrado THEN sistema SHALL retornar `409 Conflict` com mensagem descritiva
3. WHEN `email` inválido ou `password` com menos de 8 caracteres THEN sistema SHALL retornar `422 Unprocessable Entity` com lista de erros de validação
4. WHEN registro bem-sucedido THEN sistema SHALL armazenar a senha com hash (bcrypt ou argon2)

**Independent Test:** `POST /v1/auth/register` → `201` com dados do casal; mesmo email → `409`.

---

### P1: Login do casal ⭐ MVP

**User Story:** Como casal, quero fazer login com email e senha para receber um token de acesso.

**Why P1:** Necessário para acessar qualquer rota protegida.

**Acceptance Criteria:**

1. WHEN casal envia `POST /v1/auth/login` com `email` e `password` corretos THEN sistema SHALL retornar `200 OK` com `access_token` (JWT) e `expires_at`
2. WHEN senha incorreta ou email não encontrado THEN sistema SHALL retornar `401 Unauthorized` (mesma mensagem para ambos, sem revelar qual está errado)
3. WHEN token é enviado em rotas protegidas via header `Authorization: Bearer <token>` THEN sistema SHALL validar e autorizar o acesso
4. WHEN token expirado ou inválido THEN sistema SHALL retornar `401 Unauthorized`

**Independent Test:** Login com credenciais válidas → token JWT; usar token em rota protegida → `200`.

---

### P2: Refresh de token

**User Story:** Como casal, quero renovar meu token sem precisar fazer login novamente.

**Why P2:** UX — evita logout forçado após expiração do token.

**Acceptance Criteria:**

1. WHEN casal envia `POST /v1/auth/refresh` com refresh token válido THEN sistema SHALL retornar novo `access_token`
2. WHEN refresh token inválido ou expirado THEN sistema SHALL retornar `401 Unauthorized`

**Independent Test:** Token expirado + refresh válido → novo access token.

---

### P2: Logout

**User Story:** Como casal, quero encerrar minha sessão.

**Why P2:** Segurança básica de sessão.

**Acceptance Criteria:**

1. WHEN casal envia `POST /v1/auth/logout` com token válido THEN sistema SHALL invalidar o token (via blacklist ou revogação de refresh token)
2. WHEN token invalidado é reutilizado THEN sistema SHALL retornar `401 Unauthorized`

**Independent Test:** Logout → reutilizar token → `401`.

---

## Edge Cases

- WHEN body da requisição está vazio ou malformado THEN sistema SHALL retornar `400 Bad Request`
- WHEN campo `password` enviado na resposta THEN sistema SHALL nunca expô-lo (hash nunca volta na API)

---

## Requirement Traceability

| Requirement ID | Story                | Status  |
|----------------|----------------------|---------|
| AUTH-01        | P1: Registro         | Pending |
| AUTH-02        | P1: Login            | Pending |
| AUTH-03        | P1: Validação de JWT | Pending |
| AUTH-04        | P2: Refresh token    | Pending |
| AUTH-05        | P2: Logout           | Pending |

---

## Success Criteria

- [ ] Casal consegue se registrar e logar em < 5 segundos de latência
- [ ] Senhas nunca expostas em nenhum endpoint
- [ ] Rotas protegidas retornam `401` consistentemente sem token válido
