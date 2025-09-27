# 📱 Envio de Mensagem de Texto com Reply/Citação

Este documento demonstra como usar a API para enviar mensagens de texto com citação/reply de mensagens anteriores.

## 🚀 Endpoint

```
POST /sessions/{sessionId}/messages/send/text
```

## 📋 Parâmetros

### Headers
```
Content-Type: application/json
Authorization: Bearer {token} (se configurado)
```

### Path Parameters
- `sessionId` (string): ID da sessão ou nome da sessão

### Request Body

#### Mensagem Simples (sem reply)
```json
{
  "to": "5511987654321@s.whatsapp.net",
  "text": "Olá! Esta é uma mensagem de texto simples."
}
```

#### Mensagem com Reply/Citação
```json
{
  "to": "5511987654321@s.whatsapp.net",
  "text": "Esta é uma resposta à sua mensagem anterior!",
  "replyTo": {
    "messageId": "3EB0C431C26A1916E07A",
    "participant": "5511987654321@s.whatsapp.net"
  }
}
```

## 📝 Campos do Request

### Campos Obrigatórios
- `to` (string): Número do destinatário no formato JID
- `text` (string): Texto da mensagem

### Campos Opcionais para Reply
- `replyTo` (object): Informações da mensagem sendo citada
  - `messageId` (string): ID da mensagem original
  - `participant` (string): Participante (obrigatório para grupos)

## 📤 Response

### Sucesso (200)
```json
{
  "success": true,
  "message": "Text message sent successfully",
  "data": {
    "id": "3EB0C431C26A1916E07A",
    "status": "sent",
    "timestamp": 1704067200
  }
}
```

### Erro (400/404/500)
```json
{
  "success": false,
  "error": "Session not found"
}
```

## 🔧 Exemplos de Uso

### 1. Mensagem Simples

```bash
curl -X POST "http://localhost:8080/sessions/mySession/messages/send/text" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "5511987654321@s.whatsapp.net",
    "text": "Olá! Como você está?"
  }'
```

### 2. Resposta a uma Mensagem (Chat Individual)

```bash
curl -X POST "http://localhost:8080/sessions/mySession/messages/send/text" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "5511987654321@s.whatsapp.net",
    "text": "Obrigado pela sua mensagem!",
    "replyTo": {
      "messageId": "3EB0C431C26A1916E07A"
    }
  }'
```

### 3. Resposta em Grupo

```bash
curl -X POST "http://localhost:8080/sessions/mySession/messages/send/text" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "120363025343298765@g.us",
    "text": "Concordo com você!",
    "replyTo": {
      "messageId": "3EB0C431C26A1916E07A",
      "participant": "5511987654321@s.whatsapp.net"
    }
  }'
```

## 🎯 Casos de Uso

### 1. **Atendimento ao Cliente**
```json
{
  "to": "5511987654321@s.whatsapp.net",
  "text": "Entendi sua solicitação. Vou verificar e retorno em breve.",
  "replyTo": {
    "messageId": "3EB0C431C26A1916E07A"
  }
}
```

### 2. **Confirmação de Pedido**
```json
{
  "to": "5511987654321@s.whatsapp.net",
  "text": "✅ Pedido confirmado! Número: #12345",
  "replyTo": {
    "messageId": "3EB0C431C26A1916E07A"
  }
}
```

### 3. **Resposta em Grupo de Trabalho**
```json
{
  "to": "120363025343298765@g.us",
  "text": "Vou trabalhar nessa tarefa hoje.",
  "replyTo": {
    "messageId": "3EB0C431C26A1916E07A",
    "participant": "5511123456789@s.whatsapp.net"
  }
}
```

## ⚠️ Observações Importantes

### 1. **Formato do JID**
- Chat individual: `5511987654321@s.whatsapp.net`
- Grupo: `120363025343298765@g.us`

### 2. **MessageID**
- Deve ser o ID exato da mensagem original
- Pode ser obtido através de webhooks ou logs

### 3. **Participant em Grupos**
- Obrigatório quando replying em grupos
- Deve ser o JID do autor da mensagem original

### 4. **Limitações**
- Sessão deve estar conectada
- Destinatário deve existir
- MessageID deve ser válido

## 🔍 Troubleshooting

### Erro: "Session not found"
```bash
# Verificar se a sessão existe
curl -X GET "http://localhost:8080/sessions/mySession/info"
```

### Erro: "Session is not connected"
```bash
# Conectar a sessão
curl -X POST "http://localhost:8080/sessions/mySession/connect"
```

### Erro: "Invalid recipient JID"
```bash
# Verificar formato do número
# Correto: 5511987654321@s.whatsapp.net
# Incorreto: +55 11 98765-4321
```

## 📚 Recursos Relacionados

- [Configuração de Sessão](./session_setup.md)
- [Envio de Mídia](./send_media.md)
- [Webhooks](./webhooks.md)
- [Gerenciamento de Grupos](./group_management.md)
