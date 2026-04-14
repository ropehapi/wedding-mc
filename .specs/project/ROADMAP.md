# Roadmap — wedding-mc

## v1 — MVP (escopo atual)

**Objetivo:** API funcional com todos os módulos core. Pronta para consumo por um frontend externo.

### Milestone 1: Fundação

- [ ] Definição da stack técnica (framework, banco, ORM, auth)
- [ ] Setup do projeto Go (estrutura de pastas, configuração, CI)
- [ ] Banco de dados: schema inicial + migrations
- [ ] Módulo de Auth: registro e login do casal (JWT)

### Milestone 2: Core do Casamento

- [ ] Módulo de Casamento: CRUD do perfil (data, local, história, links)
- [ ] Upload de fotos (S3 ou local — configurável)
- [ ] Módulo de Convidados: CRUD da lista de convidados

### Milestone 3: Interação dos Convidados

- [ ] Módulo de Presentes: CRUD da lista de presentes com links externos
- [ ] Endpoint público da página do casamento
- [ ] Endpoint público de confirmação de presença (sem auth)
- [ ] Endpoint público de reserva de presente (sem auth)

### Milestone 4: Qualidade e Entrega

- [ ] Testes de integração dos endpoints críticos
- [ ] Documentação da API (Swagger/OpenAPI)
- [ ] Deploy inicial (ambiente de produção)

---

## v2 — Evolução (planejado, sem data)

- Contribuição financeira para presentes (Pix / cartão)
- Notificações automáticas por email/WhatsApp (confirmação, reserva de presente)
- Dashboard de métricas para o casal (stats de confirmação, presentes reservados)
- Personalização avançada da landing page (temas, cores)
- Monetização (planos pagos, recursos premium)

---

## v3 — Expansão (ideia, sem compromisso)

- Múltiplos eventos por conta (casais que organizam mais de um evento)
- Módulo de fornecedores (fotógrafo, buffet, etc.)
- Módulo de cronograma do evento
- Módulo financeiro (orçamento vs. gastos)
- Painel administrativo do SaaS
