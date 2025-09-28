# 🚀 Plano Final - Integração Chatwoot zpwoot

## 📋 Visão Geral
Implementação completa da integração Chatwoot baseada na análise da Evolution API (2.560 linhas), seguindo a arquitetura Clean Architecture do zpwoot.

## 🗃️ Estrutura de Banco de Dados

### ✅ Migrações Criadas
- **003_create_chatwoot_config_table.up.sql** - Tabela zpChatwoot com sessionId
- **004_create_zp_message_table.up.sql** - Tabela zpMessage para mapeamento

### 📊 Tabela zpChatwoot (1:1 com zpSessions)
```sql
CREATE TABLE "zpChatwoot" (
    "id" UUID PRIMARY KEY,
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    "url" VARCHAR(2048) NOT NULL,
    "token" VARCHAR(255) NOT NULL,
    "accountId" VARCHAR(50) NOT NULL,
    "inboxId" VARCHAR(50),
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- UMA configuração Chatwoot por sessão
CREATE UNIQUE INDEX "idx_zp_chatwoot_unique_session" ON "zpChatwoot" ("sessionId");
```

### 📊 Tabela zpMessage (Mapeamento WhatsApp ↔ Chatwoot)
```sql
CREATE TABLE "zpMessage" (
    "id" UUID PRIMARY KEY,
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"("id") ON DELETE CASCADE,
    
    -- WhatsApp (baseado em whatsmeow)
    "zpMessageId" VARCHAR(255) NOT NULL,
    "zpSender" VARCHAR(255) NOT NULL,
    "zpChat" VARCHAR(255) NOT NULL,
    "zpTimestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
    "zpFromMe" BOOLEAN NOT NULL,
    "zpType" VARCHAR(50) NOT NULL, -- text, image, audio, video, document, contact
    "content" TEXT,
    
    -- Chatwoot
    "cwMessageId" INTEGER,
    "cwConversationId" INTEGER,
    
    -- Status
    "syncStatus" VARCHAR(20) DEFAULT 'pending' CHECK ("syncStatus" IN ('pending', 'synced', 'failed')),
    
    -- Timestamps
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Evitar duplicatas
CREATE UNIQUE INDEX "idx_zp_message_unique_zp" ON "zpMessage" ("sessionId", "zpMessageId");
```

## 🏗️ Estrutura de Arquivos

### 📁 internal/infra/chatwoot/ (Nova Integração)
```
internal/infra/chatwoot/
├── client.go        # Cliente HTTP Chatwoot API
├── manager.go       # Gerenciamento configurações
├── webhook.go       # Processamento webhooks
├── contact.go       # Sincronização contatos
├── conversation.go  # Sincronização conversas
├── formatter.go     # Formatação mensagens
├── import.go        # Importação histórico
└── utils.go         # Utilitários
```

### 📁 Arquivos a Modificar
- `internal/app/chatwoot/dto.go` - Expandir DTOs
- `internal/ports/chatwoot.go` - Interfaces
- `internal/infra/wameow/events.go` - Integração WhatsApp → Chatwoot
- `internal/infra/http/handlers/chatwoot.go` - Novos endpoints

## 🎯 Funcionalidades Implementadas

### ✅ Core Features
- **Sincronização bidirecional** WhatsApp ↔ Chatwoot
- **Auto-criação** de inbox e contatos
- **Mapeamento de mensagens** para reply/edit/delete
- **Processamento de webhooks** com delay 500ms

### ✅ Evolution API Features
- **Merge contatos brasileiros** (+55)
- **Importação histórico** limitada por dias
- **Assinatura mensagens** personalizadas
- **Formatação markdown** automática
- **Processamento mídias** e anexos

### ✅ Funcionalidades Avançadas
- **Filtros eventos** e mensagens
- **Quoted messages** e reactions
- **Gestão conversas** (reopen, status)
- **Contato bot** automático (123456)
- **Ephemeral messages** e grupos

