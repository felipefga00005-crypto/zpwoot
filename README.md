# zpwoot - Wameow Multi-Session API

Uma API REST completa para gerenciamento de múltiplas sessões do Wameow usando Go, Fiber, SQLx, PostgreSQL, Wameow (whatsmeow), integração com Chatwoot e webhooks.

## Características

- 🚀 **Multi-sessão**: Gerencie múltiplas sessões do Wameow simultaneamente
- 📱 **Wameow Web API**: Integração completa com whatsmeow
- 🔄 **Webhooks**: Sistema de webhooks para eventos em tempo real
- 💬 **Chatwoot**: Integração nativa com Chatwoot para atendimento
- 🗄️ **PostgreSQL**: Persistência de dados robusta
- 🔒 **Proxy Support**: Suporte a proxy HTTP e SOCKS5
- 📊 **Monitoramento**: Health checks e métricas

## Tecnologias

- **Go 1.21+**
- **Fiber v2** - Framework web rápido
- **SQLx** - Extensões SQL para Go
- **PostgreSQL** - Banco de dados
- **whatsmeow** - Biblioteca Wameow Web API
- **Zap** - Logger estruturado
- **UUID** - Geração de identificadores únicos

## Instalação

1. Clone o repositório:
```bash
git clone <repository-url>
cd zpwoot
```

2. Copie o arquivo de configuração:
```bash
cp .env.example .env
```

3. Configure as variáveis de ambiente no arquivo `.env`

4. Execute as migrações do banco de dados:
```bash
# As migrações são executadas automaticamente na inicialização
# Ou você pode executá-las manualmente:
make migrate-up

# Para verificar o status das migrações:
make migrate-status

# Para criar uma nova migração:
make migrate-create NAME=nome_da_migracao
```

5. Execute a aplicação:
```bash
go run cmd/zpwoot/main.go
```

## API Endpoints

### Gerenciamento de Sessões

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/sessions` | Criar uma nova sessão do Wameow |
| `GET` | `/sessions` | Listar todas as sessões (com filtros opcionais) |
| `GET` | `/sessions/{id}` | Detalhes de uma sessão específica |
| `DELETE` | `/sessions/{id}` | Remove permanentemente uma sessão |
| `POST` | `/sessions/{sessionId}/connect` | Estabelece a conexão da sessão com o Wameow |
| `POST` | `/sessions/{sessionId}/logout` | Faz o logout da sessão do Wameow |
| `GET` | `/sessions/{sessionId}/qr` | Recupera o QR Code atual (se existir) |
| `POST` | `/sessions/{sessionId}/pair` | Emparelha um telefone com a sessão |
| `POST` | `/sessions/{sessionId}/proxy` | Define proxy para a sessão |
| `GET` | `/sessions/{sessionId}/proxy` | Obtém configuração de proxy para a sessão |

### Wameow Messaging

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/Wameow/{sessionId}/send/text` | Enviar mensagem de texto |
| `POST` | `/Wameow/{sessionId}/send/media` | Enviar mensagem de mídia |
| `GET` | `/Wameow/{sessionId}/contacts` | Listar contatos |
| `GET` | `/Wameow/{sessionId}/chats` | Listar conversas |
| `GET` | `/Wameow/{sessionId}/status` | Status da sessão |

### Webhooks

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/webhooks/config` | Criar configuração de webhook |
| `GET` | `/webhooks/config` | Listar configurações de webhook |
| `PUT` | `/webhooks/config/{id}` | Atualizar configuração de webhook |
| `DELETE` | `/webhooks/config/{id}` | Remover configuração de webhook |
| `GET` | `/webhooks/events` | Listar eventos suportados |
| `POST` | `/webhooks/test/{id}` | Testar webhook |
| `POST` | `/webhooks/incoming/{sessionId}` | Webhook de entrada |

### Integração Chatwoot

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/sessions/{sessionId}/chatwoot/set` | Configurar Chatwoot (criar/atualizar) |
| `GET` | `/sessions/{sessionId}/chatwoot/find` | Buscar configuração Chatwoot |
| `POST` | `/chatwoot/sync/contacts` | Sincronizar contatos |
| `POST` | `/chatwoot/sync/conversations` | Sincronizar conversas |
| `POST` | `/chatwoot/webhook` | Webhook do Chatwoot |
| `POST` | `/chatwoot/test` | Testar conexão Chatwoot |
| `GET` | `/chatwoot/stats` | Estatísticas da integração |

**Nota:** A configuração do Chatwoot agora é específica por sessão e simplificada em apenas 2 endpoints:
- `set`: Cria ou atualiza a configuração (apenas uma configuração por sessão)
- `find`: Busca a configuração existente

## 🔐 Autenticação

Todos os endpoints da API (exceto `/health` e `/swagger/*`) requerem autenticação via API Key.

### Configuração da API Key

1. Configure a variável de ambiente `ZP_API_KEY` no arquivo `.env`:
```bash
ZP_API_KEY=dev-api-key-12345
```

2. Inclua a API Key no header `Authorization` de todas as requisições:
```bash
Authorization: dev-api-key-12345
```

