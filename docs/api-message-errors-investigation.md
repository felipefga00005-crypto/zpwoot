# Investigação de Erros das Rotas de Mensagens

Este documento analisa os erros encontrados durante os testes das rotas de mensagens, com base no código fonte e na documentação do whatsmeow.

## 📋 Resumo dos Erros Investigados

| Erro | Rota | Causa Identificada | Solução Proposta |
|------|------|-------------------|------------------|
| 1 | `/send/media` | Tipo "media" não suportado | Usar rotas específicas (image, audio, etc.) |
| 2 | `/send/image` | Timeout DNS | Configurar proxy ou usar base64 |
| 3 | `/messages/revoke` | Limitações do whatsmeow | Verificar formato MessageID |
| 4 | `/poll/results` | Funcionalidade não implementada | Implementar coleta via eventos |

---

## 🔍 Análise Detalhada dos Erros

### 1. Erro: Tipo de Mídia "media" Não Suportado

**Rota**: `POST /sessions/:sessionId/messages/send/media`

**Erro observado**: `"invalid request: unsupported message type: media"`

**Causa raiz**:
No arquivo `internal/infra/wameow/manager.go` linha 1186:
```go
default:
    return nil, fmt.Errorf("unsupported message type: %s", messageType)
```

O switch statement não inclui o caso "media", apenas tipos específicos como "text", "image", "audio", etc.

**Solução**:
- **Opção 1**: Remover a rota `/send/media` e usar apenas rotas específicas
- **Opção 2**: Implementar lógica para detectar tipo de mídia automaticamente baseado no MIME type
- **Opção 3**: Mapear "media" para "document" como fallback

### 2. Erro: Timeout DNS ao Baixar Imagens

**Rota**: `POST /sessions/:sessionId/messages/send/image`

**Erro observado**: `"failed to download file from URL: Get \"https://via.placeholder.com/300x200.png\": dial tcp: lookup via.placeholder.com on 168.63.129.16:53: i/o timeout"`

**Causa raiz**:
No arquivo `internal/domain/message/service.go` linha 129:
```go
resp, err := client.Do(req)
if err != nil {
    return nil, fmt.Errorf("failed to download file from URL: %w", err)
}
```

O ambiente não tem acesso à internet ou há problemas de DNS.

**Soluções**:
1. **Configurar proxy HTTP** no MediaProcessor
2. **Implementar suporte a base64** para mídias
3. **Usar URLs locais** para testes
4. **Configurar DNS alternativo**

### 3. Erro: Falha na Revogação de Mensagens

**Rota**: `POST /sessions/:sessionId/messages/revoke`

**Erro observado**: `"Failed to revoke message"`

**Causa raiz**:
Baseado na documentação do whatsmeow e no código em `internal/infra/wameow/client.go` linha 1605:
```go
message := c.client.BuildRevoke(jid, jid, messageID)
```

**Possíveis causas**:
1. **Formato do MessageID**: Deve ser exatamente como retornado pelo WhatsApp
2. **Limitação de tempo**: WhatsApp permite revogação apenas dentro de ~68 minutos
3. **Permissões**: Só é possível revogar mensagens próprias
4. **JID incorreto**: O JID deve corresponder exatamente ao chat original

**Soluções**:
1. Validar formato do MessageID antes de enviar
2. Adicionar verificação de tempo desde o envio
3. Melhorar tratamento de erros específicos do whatsmeow

### 4. Erro: Resultados de Enquete Não Implementados

**Rota**: `GET /sessions/:sessionId/messages/poll/:messageId/results`

**Erro observado**: `"Failed to get poll results"`

**Causa raiz**:
No arquivo `internal/app/message/usecase.go` linha 169:
```go
return &GetPollResultsResponse{
    // ...
}, fmt.Errorf("poll results collection not yet implemented - requires event handling")
```

**Análise**:
O whatsmeow não fornece um método direto para obter resultados de enquetes. Os votos são recebidos via eventos `DecryptPollVote`.

**Solução**:
Implementar sistema de coleta de votos via event handlers:

```go
// Exemplo de implementação necessária
func (c *WameowClient) setupPollEventHandlers() {
    c.client.AddEventHandler(func(evt *events.Message) {
        if evt.Message.GetPollUpdateMessage() != nil {
            // Processar voto da enquete
            c.processPollVote(evt)
        }
    })
}
```

