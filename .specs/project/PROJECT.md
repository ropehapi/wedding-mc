# wedding-mc

**Vision:** SaaS para gestão de casamentos que oferece ao casal uma landing page personalizada e ferramentas para gerenciar lista de convidados e lista de presentes, com acesso simplificado para convidados sem necessidade de autenticação.

**For:** Casais que estão se casando e desejam organizar seu evento de forma digital e centralizada.

**Solves:** A fragmentação das ferramentas de organização de casamento — planilhas para convidados, listas de presentes em sites separados, ausência de uma página própria do evento. O wedding-mc centraliza tudo em uma única plataforma simples.

---

## Goals

- Oferecer ao casal uma página pública do casamento com fotos, informações e links em menos de 30 minutos de configuração
- Permitir que 100% dos convidados confirmem presença sem precisar criar conta ou logar
- Permitir que convidados reservem presentes com zero fricção (sem cadastro, sem pagamento na plataforma)
- API REST bem estruturada, pronta para ser consumida por um frontend externo

---

## Tech Stack

**Core:**

- Language: Go (versão a definir)
- Framework HTTP: a definir
- Banco de dados: a definir
- ORM/Query builder: a definir
- Autenticação: a definir (JWT próprio ou terceiro)
- Storage de arquivos: Amazon S3 ou sistema de arquivos local (configurável)

**Key dependencies:** a definir após decisão de stack

---

## Actors

| Ator       | Descrição                                                                 | Autenticação |
|------------|---------------------------------------------------------------------------|--------------|
| Casal      | Administrador do casamento. Gerencia convidados, presentes e a página.   | Sim (JWT)    |
| Convidado  | Acessa a página pública do casamento. Confirma presença e reserva presentes. | Não         |

---

## Módulos

| Módulo         | Descrição                                                                 | MVP |
|----------------|---------------------------------------------------------------------------|-----|
| Auth           | Registro e login do casal                                                | P1  |
| Casamento      | Perfil do casamento (data, local, história, fotos, links)                | P1  |
| Convidados     | Gestão da lista de convidados e rastreamento de confirmações             | P1  |
| Presentes      | Lista de presentes com links externos e rastreamento de reservas         | P1  |
| Página Pública | Landing page do casamento acessível por convidados sem autenticação      | P1  |

---

## Scope

**v1 inclui:**

- Registro e autenticação do casal
- CRUD completo do perfil do casamento
- CRUD da lista de convidados com status de confirmação (vai / não vai)
- CRUD da lista de presentes com links externos e status de reserva
- Endpoint público da página do casamento (leitura)
- Endpoint público de confirmação de presença (sem auth)
- Endpoint público de reserva de presente (sem auth)
- Upload de fotos do casal (S3 ou local)

**Explicitamente fora do escopo (v1):**

| Feature                          | Razão                                 |
|----------------------------------|---------------------------------------|
| Pagamento na plataforma          | Complexidade — planejado para v2      |
| Contribuição financeira (vaquinha) | Planejado para v2                   |
| Múltiplos casamentos por casal   | Um casal, um casamento                |
| Notificações automáticas (email/WhatsApp) | Planejado para v2            |
| Painel administrativo do SaaS    | Sem necessidade imediata (gratuito)   |
| Frontend                         | Será desenvolvido em outro repositório|

---

## Constraints

- **Técnico:** API REST pura, sem SSR — o frontend é externo
- **Modelo de negócio:** Gratuito em v1, monetização planejada para versões futuras
- **Simplicidade:** Convidados acessam sem criar conta. Nenhuma fricção desnecessária.
- **Storage:** Fotos armazenadas no S3 ou localmente — configurável por variável de ambiente
