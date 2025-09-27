# zpwoot - Clean Architecture Implementation

## 📋 Visão Geral

Este documento descreve a implementação completa da arquitetura Clean Architecture no projeto zpwoot, uma API REST para gerenciamento de múltiplas sessões do WhatsApp usando Go, Fiber, PostgreSQL e whatsmeow.

## 🏗️ Estrutura Completa do Projeto

```
zpwoot/
├── .dockerignore                  # Configuração Docker
├── .env                          # Variáveis de ambiente (local)
├── .env.example                  # Template de variáveis de ambiente
├── .gitignore                    # Configuração Git
├── ARCHITECTURE.md               # 📖 Este documento
├── Dockerfile                    # Configuração de container
├── Makefile                      # Comandos de automação
├── README.md                     # Documentação principal
├── docker-compose.chatwoot.yml   # Compose para Chatwoot
├── docker-compose.dev.yml        # Compose para desenvolvimento
├── docker-compose.yml            # Compose principal
├── go.mod                        # Dependências Go
├── go.sum                        # Checksums das dependências
│
├── cmd/                          # 🚀 ENTRY POINTS
│   └── zpwoot/
│       └── main.go               # Ponto de entrada da aplicação
│
├── docs/                         # 📚 DOCUMENTAÇÃO
│   ├── API.md                    # Documentação da API
│   └── swagger/                  # Documentação Swagger gerada
│       ├── docs.go               # Código Swagger gerado
│       ├── swagger.json          # Especificação JSON
│       └── swagger.yaml          # Especificação YAML
│
├── internal/                     # 🏛️ CÓDIGO INTERNO DA APLICAÇÃO
│   ├── app/                      # 🎯 APPLICATION LAYER (Use Cases)
│   │   ├── container.go          # Dependency Injection Container
│   │   ├── chatwoot/             # Use cases de integração Chatwoot
│   │   │   ├── dto.go            # DTOs de Chatwoot
│   │   │   └── usecase.go        # Lógica de aplicação Chatwoot
│   │   ├── common/               # Use cases comuns
│   │   │   ├── dto.go            # DTOs comuns (responses, health)
│   │   │   └── usecase.go        # Health checks, estatísticas
│   │   ├── message/              # Use cases de mensagens
│   │   │   ├── dto.go            # DTOs de mensagens WhatsApp
│   │   │   └── usecase.go        # Lógica de envio de mensagens
│   │   ├── session/              # Use cases de sessões
│   │   │   ├── dto.go            # DTOs de sessões WhatsApp
│   │   │   └── usecase.go        # Lógica de gerenciamento de sessões
│   │   └── webhook/              # Use cases de webhooks
│   │       ├── dto.go            # DTOs de webhooks
│   │       └── usecase.go        # Lógica de webhooks
│   │
│   ├── domain/                   # 🏛️ DOMAIN LAYER (Entidades e Regras de Negócio)
│   │   ├── chatwoot/             # Domínio Chatwoot
│   │   │   ├── entity.go         # Entidades de Chatwoot
│   │   │   └── service.go        # Serviços de domínio Chatwoot
│   │   ├── message/              # Domínio de mensagens
│   │   │   ├── entity.go         # Entidades de mensagem
│   │   │   └── service.go        # Serviços de domínio de mensagem
│   │   ├── session/              # Domínio de sessões
│   │   │   ├── entity.go         # Entidades de sessão WhatsApp
│   │   │   └── service.go        # Serviços de domínio de sessão
│   │   └── webhook/              # Domínio de webhooks
│   │       ├── entity.go         # Entidades de webhook
│   │       └── service.go        # Serviços de domínio de webhook
│   │
│   ├── ports/                    # 🔌 INTERFACES (Contratos)
│   │   ├── chatwoot_repository.go # Interface repositório Chatwoot
│   │   ├── session_repository.go  # Interface repositório sessões
│   │   └── webhook_repository.go  # Interface repositório webhooks
│   │
│   └── infra/                    # 🔧 INFRASTRUCTURE LAYER (Implementações)
│       ├── db/                   # Banco de dados
│       │   ├── migrator.go       # Sistema de migrações
│       │   └── migrations/       # Scripts de migração
│       │       ├── 001_create_sessions_table.up.sql
│       │       ├── 001_create_sessions_table.down.sql
│       │       ├── 002_create_webhooks_table.up.sql
│       │       ├── 002_create_webhooks_table.down.sql
│       │       ├── 003_create_chatwoot_config_table.up.sql
│       │       └── 003_create_chatwoot_config_table.down.sql
│       │
│       ├── http/                 # Camada HTTP
│       │   ├── handlers/         # Handlers HTTP
│       │   │   ├── chatwoot.go   # Handler Chatwoot
│       │   │   ├── message.go    # Handler mensagens
│       │   │   ├── session.go    # Handler sessões
│       │   │   └── webhook.go    # Handler webhooks
│       │   ├── helpers/          # Utilitários HTTP
│       │   │   └── session_resolver.go # Resolução de sessões
│       │   ├── middleware/       # Middlewares
│       │   │   ├── auth.go       # Autenticação API Key
│       │   │   ├── logger.go     # Logging HTTP
│       │   │   ├── metrics.go    # Métricas
│       │   │   └── request_id.go # Request ID
│       │   └── routers/          # Roteamento
│       │       └── routes.go     # Configuração de rotas
│       │
│       ├── repository/           # 💾 IMPLEMENTAÇÕES DOS REPOSITÓRIOS
│       │   ├── repository.go     # Factory de repositórios
│       │   ├── chatwoot_repository.go # Repositório Chatwoot
│       │   ├── session_repository.go  # Repositório sessões
│       │   └── webhook_repository.go  # Repositório webhooks
│       │
│       └── wameow/              # 📱 INTEGRAÇÃO WHATSAPP
│           ├── README.md         # Documentação WhatsApp
│           ├── client.go         # Cliente WhatsApp
│           ├── connection.go     # Gerenciamento de conexões
│           ├── events.go         # Manipulação de eventos
│           ├── factory.go        # Factory de clientes
│           ├── manager.go        # Gerenciador de sessões
│           └── utils.go          # Utilitários WhatsApp
│
├── pkg/                          # 📦 PACOTES UTILITÁRIOS
│   ├── errors/                   # Sistema de erros
│   │   └── errors.go             # Definições de erro
│   └── uuid/                     # Geração de UUID
│       └── generator.go          # Gerador de UUID
│
└── platform/                    # 🛠️ PLATAFORMA E CONFIGURAÇÕES
    ├── config/                   # Configurações
    │   └── config.go             # Carregamento de configurações
    ├── db/                       # Abstração de banco
    │   └── db.go                 # Conexão e utilitários DB
    └── logger/                   # Sistema de logging
        ├── config.go             # Configuração de logs
        ├── logger.go             # Logger principal
        └── middleware.go         # Middleware de logging
```