## 📋 PLANO DE IMPLEMENTAÇÃO

### 🔥 FASE 1: Estrutura Base e Migrações (2-3 horas)

#### 1.1 Executar Migrações de Banco (30 min)
```bash
# Verificar status atual
make migrate-status

# Aplicar migrações (se necessário)
make migrate-up

# Verificar tabelas criadas
psql -d zpwoot -c "\dt"
```

**Validação:**
- [ ] Tabela `zpChatwoot` criada com `sessionId`
- [ ] Tabela `zpMessage` criada com campos corretos
- [ ] Índices e constraints aplicados

#### 1.2 Criar Estrutura internal/infra/chatwoot/ (45 min)
```bash
mkdir -p internal/infra/chatwoot
touch internal/infra/chatwoot/{client,manager,webhook,contact,conversation,formatter,import,utils}.go
```

**Arquivos a criar:**
- [ ] `client.go` - Estrutura básica do cliente
- [ ] `manager.go` - Estrutura básica do manager
- [ ] `webhook.go` - Estrutura básica webhook handler
- [ ] `contact.go` - Estrutura básica sync contatos
- [ ] `conversation.go` - Estrutura básica sync conversas
- [ ] `formatter.go` - Estrutura básica formatação
- [ ] `import.go` - Estrutura básica importação
- [ ] `utils.go` - Estrutura básica utilitários

#### 1.3 Implementar Interfaces e Contratos (45 min)
**Arquivo:** `internal/ports/chatwoot.go`

```go
type ChatwootClient interface {
    CreateInbox(name, webhookURL string) (*Inbox, error)
    CreateContact(phone, name string, inboxID int) (*Contact, error)
    CreateConversation(contactID, inboxID int) (*Conversation, error)
    SendMessage(conversationID int, content string) error
    SendMediaMessage(conversationID int, content string, attachment io.Reader, filename string) error
    FindContact(phone string, inboxID int) (*Contact, error)
    GetConversation(contactID, inboxID int) (*Conversation, error)
    ListInboxes() ([]Inbox, error)
}

type ChatwootManager interface {
    GetClient(sessionID string) (ChatwootClient, error)
    InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error
    IsEnabled(sessionID string) bool
}

type WebhookHandler interface {
    ProcessWebhook(ctx context.Context, webhook *dto.WebhookRequest, sessionID string) error
}
```

#### 1.4 Criar DTOs Expandidos (30 min)
**Arquivo:** `internal/app/chatwoot/dto.go`

```go
type WebhookRequest struct {
    Account           Account                `json:"account"`
    Conversation      Conversation           `json:"conversation"`
    Message           Message                `json:"message"`
    Contact           Contact                `json:"contact"`
    Event             string                 `json:"event"`
    Private           bool                   `json:"private"`
    ContentAttributes map[string]interface{} `json:"content_attributes"`
    Meta              Meta                   `json:"meta"`
}

type Account struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type Conversation struct {
    ID                   int                    `json:"id"`
    Status               string                 `json:"status"`
    InboxID              int                    `json:"inbox_id"`
    ContactID            int                    `json:"contact_id"`
    Messages             []Message              `json:"messages"`
    AdditionalAttributes map[string]interface{} `json:"additional_attributes"`
}

type Message struct {
    ID          int          `json:"id"`
    Content     string       `json:"content"`
    MessageType string       `json:"message_type"`
    Attachments []Attachment `json:"attachments"`
    Sender      Sender       `json:"sender"`
    SourceID    string       `json:"source_id"`
}

type Meta struct {
    Sender Sender `json:"sender"`
}

type Sender struct {
    ID            int    `json:"id"`
    Name          string `json:"name"`
    Identifier    string `json:"identifier"`
    AvailableName string `json:"available_name"`
}
```

### 🔥 FASE 2: Cliente Chatwoot API (4-5 horas)

#### 2.1 Implementar client.go (2 horas)
**Arquivo:** `internal/infra/chatwoot/client.go`

