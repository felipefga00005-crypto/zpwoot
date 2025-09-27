# WhatsApp Message Sending API

Este documento descreve como usar a API de envio de mensagens do zpwoot para enviar diferentes tipos de mensagens através do WhatsApp.

## 🔄 Padronização de Campos de Texto

**A partir da versão atual, todos os endpoints foram padronizados para usar `body` como campo principal para conteúdo textual, seguindo o padrão do WhatsApp.**

### Mudanças Importantes:
- ✅ **Padrão unificado**: Use `body` para conteúdo textual (alinhado com WhatsApp)
- ⚠️ **Deprecated**: O campo `text` ainda é aceito mas será removido em versões futuras
- 🔄 **Compatibilidade**: Durante o período de transição, ambos os campos são aceitos
- 📝 **Prioridade**: Se ambos `body` e `text` forem fornecidos, `body` tem prioridade

### Campos Afetados:
- `body` / `text` - Conteúdo de mensagens de texto
- `newBody` / `newText` - Novo conteúdo ao editar mensagens
- Mensagens de botão e lista também seguem o mesmo padrão

## Endpoints Disponíveis

### Endpoint Principal (Genérico)
```
POST /sessions/{sessionId}/messages/send
```

### Endpoints Específicos por Tipo
```
POST /sessions/{sessionId}/messages/send/text      - Mensagens de texto
POST /sessions/{sessionId}/messages/send/media     - Mídia genérica
POST /sessions/{sessionId}/messages/send/image     - Imagens
POST /sessions/{sessionId}/messages/send/audio     - Áudios
POST /sessions/{sessionId}/messages/send/video     - Vídeos
POST /sessions/{sessionId}/messages/send/document  - Documentos
POST /sessions/{sessionId}/messages/send/sticker   - Stickers
POST /sessions/{sessionId}/messages/send/location  - Localização
POST /sessions/{sessionId}/messages/send/contact   - Contatos
POST /sessions/{sessionId}/messages/send/button    - Mensagens com botões
POST /sessions/{sessionId}/messages/send/list      - Mensagens com lista
POST /sessions/{sessionId}/messages/send/reaction  - Reações
POST /sessions/{sessionId}/messages/send/presence  - Presença (typing, online, etc.)
POST /sessions/{sessionId}/messages/edit           - Editar mensagem
POST /sessions/{sessionId}/messages/delete         - Deletar mensagem
```

## Tipos de Mensagem Suportados

A API suporta os seguintes tipos de mensagem:
- `text` - Mensagens de texto simples
- `image` - Imagens (JPEG, PNG, etc.)
- `audio` - Arquivos de áudio (OGG, MP3, etc.)
- `video` - Vídeos (MP4, etc.)
- `document` - Documentos (PDF, DOC, etc.)
- `sticker` - Stickers (WebP)
- `location` - Localização geográfica
- `contact` - Contatos (vCard)
- `button` - Mensagens com botões interativos (placeholder)
- `list` - Mensagens com lista interativa (placeholder)
- `reaction` - Reações a mensagens (placeholder)
- `presence` - Status de presença (placeholder)

## Formato da Requisição

### Estrutura Base

```json
{
  "to": "5511999999999@s.whatsapp.net",
  "type": "text|image|audio|video|document|location|contact",
  "body": "Texto da mensagem (padrão WhatsApp)",
  "text": "Texto da mensagem (deprecated - use 'body')",
  "caption": "Legenda para mídia (opcional)",
  "file": "URL ou base64 do arquivo (para mídia)",
  "filename": "Nome do arquivo (para documentos)",
  "latitude": 0.0,
  "longitude": 0.0,
  "contactName": "Nome do contato",
  "contactPhone": "Telefone do contato"
}
```

> **⚠️ Aviso de Compatibilidade**: O campo `text` está deprecated. Use `body` para novos desenvolvimentos seguindo o padrão WhatsApp. Ambos os campos são aceitos durante o período de transição.

## Exemplos de Uso

### 1. Mensagem de Texto

**Formato Recomendado (usando `body` - padrão WhatsApp):**
```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "text",
    "body": "Olá! Esta é uma mensagem de teste."
  }'
```

**Formato Legacy (usando `text` - deprecated):**
```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "text",
    "text": "Olá! Esta é uma mensagem de teste."
  }'
```

### 2. Imagem via URL

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "image",
    "file": "https://example.com/image.jpg",
    "caption": "Veja esta imagem!"
  }'
```

### 3. Imagem via Base64

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "image",
    "file": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD...",
    "caption": "Imagem enviada via base64"
  }'
```

### 4. Áudio

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "audio",
    "file": "https://example.com/audio.ogg"
  }'
```

### 5. Vídeo

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "video",
    "file": "https://example.com/video.mp4",
    "caption": "Confira este vídeo!"
  }'
```