**Nota:** Não use prefixo "Bearer " - apenas o valor da API Key diretamente.

## 🏷️ Nomes de Sessão URL-Friendly

O zpwoot suporta tanto UUID quanto **nomes de sessão legíveis** nas URLs da API, tornando-as mais intuitivas:

### ✅ Nomes Válidos
- **Formato**: 3-50 caracteres, começar com letra, usar apenas letras, números, hífens e underscores
- **Exemplos**: `my-Wameow-session`, `customer-support`, `sales-team-1`

### ❌ Nomes Inválidos
- Palavras reservadas: `create`, `list`, `info`, `delete`, etc.
- Caracteres especiais: `my@session`, `session#1`
- Muito curto/longo: `ab` ou nomes com mais de 50 caracteres

### 🔄 Sugestões Automáticas
Se você fornecer um nome inválido, a API sugerirá uma alternativa válida.

## Exemplos de Uso

### Criar uma nova sessão

```bash
curl -X POST http://localhost:8080/sessions/create \
  -H "Content-Type: application/json" \
  -H "Authorization: dev-api-key-12345" \
  -d '{
    "name": "my-Wameow-session",
    "deviceJid": "5511999999999@s.Wameow.net"
  }'
```

### Usar nome da sessão nas URLs

```bash
# Obter informações da sessão usando o nome
curl -H "Authorization: dev-api-key-12345" \
  http://localhost:8080/sessions/my-Wameow-session/info

# Conectar sessão usando o nome
curl -X POST -H "Authorization: dev-api-key-12345" \
  http://localhost:8080/sessions/my-Wameow-session/connect

# Obter QR code usando o nome
curl -H "Authorization: dev-api-key-12345" \
  http://localhost:8080/sessions/my-Wameow-session/qr
```

### Listar sessões

```bash
curl http://localhost:8080/sessions/list?status=connected&limit=10
```

### Conectar uma sessão

```bash
curl -X POST http://localhost:8080/sessions/{session-id}/connect
```

### Obter QR Code

```bash
curl http://localhost:8080/sessions/{session-id}/qr
```

### Configurar Proxy

```bash
curl -X POST http://localhost:8080/sessions/{session-id}/proxy/set \
  -H "Content-Type: application/json" \
  -d '{
    "type": "http",
    "host": "proxy.example.com",
    "port": 8080,
    "username": "user",
    "password": "pass"
  }'
```

## Status de Sessão

- `created` - Sessão criada mas não conectada
- `connecting` - Tentando conectar
- `connected` - Conectada e ativa
- `disconnected` - Desconectada
- `error` - Erro na conexão
- `logged_out` - Logout realizado

## Configuração de Proxy

Suporte para proxies HTTP e SOCKS5:

```json
{
  "type": "http|socks5",
  "host": "proxy.example.com",
  "port": 8080,
  "username": "optional",
  "password": "optional"
}
```

## Estrutura do Projeto

```
zpwoot/
├── cmd/zpwoot/           # Ponto de entrada da aplicação
├── internal/
│   ├── domain/           # Lógica de negócio
│   │   └── session/      # Domínio de sessões
│   ├── app/              # Casos de uso
│   ├── ports/            # Interfaces/Portas
│   └── infra/            # Infraestrutura
│       ├── http/         # Handlers HTTP
│       ├── repositories/ # Repositórios
│       └── integrations/ # Integrações externas
├── platform/             # Plataforma (config, db, logger)
├── pkg/                  # Pacotes utilitários
└── migrations/           # Migrações do banco
```

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 📝 Sistema de Logging

O zpwoot utiliza **zerolog** para um sistema de logging estruturado e performático.

### Configuração de Logs

| Variável | Valores | Descrição |
|----------|---------|-----------|
| `LOG_LEVEL` | `trace`, `debug`, `info`, `warn`, `error`, `fatal` | Nível de log |
| `LOG_FORMAT` | `console`, `json` | Formato de saída |
| `LOG_OUTPUT` | `stdout`, `stderr`, `file`, `path/to/file` | Destino dos logs |

### Exemplos de Uso

```go
// Logger básico
logger := logger.New("info")
logger.Info("Aplicação iniciada")
logger.Error("Erro na conexão")

// Logger com campos estruturados
logger.InfoWithFields("Sessão criada", map[string]interface{}{
    "session_id": "sess_123",
    "user_id": "user_456",
    "timestamp": time.Now(),
})

// Logger com contexto persistente
sessionLogger := logger.WithFields(map[string]interface{}{
    "session_id": "sess_123",
    "component": "Wameow",
})
sessionLogger.Info("Mensagem enviada")
sessionLogger.Debug("QR code gerado")
```

### Logs Estruturados

O sistema suporta logs estruturados para melhor observabilidade:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "caller": "session/service.go:45",
  "message": "Session created",
  "session_id": "sess_abc123",
  "user_id": "user_456",
  "component": "session_manager"
}
```

## Suporte

Para suporte e dúvidas, abra uma issue no repositório do projeto.
