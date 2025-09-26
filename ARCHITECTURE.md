# zpwoot - Clean Architecture Implementation

## 📋 Visão Geral

Este documento descreve a implementação completa da arquitetura Clean Architecture no projeto zpwoot, incluindo repositórios, use cases, DTOs e documentação Swagger.

## 🏗️ Estrutura da Arquitetura

```
zpwoot/
├── cmd/zpwoot/                    # Entry point da aplicação
├── internal/
│   ├── app/                       # 🎯 APPLICATION LAYER
│   │   ├── app.go                 # Entry point e índice da camada de aplicação
│   │   ├── container.go           # Dependency Injection Container
│   │   ├── README.md              # Documentação da camada
│   │   ├── common/
│   │   │   ├── dto.go             # DTOs comuns (SuccessResponse, etc.)
│   │   │   └── usecase.go         # Use cases comuns (health, stats)
│   │   ├── session/
│   │   │   ├── dto.go             # DTOs de sessões Wameow
│   │   │   └── usecase.go         # Use cases de sessões
│   │   ├── webhook/
│   │   │   ├── dto.go             # DTOs de webhooks
│   │   │   └── usecase.go         # Use cases de webhooks
│   │   └── chatwoot/
│   │       ├── dto.go             # DTOs de integração Chatwoot
│   │       └── usecase.go         # Use cases de Chatwoot
│   ├── domain/                    # 🏛️ DOMAIN LAYER
│   │   ├── session/
│   │   │   ├── entity.go          # Entidades de sessão
│   │   │   └── service.go         # Serviços de domínio
│   │   ├── webhook/
│   │   │   ├── entity.go          # Entidades de webhook
│   │   │   └── service.go         # Serviços de domínio
│   │   └── chatwoot/
│   │       ├── entity.go          # Entidades de Chatwoot
│   │       └── service.go         # Serviços de domínio
│   ├── ports/                     # 🔌 INTERFACES
│   │   ├── session_repository.go  # Interface do repositório de sessões
│   │   ├── webhook_repository.go  # Interface do repositório de webhooks
│   │   └── chatwoot_repository.go # Interface do repositório de Chatwoot
│   └── infra/                     # 🔧 INFRASTRUCTURE LAYER
│       ├── db/
│       │   └── migrations/        # Migrações do banco de dados
│       ├── http/
│       │   ├── handlers/          # HTTP handlers
│       │   └── routers/           # Roteamento HTTP
│       └── repository/            # 💾 IMPLEMENTAÇÕES DOS REPOSITÓRIOS
│           ├── repository.go      # Factory de repositórios
│           ├── session_repository.go    # Repositório de sessões
│           ├── webhook_repository.go    # Repositório de webhooks
│           └── chatwoot_repository.go   # Repositório de Chatwoot
├── docs/
│   ├── API.md                     # Documentação da API
│   └── swagger/                   # Documentação Swagger gerada
└── platform/                     # Utilitários e configurações
```

## 🎯 Camada de Aplicação (Use Cases)

### **DTOs (Data Transfer Objects)**
- **Localização**: `internal/app/{domain}/dto.go`
- **Responsabilidade**: Contratos da API, validação e serialização
- **Exemplos**:
  - `CreateSessionRequest/Response`
  - `ListWebhooksRequest/Response`
  - `ChatwootConfigResponse`

### **Use Cases**
- **Localização**: `internal/app/{domain}/usecase.go`
- **Responsabilidade**: Orquestração da lógica de negócio
- **Padrões**:
  - Interface + Implementação
  - Conversão entre DTOs e entidades de domínio
  - Coordenação entre repositórios e serviços

### **Container de Dependências**
- **Arquivo**: `internal/app/container.go`
- **Responsabilidade**: Dependency Injection e configuração
- **Funcionalidades**:
  - Criação de todos os use cases
  - Configuração de dependências
  - Factory pattern para componentes

## 🏛️ Camada de Domínio

### **Entidades**
- **Localização**: `internal/domain/{domain}/entity.go`
- **Características**:
  - Regras de negócio puras
  - Sem dependências externas
  - Validações de domínio
  - Erros específicos do domínio