## 🎯 Camada de Aplicação (Use Cases)

### **DTOs (Data Transfer Objects)**
- **Localização**: `internal/app/{domain}/dto.go`
- **Responsabilidade**: Contratos da API, validação e serialização
- **Domínios Implementados**:
  - **Common**: Responses padrão, health checks, estatísticas
  - **Session**: Criação, listagem, conexão de sessões WhatsApp
  - **Message**: Envio de mensagens (texto, mídia, documentos, etc.)
  - **Webhook**: Configuração e gerenciamento de webhooks
  - **Chatwoot**: Integração com plataforma de atendimento

### **Use Cases Implementados**
- **Localização**: `internal/app/{domain}/usecase.go`
- **Responsabilidade**: Orquestração da lógica de negócio
- **Padrões Aplicados**:
  - Interface + Implementação para testabilidade
  - Conversão entre DTOs e entidades de domínio
  - Coordenação entre repositórios e serviços de domínio
  - Tratamento de erros específicos por contexto

### **Container de Dependências**
- **Arquivo**: `internal/app/container.go`
- **Responsabilidade**: Dependency Injection e configuração
- **Funcionalidades**:
  - Criação e configuração de todos os use cases
  - Injeção de dependências (repositórios, serviços, logger)
  - Factory pattern para componentes complexos
  - Configuração centralizada de integrações externas

## 🏛️ Camada de Domínio

### **Entidades de Domínio**
- **Localização**: `internal/domain/{domain}/entity.go`
- **Características**:
  - Regras de negócio puras sem dependências externas
  - Validações de domínio específicas
  - Erros customizados por contexto
  - Métodos de comportamento das entidades

### **Domínios Implementados**:

#### **Session Domain**
- **Entidades**: Session, ProxyConfig, DeviceInfo, QRCodeResponse
- **Responsabilidades**: Gerenciamento de sessões WhatsApp, conexões, QR codes
- **Status**: Created, Connecting, Connected, Disconnected, Error, LoggedOut

#### **Message Domain**
- **Entidades**: Message, MediaMessage, ContactMessage, LocationMessage
- **Responsabilidades**: Estruturas de mensagens WhatsApp, validações de formato
- **Tipos**: Text, Image, Audio, Video, Document, Sticker, Location, Contact

