# Catálogo de Rotas de Mensagens da API

Este documento cataloga todas as rotas do `messageHandler` para facilitar os testes com uma sessão já conectada.

## 📋 Índice de Rotas

### 1. Envio de Mensagens Básicas
| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| POST | `/sessions/:sessionId/messages/send/text` | SendText | Envio de mensagem de texto |
| POST | `/sessions/:sessionId/messages/send/media` | SendMedia | Envio de mídia genérica |
| POST | `/sessions/:sessionId/messages/send/image` | SendImage | Envio de imagem |
| POST | `/sessions/:sessionId/messages/send/audio` | SendAudio | Envio de áudio |
| POST | `/sessions/:sessionId/messages/send/video` | SendVideo | Envio de vídeo |
| POST | `/sessions/:sessionId/messages/send/document` | SendDocument | Envio de documento |
| POST | `/sessions/:sessionId/messages/send/sticker` | SendSticker | Envio de sticker |

### 2. Mensagens Interativas
| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| POST | `/sessions/:sessionId/messages/send/button` | SendButtonMessage | Mensagem com botões |
| POST | `/sessions/:sessionId/messages/send/contact` | SendContact | Envio de contato(s) |
| POST | `/sessions/:sessionId/messages/send/list` | SendListMessage | Mensagem com lista |
| POST | `/sessions/:sessionId/messages/send/location` | SendLocation | Envio de localização |
| POST | `/sessions/:sessionId/messages/send/poll` | SendPoll | Criação de enquete |

### 3. Reações e Presença
| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| POST | `/sessions/:sessionId/messages/send/reaction` | SendReaction | Envio de reação |
| POST | `/sessions/:sessionId/messages/send/presence` | SendPresence | Envio de presença |

### 4. Gerenciamento de Mensagens
| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| POST | `/sessions/:sessionId/messages/edit` | EditMessage | Edição de mensagem |
| POST | `/sessions/:sessionId/messages/mark-read` | MarkAsRead | Marcar como lida |
| POST | `/sessions/:sessionId/messages/revoke` | RevokeMessage | Revogar mensagem |
| GET | `/sessions/:sessionId/messages/poll/:messageId/results` | GetPollResults | Resultados da enquete |

## 🔧 Configuração para Testes

### Pré-requisitos
- **SessionId**: Uma sessão válida e conectada
- **Número de destino**: `559981769536` (conforme especificado)
- **ContextInfo**: Para testes de reply, usar estrutura adequada

### Estrutura Base do ContextInfo
```json
{
  "contextInfo": {
    "stanzaId": "ABCD1234abcd",
    "participant": "5511999999999@s.whatsapp.net"
  }
}
```

## 📝 Observações Importantes

1. **SessionId obrigatório**: Todas as rotas exigem um `sessionId` válido e ativo
2. **Número de destino**: Usar `559981769536` para todos os testes
3. **ContextInfo**: Incluir quando testar funcionalidade de reply
4. **Validação**: Cada rota tem validações específicas de campos obrigatórios
5. **Respostas**: Todas retornam estruturas padronizadas com timestamp e status

## 🚀 Próximos Passos

1. Criar arquivo de registro de testes em tabela markdown
2. Implementar testes sistemáticos para cada rota
3. Documentar payloads de exemplo para cada tipo de mensagem
4. Registrar sucessos e erros encontrados

---

**Nota**: Este catálogo serve como referência rápida. Para payloads detalhados e exemplos de uso, consulte a documentação específica de cada endpoint.
