# zpwoot API Documentation

## 📖 Swagger Documentation

A documentação completa da API está disponível através do Swagger UI.

### 🚀 Como acessar

1. **Inicie o servidor:**
   ```bash
   make swagger-serve
   ```

2. **Acesse a documentação:**
   - **Swagger UI**: http://localhost:8080/swagger/
   - **JSON Schema**: http://localhost:8080/swagger/doc.json
   - **YAML Schema**: http://localhost:8080/swagger/swagger.yaml

### 🛠️ Comandos disponíveis

```bash
# Gerar documentação Swagger
make swagger

# Servir documentação localmente
make swagger-serve

# Testar endpoints da documentação
make swagger-test
```

## 📋 Endpoints Principais

### Health Check
- **GET** `/health` - Verificar status da API

### Sessions (Sessões Wameow)
- **POST** `/sessions/create` - Criar nova sessão
- **GET** `/sessions/list` - Listar sessões
- **GET** `/sessions/{sessionId}/info` - Informações da sessão
- **DELETE** `/sessions/{sessionId}/delete` - Deletar sessão
- **POST** `/sessions/{sessionId}/connect` - Conectar sessão
- **POST** `/sessions/{sessionId}/logout` - Desconectar sessão
- **GET** `/sessions/{sessionId}/qr` - Obter QR Code
- **POST** `/sessions/{sessionId}/pair` - Parear telefone

### Webhooks
- **POST** `/sessions/{sessionId}/webhook/config` - Configurar webhook
- **GET** `/sessions/{sessionId}/webhook/config` - Obter configuração webhook

### Chatwoot Integration
- **POST** `/sessions/{sessionId}/chatwoot/config` - Configurar Chatwoot
- **GET** `/sessions/{sessionId}/chatwoot/config` - Obter configuração Chatwoot

## 🏗️ Estrutura de DTOs

### Camada de Aplicação (`internal/app/`)

```
internal/app/
├── dto.go              # Índice e re-exports de todos os DTOs
├── common/
│   └── dto.go          # DTOs comuns (SuccessResponse, ErrorResponse, etc.)
├── session/
│   └── dto.go          # DTOs para sessões Wameow
├── webhook/
│   └── dto.go          # DTOs para webhooks
└── chatwoot/
    └── dto.go          # DTOs para integração Chatwoot
```

### 📦 Arquivo de Índice (`internal/app/dto.go`)

O arquivo `dto.go` na raiz do pacote `app` serve como um índice central que:
- Re-exporta todos os DTOs para facilitar imports
- Evita imports longos nos handlers
- Centraliza as funções de conversão
- Mantém a organização por domínio

### Principais DTOs

#### Respostas Comuns
- `SuccessResponse` - Resposta de sucesso padrão
- `ErrorResponse` - Resposta de erro padrão
- `ValidationErrorResponse` - Erros de validação
- `PaginationResponse` - Metadados de paginação

#### Sessões
- `CreateSessionRequest/Response` - Criar sessão
- `SessionInfoResponse` - Informações da sessão
- `QRCodeResponse` - QR Code para pareamento
- `ListSessionsRequest/Response` - Listar sessões

#### Webhooks
- `SetConfigRequest/Response` - Configurar webhook
- `WebhookEventResponse` - Eventos de webhook
- `TestWebhookRequest/Response` - Testar webhook

#### Chatwoot
- `CreateChatwootConfigRequest/Response` - Configurar Chatwoot
- `SyncContactRequest/Response` - Sincronizar contatos
- `SendMessageToChatwootRequest/Response` - Enviar mensagens

## 🔧 Desenvolvimento

### Adicionando novos endpoints

1. **Criar DTOs** em `internal/app/{domain}/dto.go`
2. **Adicionar comentários Swagger** nos handlers:
   ```go
   // @Summary Descrição breve
   // @Description Descrição detalhada
   // @Tags NomeTag
   // @Accept json
   // @Produce json
   // @Param id path string true "ID do recurso"
   // @Success 200 {object} object "Sucesso"
   // @Failure 400 {object} object "Erro"
   // @Router /endpoint [method]
   ```
3. **Regenerar documentação**: `make swagger`

### Estrutura dos comentários Swagger

```go
// @Summary         - Resumo do endpoint
// @Description     - Descrição detalhada
// @Tags           - Categoria/grupo do endpoint
// @Accept         - Tipo de conteúdo aceito (json, xml, etc.)
// @Produce        - Tipo de conteúdo retornado
// @Param          - Parâmetros (path, query, body)
// @Success        - Resposta de sucesso
// @Failure        - Respostas de erro
// @Router         - Rota e método HTTP
```

## 🧪 Testando a API

### Usando curl

```bash
# Health check
curl http://localhost:8080/health

# Listar sessões
curl http://localhost:8080/sessions/list

# Criar sessão
curl -X POST http://localhost:8080/sessions/create \
  -H "Content-Type: application/json" \
  -d '{"name": "Minha Sessão"}'
```

### Usando Swagger UI

1. Acesse http://localhost:8080/swagger/
2. Explore os endpoints disponíveis
3. Teste diretamente pela interface
4. Veja exemplos de request/response

## 📚 Recursos Adicionais

- **Swagger Specification**: https://swagger.io/specification/
- **Fiber Swagger**: https://github.com/swaggo/fiber-swagger
- **Swaggo**: https://github.com/swaggo/swag

## 🐛 Troubleshooting

### Erro: "cannot find type definition"
- Verifique se os tipos estão importados corretamente
- Use `object` como tipo genérico se necessário
- Regenere a documentação: `make swagger`

### Swagger UI não carrega
- Verifique se o servidor está rodando
- Confirme que a rota `/swagger/*` está configurada
- Verifique os logs do servidor

### Documentação desatualizada
- Sempre execute `make swagger` após mudanças
- Reinicie o servidor após regenerar docs