### **Serviços de Domínio**
- **Localização**: `internal/domain/{domain}/service.go`
- **Responsabilidade**: Lógica de negócio complexa
- **Exemplos**:
  - Validação de configurações
  - Processamento de eventos
  - Transformações de dados

## 💾 Camada de Infraestrutura

### **Repositórios**
- **Localização**: `internal/infra/repository/`
- **Implementações**:
  - `SessionRepository` - PostgreSQL com JSONB
  - `WebhookRepository` - PostgreSQL com arrays
  - `ChatwootRepository` - PostgreSQL com relacionamentos

### **Características dos Repositórios**
- **Mapeamento OR**: Conversão entre modelos de banco e entidades
- **Queries Otimizadas**: Índices e consultas eficientes
- **Tratamento de Erros**: Erros específicos do domínio
- **Logging**: Rastreamento de operações

### **Migrações**
- **Localização**: `internal/infra/db/migrations/`
- **Tabelas Criadas**:
  - `zpSessions` - Sessões Wameow
  - `zpWebhooks` - Configurações de webhook
  - `zpChatwoot` - Configurações Chatwoot

## 🔌 Interfaces (Ports)

### **Repositórios**
- `SessionRepository` - CRUD de sessões
- `WebhookRepository` - CRUD de webhooks + estatísticas
- `ChatwootRepository` - CRUD de configurações + sincronização

### **Integrações Externas**
- `WameowManager` - Gerenciamento de sessões Wameow
- `ChatwootIntegration` - API do Chatwoot

## 📖 Documentação Swagger

### **Configuração**
- **Geração**: `make swagger`
- **Servidor**: `make swagger-serve`
- **URL**: http://localhost:8080/swagger/

### **Estrutura**
- **Comentários**: Nos handlers HTTP
- **DTOs**: Documentados com exemplos
- **Tags**: Organizados por domínio
- **Validações**: Especificadas nos DTOs

## 🛠️ Comandos Disponíveis

### **Desenvolvimento**
```bash
# Compilar projeto
go build ./...

# Gerar documentação Swagger
make swagger

# Servir com Swagger UI
make swagger-serve

# Iniciar serviços principais
make up

# Iniciar Chatwoot
make up-cw
```

### **Testes**
```bash
# Executar testes
go test ./...

# Testes com cobertura
make test-coverage

# Testar Swagger
make swagger-test
```

## 🔄 Fluxo de Dados

```
HTTP Request → Handler → DTO → Use Case → Domain Service → Repository → Database
     ↓           ↑        ↑       ↓           ↑             ↑          ↓
  Validation  Response  Convert  Business   Domain      Database   Storage
                                  Logic     Entity      Model
```

## ✅ Benefícios da Implementação

### **1. Separação de Responsabilidades**
- Cada camada tem responsabilidades bem definidas
- Baixo acoplamento entre componentes
- Alta coesão dentro de cada módulo

### **2. Testabilidade**
- Use cases testáveis independentemente
- Mocks fáceis de criar
- Testes unitários isolados

### **3. Manutenibilidade**
- Mudanças na API não afetam o domínio
- Mudanças no banco não afetam a lógica
- Evolução independente das camadas

### **4. Documentação**
- API documentada automaticamente
- Contratos claros via DTOs
- Exemplos de uso disponíveis

## 🚀 Próximos Passos

### **1. Implementação Completa**
- [ ] Conectar use cases aos handlers
- [ ] Implementar serviços de domínio completos
- [ ] Adicionar validações nos DTOs

### **2. Integrações**
- [ ] Wameow Web API
- [ ] Chatwoot API client
- [ ] Sistema de webhooks

### **3. Testes**
- [ ] Testes unitários para use cases
- [ ] Testes de integração para repositórios
- [ ] Testes end-to-end para API

### **4. Observabilidade**
- [ ] Métricas de performance
- [ ] Tracing distribuído
- [ ] Alertas e monitoramento

## 📚 Referências

- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Fiber Framework](https://docs.gofiber.io/)

---

**Status**: ✅ Estrutura completa implementada e compilando
**Última atualização**: 2024-01-01