**Estrutura base:**
```go
type Client struct {
    baseURL     string
    token       string
    accountID   string
    httpClient  *http.Client
    logger      *logger.Logger
}

func NewClient(baseURL, token, accountID string, logger *logger.Logger) *Client
```

**Métodos a implementar:**
- [ ] `CreateInbox(name, webhookURL string) (*Inbox, error)`
- [ ] `CreateContact(phone, name string, inboxID int) (*Contact, error)`
- [ ] `CreateConversation(contactID, inboxID int) (*Conversation, error)`
- [ ] `SendMessage(conversationID int, content string) error`
- [ ] `SendMediaMessage(conversationID int, content string, attachment io.Reader, filename string) error`
- [ ] `FindContact(phone string, inboxID int) (*Contact, error)`
- [ ] `GetConversation(contactID, inboxID int) (*Conversation, error)`
- [ ] `ListInboxes() ([]Inbox, error)`

**Baseado na Evolution API:**
- Timeout 30s para requests
- Retry automático (3 tentativas)
- Headers: `api_access_token`, `Content-Type: application/json`
- Error handling com logs estruturados

#### 2.2 Implementar manager.go (1.5 horas)
**Arquivo:** `internal/infra/chatwoot/manager.go`

**Funcionalidades:**
- [ ] Cache de clientes por sessionID
- [ ] Auto-criação de inbox (baseado Evolution API)
- [ ] Verificação de contato bot (123456)
- [ ] Inicialização com QR code
- [ ] Cleanup de recursos

**Baseado na Evolution API `initInstanceChatwoot()`:**
```go
func (m *Manager) InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error {
    // 1. Listar inboxes existentes
    // 2. Verificar se inbox já existe
    // 3. Criar nova inbox se não existir
    // 4. Criar contato bot se habilitado
    // 5. Cache do cliente
}
```

#### 2.3 Implementar utils.go (1 hora)
**Arquivo:** `internal/infra/chatwoot/utils.go`

**Utilitários:**
- [ ] Conversão JID WhatsApp ↔ Phone
- [ ] Validação de URLs e tokens
- [ ] Formatação de números brasileiros (+55)
- [ ] Retry logic com backoff
- [ ] Error wrapping e logging

#### 2.4 Testes Unitários do Cliente (30 min)
**Arquivo:** `internal/infra/chatwoot/client_test.go`

- [ ] Mock do HTTP client
- [ ] Testes de sucesso e erro
- [ ] Validação de headers e payloads
- [ ] Testes de retry logic

### 🔥 FASE 3: Sincronização Bidirecional (6-7 horas)

#### 3.1 Implementar webhook.go (2.5 horas)
**Arquivo:** `internal/infra/chatwoot/webhook.go`

**Baseado na Evolution API `receiveWebhook()` - Linha 1236:**
```go
func (h *WebhookHandler) ProcessWebhook(ctx context.Context, webhook *dto.WebhookRequest, sessionID string) error {
    // 1. Delay 500ms para evitar race conditions
    time.Sleep(500 * time.Millisecond)
    
    // 2. Filtros de eventos
    if webhook.Private || 
       (webhook.Event == "message_updated" && webhook.ContentAttributes["deleted"] == nil) {
        return nil
    }
    
    // 3. Processar status de conversa
    if webhook.Event == "conversation_status_changed" && webhook.Conversation.Status == "resolved" {
        return h.handleConversationResolved(sessionID, webhook)
    }
    
    // 4. Processar mensagens deletadas
    if webhook.Event == "message_updated" && webhook.ContentAttributes["deleted"] != nil {
        return h.handleMessageDeleted(sessionID, webhook)
    }
    
    // 5. Filtrar mensagens do bot
    if h.isBotMessage(webhook) {
        return nil
    }
    
    // 6. Enviar para WhatsApp
    return h.sendToWhatsApp(sessionID, webhook)
}
```

