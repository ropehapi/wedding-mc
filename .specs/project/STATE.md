# State — wedding-mc

**Última atualização:** 2026-04-13

---

## Decisões Tomadas

| Data       | Decisão                                                                              | Razão                                                       |
|------------|--------------------------------------------------------------------------------------|-------------------------------------------------------------|
| 2026-04-13 | Convidados não precisam de autenticação                                              | Máxima simplicidade de acesso                               |
| 2026-04-13 | Confirmação de presença via seleção de nome (sem input livre)                        | Evitar nomes duplicados/errados, manter integridade         |
| 2026-04-13 | Lista de presentes com links externos (sem pagamento na plataforma)                  | Simplicidade — evolução para pagamento em v2                |
| 2026-04-13 | Um casal = um casamento                                                              | Simplifica modelo de dados no MVP                           |
| 2026-04-13 | Fotos armazenadas no S3 ou localmente (configurável por env var)                     | Flexibilidade de ambiente (dev vs. prod)                    |
| 2026-04-13 | Frontend em repositório separado                                                     | Separação de responsabilidades                              |
| 2026-04-13 | Gratuito em v1                                                                       | Foco em produto antes de monetização                        |
| 2026-04-13 | Framework HTTP: **Chi**                                                              | Go idiomático, 100% compatível com net/http                 |
| 2026-04-13 | Banco de dados: **PostgreSQL**                                                       | Robusto, ACID, padrão para SaaS                             |
| 2026-04-13 | Query builder: **sqlx + golang-migrate**                                             | Controle total do SQL + migrations versionadas              |
| 2026-04-13 | Autenticação: **JWT próprio (golang-jwt/jwt)**                                       | Sem dependência externa, suficiente para o caso de uso      |
| 2026-04-13 | Versionamento de API: **/v1/ no path**                                               | Padrão REST amplamente adotado                              |
| 2026-04-13 | Injeção de dependência: **manual no main.go**                                        | Go idiomático, sem magic, fácil de testar                   |
| 2026-04-13 | Testes: **unit + e2e, cobertura mínima de 80%**                                      | Qualidade de software real                                  |
| 2026-04-13 | Documentação: **swaggo/swag (Swagger)** + **exportação OpenAPI para Bruno**          | Swagger gerado via anotações; Bruno para versionamento local|
| 2026-04-13 | CI/CD: **deixado para depois** (código primeiro)                                     | Foco no MVP                                                 |
| 2026-04-13 | Deploy: **cloud (GCP provável)** — decisão adiada                                   | Detalhes a definir após MVP                                 |

---

## Stack Técnica (Consolidada)

| Camada            | Tecnologia                        |
|-------------------|-----------------------------------|
| Linguagem         | Go (versão a definir — usar LTS)  |
| Framework HTTP    | Chi                               |
| Banco de dados    | PostgreSQL                        |
| Query builder     | sqlx                              |
| Migrations        | golang-migrate                    |
| Autenticação      | JWT (golang-jwt/jwt)              |
| Validação         | go-playground/validator           |
| Configuração      | godotenv + os.Getenv              |
| Logging           | zerolog                           |
| Documentação API  | swaggo/swag + OpenAPI export      |
| Storage (fotos)   | AWS S3 / local (configurável)     |
| Testes            | testing stdlib + testcontainers   |

---

## Arquitetura

**Padrão:** Arquitetura em camadas (handler → service → repository)

```
cmd/
  api/
    main.go              ← entrypoint, DI manual, inicialização
internal/
  handler/               ← recebe HTTP, valida, chama service, responde
  service/               ← regras de negócio, orquestra repositórios
  repository/            ← queries SQL via sqlx
  domain/                ← structs, enums, interfaces (contratos)
  middleware/            ← auth JWT, logger, recover, CORS
  config/                ← leitura de env vars
migrations/              ← arquivos .up.sql / .down.sql
docs/                    ← swagger gerado pelo swaggo
```

**Fluxo de uma requisição:**
```
HTTP Request → Middleware → Handler → Service → Repository → PostgreSQL
                                   ↓
                              HTTP Response
```

---

## Blockers

Nenhum no momento.

---

## Backlog Futuro (Para Implementar Depois)

### Infraestrutura
- [ ] **CI/CD com GitHub Actions** — lint, testes, build, deploy automático
- [ ] **Deploy na GCP** — definir serviço (Cloud Run é o mais indicado para containers stateless), configurar variáveis de ambiente, secrets no Secret Manager
- [ ] **Dockerfile de produção** (multi-stage build) + **docker-compose** para dev local
- [ ] **Health check endpoint** (`GET /health`) para load balancer / Cloud Run

### Features v2
- [ ] **Contribuição financeira para presentes** (Pix / cartão via gateway)
- [ ] **Notificações automáticas** — email/WhatsApp quando convidado confirma ou reserva presente
- [ ] **Recuperação de senha** (forgot password via email)
- [ ] **Importação de convidados via CSV**
- [ ] **Cancelamento de reserva pelo próprio convidado** (com validação por nome)
- [ ] **Dashboard de métricas** — confirmações por período, presentes mais vistos

### Features v3 (ideia, sem compromisso)
- [ ] Múltiplos eventos por conta
- [ ] Módulo de fornecedores
- [ ] Módulo de cronograma do evento
- [ ] Módulo financeiro (orçamento vs. gastos)
- [ ] Painel administrativo do SaaS
- [ ] Login social (Google)
- [ ] Planos pagos e monetização

---

## Ideias Adiadas (Deferred — sem data)

| Ideia                              | Contexto                                                  |
|------------------------------------|-----------------------------------------------------------|
| Múltiplos casamentos por conta     | Fora do escopo v1                                         |
| Módulo de fornecedores             | Mencionado como possibilidade futura                      |
| Módulo financeiro (orçamento)      | Mencionado como possibilidade futura                      |
| Cronograma do evento               | Mencionado como possibilidade futura                      |

---

## Preferências

- Respostas em português
- Usuário tem background em Go (backend), ainda aprendendo frontend
- Seguir boas práticas de mercado: arquitetura limpa em camadas, Go idiomático, sem over-engineering