#### **Webhook Domain**
- **Entidades**: WebhookConfig, WebhookEvent, WebhookDelivery
- **Responsabilidades**: Configuração de webhooks, eventos, estatísticas de entrega
- **Eventos**: Message, Connection, QR, PairSuccess, etc.

#### **Chatwoot Domain**
- **Entidades**: ChatwootConfig, ChatwootContact, ChatwootConversation
- **Responsabilidades**: Integração com Chatwoot, sincronização de dados
- **Funcionalidades**: Configuração de API, webhook bidirecional

### **Serviços de Domínio**
- **Localização**: `internal/domain/{domain}/service.go`
- **Responsabilidade**: Lógica de negócio complexa que não pertence a uma entidade específica
- **Exemplos**:
  - Validação de configurações de webhook
  - Processamento de eventos WhatsApp
  - Transformações de dados entre sistemas
  - Regras de negócio que envolvem múltiplas entidades

## 💾 Camada de Infraestrutura

### **Repositórios**
- **Localização**: `internal/infra/repository/`
- **Factory**: `repository.go` - Criação centralizada de repositórios
- **Implementações**:
  - `SessionRepository` - PostgreSQL com JSONB para proxy config
  - `WebhookRepository` - PostgreSQL com arrays JSONB para eventos
  - `ChatwootRepository` - PostgreSQL com relacionamentos

### **Características dos Repositórios**
- **Mapeamento OR**: Conversão entre modelos de banco e entidades de domínio
- **Queries Otimizadas**: Índices estratégicos para performance
- **Tratamento de Erros**: Conversão para erros específicos do domínio
- **Logging Estruturado**: Rastreamento detalhado de operações
- **Context Support**: Suporte a cancelamento e timeout
- **Transações**: Suporte a operações transacionais

### **Sistema de Migrações**
- **Localização**: `internal/infra/db/migrations/`
- **Migrator**: `migrator.go` - Sistema de controle de migrações
- **Tabelas Implementadas**:
  - `zpSessions` - Sessões WhatsApp com configurações
  - `zpWebhooks` - Configurações de webhook por sessão
  - `zpChatwoot` - Configurações de integração Chatwoot
- **Características**:
  - Migrações up/down para cada tabela
  - Índices otimizados para consultas frequentes
  - Triggers automáticos para updatedAt
  - Comentários de documentação nas tabelas

### **Integração WhatsApp (Wameow)**
- **Localização**: `internal/infra/wameow/`
- **Componentes**:
  - `manager.go` - Gerenciador principal de sessões
  - `client.go` - Cliente WhatsApp individual
  - `connection.go` - Gerenciamento de conexões
  - `events.go` - Manipulação de eventos WhatsApp
  - `factory.go` - Factory para criação de clientes
  - `utils.go` - Utilitários e helpers
- **Funcionalidades**:
  - Múltiplas sessões simultâneas
  - Eventos em tempo real (mensagens, conexão, QR)
  - Suporte a proxy HTTP/SOCKS5
  - Persistência automática de sessões

### **Camada HTTP**
- **Handlers**: `internal/infra/http/handlers/`
  - Processamento de requisições HTTP
  - Validação de entrada
  - Conversão entre DTOs e use cases
- **Middleware**: `internal/infra/http/middleware/`
  - Autenticação via API Key
  - Logging estruturado de requisições
  - Métricas de performance
  - Request ID para rastreamento
- **Helpers**: `internal/infra/http/helpers/`
  - Resolução de sessões por ID ou nome
  - Utilitários de validação

## 🔌 Interfaces (Ports)

### **Contratos de Repositório**
- **SessionRepository** - CRUD de sessões, busca por nome/ID/deviceJid
- **WebhookRepository** - CRUD de webhooks, estatísticas de entrega
- **ChatwootRepository** - CRUD de configurações, sincronização

### **Contratos de Integrações Externas**
- **WameowManager** - Gerenciamento de sessões WhatsApp
- **ChatwootIntegration** - API do Chatwoot (planejado)
- **EventHandler** - Manipulação de eventos WhatsApp

## 📦 Pacotes Utilitários (pkg/)

### **Sistema de Erros**
- **Localização**: `pkg/errors/errors.go`
- **Funcionalidades**:
  - Erros estruturados com códigos HTTP
  - Detalhes contextuais
  - Conversão automática para responses HTTP

### **Geração de UUID**
- **Localização**: `pkg/uuid/generator.go`
- **Funcionalidades**:
  - Geração de UUIDs v4
  - Validação de formato UUID

## 🛠️ Plataforma (platform/)