**Funcionalidades:**
- [ ] Processamento de eventos Chatwoot
- [ ] Filtros específicos (private, bot, etc.)
- [ ] Conversão para mensagens WhatsApp
- [ ] Envio via whatsmeow
- [ ] Atualização de status na zpMessage

#### 3.2 Integrar WhatsApp → Chatwoot (2 horas)
**Arquivo:** `internal/infra/wameow/events.go` (modificar)

**Baseado na Evolution API `eventWhatsapp()` - Linha 1915:**
```go
func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
    // Processamento atual...
    
    // NOVA INTEGRAÇÃO CHATWOOT
    if h.chatwootEnabled(sessionID) {
        go h.sendToChatwoot(evt, sessionID) // Async
    }
}

func (h *EventHandler) sendToChatwoot(evt *events.Message, sessionID string) {
    // 1. Filtros Evolution API
    if evt.Info.Chat.String() == "status@broadcast" {
        return
    }
    
    // 2. Processar mensagens efêmeras
    if evt.Message.EphemeralMessage != nil {
        evt.Message = evt.Message.EphemeralMessage.Message
    }
    
    // 3. Formatação markdown
    content := h.formatMarkdownForChatwoot(evt.Message)
    
    // 4. Processar quoted messages
    quotedMsg := h.getQuotedMessage(evt.Message)
    
    // 5. Detectar mídia e reações
    isMedia := h.isMediaMessage(evt.Message)
    
    // 6. Criar/buscar conversa no Chatwoot
    conversation := h.createOrGetConversation(sessionID, evt.Info.Chat)
    
    // 7. Enviar para Chatwoot
    h.sendMessageToChatwoot(conversation.ID, content, isMedia, quotedMsg)
    
    // 8. Salvar mapeamento na zpMessage
    h.saveMessageMapping(sessionID, evt, conversation.ID, messageID)
}
```

#### 3.3 Implementar formatter.go (1 hour)
**Arquivo:** `internal/infra/chatwoot/formatter.go`

**Baseado na Evolution API formatação markdown:**
```go
// WhatsApp → Chatwoot
func FormatMarkdownForChatwoot(content string) string {
    // * → **
    // _ → *
    // ~ → ~~
}

// Chatwoot → WhatsApp
func FormatMarkdownForWhatsApp(content string) string {
    // ** → *
    // * → _
    // ~~ → ~
}
```

**Funcionalidades:**
- [ ] Formatação markdown bidirecional
- [ ] Processamento de mídias
- [ ] Quoted messages
- [ ] Reactions
- [ ] Contact messages

#### 3.4 Sistema de Mapeamento de Mensagens (30 min)
**Arquivo:** `internal/infra/chatwoot/message_mapper.go`

```go
type MessageMapper struct {
    repo repository.MessageRepository
}

func (m *MessageMapper) SaveMapping(sessionID, zpMessageID string, cwMessageID, cwConversationID int) error
func (m *MessageMapper) FindByZpMessageID(sessionID, zpMessageID string) (*domain.ZpMessage, error)
func (m *MessageMapper) FindByCwMessageID(cwMessageID int) (*domain.ZpMessage, error)
func (m *MessageMapper) UpdateSyncStatus(id string, status string) error
```

### 🔥 FASE 4: Funcionalidades Avançadas (5-6 horas)

#### 4.1 Implementar contact.go (1.5 horas)
**Arquivo:** `internal/infra/chatwoot/contact.go`

**Funcionalidades Evolution API:**
- [ ] Merge contatos brasileiros (+55)
- [ ] Importação de contatos
- [ ] Criação automática
- [ ] Normalização de números

```go
func (c *ContactSync) MergeBrazilianContacts(phone string) string {
    // Normalizar +5511999999999 → 11999999999
    // Verificar duplicatas
    // Merge automático
}
```

#### 4.2 Implementar conversation.go (1.5 horas)
**Arquivo:** `internal/infra/chatwoot/conversation.go`