---

## 🛠️ Implementações Sugeridas

### 1. Suporte a Base64 para Mídias

```go
// Adicionar ao MediaProcessor
func (mp *MediaProcessor) processBase64(data string) (*ProcessedMedia, error) {
    // Extrair MIME type do header base64
    parts := strings.Split(data, ",")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid base64 format")
    }
    
    header := parts[0] // data:image/jpeg;base64
    content := parts[1]
    
    // Decodificar base64
    decoded, err := base64.StdEncoding.DecodeString(content)
    if err != nil {
        return nil, fmt.Errorf("failed to decode base64: %w", err)
    }
    
    // Criar arquivo temporário
    tempFile, err := os.CreateTemp(mp.tempDir, "base64-media-*")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    
    // Escrever dados
    if _, err := tempFile.Write(decoded); err != nil {
        return nil, fmt.Errorf("failed to write data: %w", err)
    }
    
    return &ProcessedMedia{
        FilePath: tempFile.Name(),
        MimeType: extractMimeType(header),
        FileSize: int64(len(decoded)),
        Cleanup: func() error { return os.Remove(tempFile.Name()) },
    }, nil
}
```

### 2. Melhor Tratamento de Erros de Revogação

```go
func (c *WameowClient) RevokeMessage(ctx context.Context, to, messageID string) error {
    // Validar formato do messageID
    if !isValidMessageID(messageID) {
        return fmt.Errorf("invalid message ID format: %s", messageID)
    }
    
    // Verificar se a mensagem não é muito antiga
    if isMessageTooOld(messageID) {
        return fmt.Errorf("message is too old to be revoked (>68 minutes)")
    }
    
    // Tentar revogar
    message := c.client.BuildRevoke(jid, jid, messageID)
    _, err = c.client.SendMessage(ctx, jid, message)
    
    if err != nil {
        // Tratar erros específicos do whatsmeow
        if strings.Contains(err.Error(), "not-authorized") {
            return fmt.Errorf("not authorized to revoke this message")
        }
        if strings.Contains(err.Error(), "too-old") {
            return fmt.Errorf("message is too old to be revoked")
        }
        return fmt.Errorf("failed to revoke message: %w", err)
    }
    
    return nil
}
```

### 3. Sistema de Coleta de Votos de Enquetes

```go
type PollVoteCollector struct {
    votes map[string][]PollVote // messageID -> votes
    mutex sync.RWMutex
}

func (pvc *PollVoteCollector) AddVote(pollMessageID string, vote PollVote) {
    pvc.mutex.Lock()
    defer pvc.mutex.Unlock()
    
    if pvc.votes[pollMessageID] == nil {
        pvc.votes[pollMessageID] = make([]PollVote, 0)
    }
    
    pvc.votes[pollMessageID] = append(pvc.votes[pollMessageID], vote)
}

func (pvc *PollVoteCollector) GetResults(pollMessageID string) *PollResults {
    pvc.mutex.RLock()
    defer pvc.mutex.RUnlock()
    
    votes := pvc.votes[pollMessageID]
    // Processar votos e retornar resultados agregados
    return aggregateVotes(votes)
}
```

---

## 📊 Prioridades de Implementação

1. **Alta**: Suporte a base64 para mídias (resolve problema de conectividade)
2. **Alta**: Melhor tratamento de erros de revogação
3. **Média**: Sistema de coleta de votos de enquetes
4. **Baixa**: Remoção ou correção da rota `/send/media`

---

## 🧪 Testes Recomendados

### Para Mídias Base64:
```bash
curl -X POST "http://localhost:8080/sessions/{sessionId}/messages/send/image" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "559981769536@s.whatsapp.net",
    "file": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
    "caption": "Teste base64"
  }'
```

### Para Revogação:
```bash
# Primeiro enviar uma mensagem
RESPONSE=$(curl -X POST "..." -d '{"to": "...", "body": "Teste"}')
MESSAGE_ID=$(echo $RESPONSE | jq -r '.data.id')

# Depois tentar revogar imediatamente
curl -X POST "http://localhost:8080/sessions/{sessionId}/messages/revoke" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d "{\"to\": \"559981769536@s.whatsapp.net\", \"messageId\": \"$MESSAGE_ID\"}"
```

---

**Última atualização**: 2025-09-27 21:15 UTC