### **Configurações**
- **Localização**: `platform/config/config.go`
- **Funcionalidades**:
  - Carregamento de variáveis de ambiente
  - Configurações padrão
  - Validação de configurações obrigatórias

### **Banco de Dados**
- **Localização**: `platform/db/db.go`
- **Funcionalidades**:
  - Conexão com PostgreSQL via SQLx
  - Pool de conexões configurável
  - Suporte a transações
  - Health checks automáticos

### **Sistema de Logging**
- **Localização**: `platform/logger/`
- **Componentes**:
  - `logger.go` - Logger principal com zerolog
  - `config.go` - Configurações de logging
  - `middleware.go` - Middleware HTTP de logging
- **Funcionalidades**:
  - Logging estruturado em JSON
  - Diferentes níveis (trace, debug, info, warn, error, fatal)
  - Context-aware logging
  - Configuração por ambiente

## 📖 Documentação Swagger

### **Configuração**
- **Geração**: `make swagger` (usando swaggo/swag)
- **Servidor**: `make swagger-serve`
- **URL**: http://localhost:8080/swagger/
- **Arquivos Gerados**:
  - `docs/swagger/docs.go` - Código Go gerado
  - `docs/swagger/swagger.json` - Especificação JSON
  - `docs/swagger/swagger.yaml` - Especificação YAML

### **Estrutura da Documentação**
- **Comentários**: Anotações nos handlers HTTP
- **DTOs**: Documentados com exemplos e validações
- **Tags**: Organizados por domínio (Sessions, Messages, Webhooks, Chatwoot)
- **Autenticação**: Documentação de API Key
- **Responses**: Exemplos de sucesso e erro

## 🛠️ Comandos Disponíveis (Makefile)

### **Desenvolvimento**
```bash
# Desenvolvimento com hot reload (recomendado)
make dev

# Executar sem hot reload
make run

# Compilar projeto
go build ./...

# Gerar documentação Swagger
make swagger

# Servir com Swagger UI
make swagger-serve

# Iniciar serviços principais (PostgreSQL, Redis)
make up

# Iniciar Chatwoot completo
make up-cw

# Parar todos os serviços
make down
```

### **Banco de Dados**
```bash
# Executar migrações
make migrate-up

# Rollback última migração
make migrate-down

# Status das migrações
make migrate-status

# Criar nova migração
make migrate-create NAME=nome_da_migracao

# Seed do banco com dados de exemplo
make seed
```

### **Testes**
```bash
# Executar testes
go test ./...

# Testes com cobertura
make test-coverage

# Testar Swagger
make swagger-test

# Linting do código
make lint
```

### **Hot Reload (Air)**
```bash
# Instalar Air
make install-air

# Desenvolvimento com hot reload
make dev

# Limpar arquivos temporários do Air
make dev-clean

# Reinicializar configuração do Air
make dev-init
```

### **Docker**
```bash
# Build da imagem Docker
make docker-build

# Executar via Docker
make docker-run

# Logs dos containers
make logs
```

## 🔄 Fluxo de Dados e Arquitetura

### **Fluxo de Requisição HTTP**
```
HTTP Request → Middleware → Handler → DTO → Use Case → Domain Service → Repository → Database
     ↓            ↓          ↑        ↑       ↓           ↑             ↑          ↓
  Auth/Log    Validation  Response  Convert  Business   Domain      Database   Storage
                                              Logic     Entity      Model
```

### **Fluxo de Eventos WhatsApp**
```
WhatsApp → whatsmeow → Manager → Event Handler → Use Case → Webhook/Chatwoot
    ↓         ↓          ↓           ↓            ↓           ↓
  Message   Parse    Session    Process      Business    External
  Event     Event    Context    Event        Logic       Systems
```

### **Dependency Flow (Clean Architecture)**
```
┌─────────────────────────────────────────────────────────────┐
│                    External Systems                         │
│              (WhatsApp, Chatwoot, HTTP)                    │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│        (Handlers, Repositories, WhatsApp Manager)          │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│              (Use Cases, DTOs, Container)                   │
└─────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│            (Entities, Services, Business Rules)             │
└─────────────────────────────────────────────────────────────┘
```

## ✅ Benefícios da Implementação

### **1. Separação de Responsabilidades**
- **Camadas bem definidas**: Cada camada tem responsabilidades específicas
- **Baixo acoplamento**: Componentes independentes e intercambiáveis
- **Alta coesão**: Funcionalidades relacionadas agrupadas logicamente
- **Inversão de dependência**: Camadas internas não dependem de externas