**Funcionalidades:**
- [ ] Reabrir conversações
- [ ] Status pending
- [ ] Criação automática
- [ ] Gestão de estados

#### 4.3 Implementar import.go (1.5 horas)
**Arquivo:** `internal/infra/chatwoot/import.go`

**Funcionalidades Evolution API:**
- [ ] Importação histórico mensagens
- [ ] Limite por dias (`daysLimitImportMessages`)
- [ ] Importação contatos
- [ ] Progress tracking

#### 4.4 Funcionalidades Evolution API (1 hora)
**Implementar em arquivos existentes:**

- [ ] **Assinatura mensagens** (`signMsg`, `signDelimiter`)
- [ ] **IgnoreJIDs** (filtrar contatos)
- [ ] **Organização/Logo** (metadados)
- [ ] **Contato bot** (123456)
- [ ] **QR code inicial** (primeira conversa)

#### 4.5 Endpoints HTTP Adicionais (30 min)
**Arquivo:** `internal/infra/http/handlers/chatwoot.go`

**Novos endpoints:**
- [ ] `POST /webhook/:sessionId` - Receber webhooks
- [ ] `POST /sessions/{id}/chatwoot/auto-create` - Auto-criar inbox
- [ ] `GET /sessions/{id}/chatwoot/status` - Status integração
- [ ] `POST /sessions/{id}/chatwoot/import` - Importar histórico

### 🔥 FASE 5: Testes e Validação (3-4 horas)

#### 5.1 Testes Unitários Completos (1.5 horas)
**Arquivos de teste:**
- [ ] `webhook_test.go`
- [ ] `contact_test.go`
- [ ] `conversation_test.go`
- [ ] `formatter_test.go`
- [ ] `import_test.go`

#### 5.2 Testes de Integração (1 hora)
- [ ] Webhook bidirecional real
- [ ] Sincronização mensagens
- [ ] Criação contatos/conversas

#### 5.3 Testes de Performance (30 min)
- [ ] Processamento muitas mensagens
- [ ] Sincronização em massa
- [ ] Import histórico

#### 5.4 Documentação e README (30 min)
**Arquivo:** `internal/infra/chatwoot/README.md`

- [ ] Configuração
- [ ] Exemplos de uso
- [ ] Troubleshooting
- [ ] API Reference

#### 5.5 Validação Final (30 min)
- [ ] Testar todos os fluxos
- [ ] Comparar com Evolution API
- [ ] Verificar compatibilidade

## 🎯 Cronograma Estimado

| Fase | Tempo | Descrição |
|------|-------|-----------|
| **FASE 1** | 2-3h | Estrutura base e migrações |
| **FASE 2** | 4-5h | Cliente Chatwoot API |
| **FASE 3** | 6-7h | Sincronização bidirecional |
| **FASE 4** | 5-6h | Funcionalidades avançadas |
| **FASE 5** | 3-4h | Testes e validação |
| **TOTAL** | **20-25h** | **3-4 dias de trabalho** |

## ✅ Checklist de Validação Final

### Funcionalidades Core
- [ ] Mensagem WhatsApp → Chatwoot
- [ ] Mensagem Chatwoot → WhatsApp
- [ ] Auto-criação inbox
- [ ] Mapeamento mensagens
- [ ] Reply/Edit/Delete

### Funcionalidades Evolution API
- [ ] Merge contatos brasileiros
- [ ] Importação histórico
- [ ] Assinatura mensagens
- [ ] Formatação markdown
- [ ] Processamento mídias

### Testes
- [ ] Testes unitários passando
- [ ] Testes integração funcionando
- [ ] Performance adequada
- [ ] Documentação completa

## 🚀 Comando de Início

```bash
# 1. Aplicar migrações
make migrate-up

# 2. Verificar estrutura
make migrate-status

# 3. Iniciar implementação FASE 1
mkdir -p internal/infra/chatwoot
```

**Pronto para implementação! 🎯**