### 6. Documento

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "document",
    "file": "https://example.com/document.pdf",
    "filename": "relatorio.pdf"
  }'
```

### 7. Localização

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "type": "location",
    "latitude": -23.5505,
    "longitude": -46.6333,
    "text": "São Paulo, SP, Brasil"
  }'
```

### 8. Contato

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send/contact \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "contactName": "João Silva",
    "contactPhone": "+5511888888888"
  }'
```

### 9. Sticker

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send/sticker \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "file": "https://example.com/sticker.webp"
  }'
```

### 10. Mensagem com Botões (Placeholder)

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send/button \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "body": "Escolha uma opção:",
    "buttons": [
      {"id": "1", "text": "Opção 1"},
      {"id": "2", "text": "Opção 2"},
      {"id": "3", "text": "Opção 3"}
    ]
  }'
```

### 11. Reação (Placeholder)

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send/reaction \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "messageId": "3EB0C431C26A1916E07E",
    "reaction": "👍"
  }'
```

### 12. Presença (Placeholder)

```bash
curl -X POST http://localhost:8080/sessions/mySession/messages/send/presence \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "presence": "typing"
  }'
```

## Resposta da API

### Sucesso (200 OK)

```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "messageId": "3EB0C431C26A1916E07E",
    "status": "sent",
    "timestamp": "2025-09-26T17:30:00Z"
  }
}
```

### Erro (400/404/500)

```json
{
  "success": false,
  "message": "Error message",
  "error": "Detailed error description"
}
```

## Códigos de Status

- `200 OK` - Mensagem enviada com sucesso
- `400 Bad Request` - Dados da requisição inválidos
- `404 Not Found` - Sessão não encontrada
- `500 Internal Server Error` - Erro interno do servidor

## Validações

### Campos Obrigatórios

- `to`: Número do destinatário no formato internacional
- `type`: Tipo da mensagem

### Validações por Tipo

- **text**: Requer `body`
- **image/video**: Requer `file`, `caption` é opcional
- **audio**: Requer `file`
- **document**: Requer `file` e `filename`
- **location**: Requer `latitude`, `longitude`, `body` é opcional
- **contact**: Requer `contactName` e `contactPhone`

## Formatos de Arquivo Suportados

### Imagens
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

### Áudio
- OGG (.ogg)
- MP3 (.mp3)
- AAC (.aac)
- AMR (.amr)

### Vídeo
- MP4 (.mp4)
- 3GP (.3gp)
- MOV (.mov)
- AVI (.avi)

### Documentos
- PDF (.pdf)
- DOC/DOCX (.doc, .docx)
- XLS/XLSX (.xls, .xlsx)
- PPT/PPTX (.ppt, .pptx)
- TXT (.txt)
- E outros formatos comuns

## Limitações

- Tamanho máximo de arquivo: 64MB
- Formatos de base64: Devem incluir o prefixo `data:mime/type;base64,`
- URLs: Devem ser acessíveis publicamente
- Números de telefone: Devem estar no formato internacional com código do país

## Tratamento de Erros

A API retorna erros específicos para diferentes situações:

- **Session not found**: Sessão não existe ou não está ativa
- **Session not connected**: Sessão não está conectada ao WhatsApp
- **Invalid request**: Dados da requisição são inválidos
- **Failed to process media**: Erro ao processar arquivo de mídia
- **Unsupported message type**: Tipo de mensagem não suportado

## Exemplo de Integração em JavaScript

```javascript
async function sendWhatsAppMessage(sessionId, messageData) {
  try {
    const response = await fetch(`/sessions/${sessionId}/messages/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
      },
      body: JSON.stringify(messageData)
    });

    const result = await response.json();
    
    if (result.success) {
      console.log('Message sent:', result.data.messageId);
      return result.data;
    } else {
      console.error('Error sending message:', result.message);
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
}

// Exemplo de uso
sendWhatsAppMessage('mySession', {
  to: '5511999999999@s.whatsapp.net',
  type: 'text',
  body: 'Hello from JavaScript!'
});
```

## Exemplo de Integração em Python

```python
import requests
import json

def send_whatsapp_message(session_id, message_data):
    url = f"http://localhost:8080/sessions/{session_id}/messages/send"
    headers = {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
    }
    
    try:
        response = requests.post(url, headers=headers, json=message_data)
        result = response.json()
        
        if result.get('success'):
            print(f"Message sent: {result['data']['messageId']}")
            return result['data']
        else:
            print(f"Error sending message: {result.get('message')}")
            raise Exception(result.get('message'))
            
    except requests.exceptions.RequestException as e:
        print(f"Network error: {e}")
        raise

# Exemplo de uso
send_whatsapp_message('mySession', {
    'to': '5511999999999@s.whatsapp.net',
    'type': 'text',
    'body': 'Hello from Python!'
})
```