### **2. Testabilidade**
- **Use cases isolados**: Testáveis independentemente da infraestrutura
- **Interfaces mockáveis**: Fácil criação de mocks para testes
- **Testes unitários**: Cada componente testável isoladamente
- **Testes de integração**: Validação de fluxos completos

### **3. Manutenibilidade**
- **Mudanças isoladas**: Alterações na API não afetam o domínio
- **Flexibilidade de banco**: Mudanças no PostgreSQL não afetam lógica
- **Evolução independente**: Cada camada pode evoluir separadamente
- **Refatoração segura**: Interfaces garantem compatibilidade

### **4. Escalabilidade**
- **Múltiplas sessões**: Suporte a centenas de sessões simultâneas
- **Performance otimizada**: Índices de banco e queries eficientes
- **Recursos assíncronos**: Processamento não-bloqueante de eventos
- **Horizontal scaling**: Arquitetura preparada para múltiplas instâncias

### **5. Observabilidade**
- **Logging estruturado**: Rastreamento completo de operações
- **Métricas detalhadas**: Monitoramento de performance e uso
- **Request tracing**: Acompanhamento de requisições end-to-end
- **Health checks**: Monitoramento automático de saúde do sistema

### **6. Documentação e Usabilidade**
- **API autodocumentada**: Swagger gerado automaticamente
- **Contratos claros**: DTOs bem definidos com validações
- **Exemplos práticos**: Casos de uso documentados
- **Nomes intuitivos**: URLs amigáveis com nomes de sessão

## 🚀 Status Atual e Próximos Passos

### **✅ Implementado**
- [x] Arquitetura Clean completa
- [x] Sistema de sessões WhatsApp
- [x] Envio de mensagens (texto, mídia, documentos)
- [x] Sistema de webhooks
- [x] Integração Chatwoot (estrutura)
- [x] Documentação Swagger
- [x] Sistema de migrações
- [x] Logging estruturado
- [x] Autenticação via API Key
- [x] Suporte a proxy
- [x] Múltiplas sessões simultâneas

### **🔄 Em Desenvolvimento**
- [ ] Testes unitários abrangentes
- [ ] Testes de integração
- [ ] Cliente Chatwoot API completo
- [ ] Métricas avançadas (Prometheus)
- [ ] Rate limiting
- [ ] Cache Redis para sessões

### **📋 Roadmap Futuro**
- [ ] Tracing distribuído (Jaeger/OpenTelemetry)
- [ ] Clustering e alta disponibilidade
- [ ] Backup automático de sessões
- [ ] Dashboard de monitoramento
- [ ] Webhooks com retry automático
- [ ] Suporte a múltiplos bancos de dados
- [ ] API GraphQL opcional
- [ ] Integração com outros sistemas de CRM

## 🔧 Tecnologias e Dependências

### **Core Technologies**
- **Go 1.21+** - Linguagem principal
- **Fiber v2** - Framework web de alta performance
- **PostgreSQL 15+** - Banco de dados principal
- **SQLx** - Extensões SQL para Go
- **whatsmeow** - Biblioteca WhatsApp Web API

### **Bibliotecas Principais**
- **zerolog** - Logging estruturado de alta performance
- **uuid** - Geração de identificadores únicos
- **swaggo/swag** - Geração automática de documentação Swagger
- **golang-migrate** - Sistema de migrações de banco
- **godotenv** - Carregamento de variáveis de ambiente

### **Ferramentas de Desenvolvimento**
- **Docker & Docker Compose** - Containerização
- **Make** - Automação de tarefas
- **Air v1.63.0** - Hot reload para desenvolvimento
- **golangci-lint** - Linting de código
- **swaggo/swag** - Geração de documentação Swagger

## 📚 Referências e Inspirações

### **Arquitetura**
- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

### **Tecnologias**
- [Fiber Framework Documentation](https://docs.gofiber.io/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [whatsmeow Library](https://github.com/tulir/whatsmeow)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/)

### **Padrões e Práticas**
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [Dependency Injection](https://martinfowler.com/articles/injection.html)

---

## 📊 Estatísticas do Projeto

- **Total de arquivos Go**: 35+
- **Linhas de código**: 8000+
- **Camadas implementadas**: 4 (Domain, Application, Infrastructure, Platform)
- **Domínios**: 4 (Session, Message, Webhook, Chatwoot)
- **Endpoints API**: 25+
- **Tabelas de banco**: 3 principais + whatsmeow
- **Migrações**: 6 (up/down)

**Status**: ✅ Arquitetura completa implementada e funcional
**Última atualização**: 2024-12-27
**Versão**: 1.0.0
