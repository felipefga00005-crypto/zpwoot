# 🌐 Exemplos de Send Text com Reply em Diferentes Linguagens

## 📱 JavaScript/Node.js

### Usando Fetch API
```javascript
// Mensagem simples
async function sendSimpleText(sessionId, to, text) {
  const response = await fetch(`http://localhost:8080/sessions/${sessionId}/messages/send/text`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      to: to,
      text: text
    })
  });
  
  return await response.json();
}

// Mensagem com reply
async function sendTextWithReply(sessionId, to, text, messageId, participant = null) {
  const payload = {
    to: to,
    text: text,
    replyTo: {
      messageId: messageId
    }
  };
  
  // Adicionar participant se fornecido (para grupos)
  if (participant) {
    payload.replyTo.participant = participant;
  }
  
  const response = await fetch(`http://localhost:8080/sessions/${sessionId}/messages/send/text`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload)
  });
  
  return await response.json();
}

// Exemplos de uso
(async () => {
  try {
    // Mensagem simples
    const result1 = await sendSimpleText('mySession', '5511987654321@s.whatsapp.net', 'Olá!');
    console.log('Mensagem enviada:', result1);
    
    // Reply individual
    const result2 = await sendTextWithReply(
      'mySession', 
      '5511987654321@s.whatsapp.net', 
      'Obrigado pela mensagem!',
      '3EB0C431C26A1916E07A'
    );
    console.log('Reply enviado:', result2);
    
    // Reply em grupo
    const result3 = await sendTextWithReply(
      'mySession',
      '120363025343298765@g.us',
      'Concordo!',
      '3EB0C431C26A1916E07A',
      '5511987654321@s.whatsapp.net'
    );
    console.log('Reply em grupo enviado:', result3);
    
  } catch (error) {
    console.error('Erro:', error);
  }
})();
```

### Usando Axios
```javascript
const axios = require('axios');

class WhatsAppTextSender {
  constructor(baseUrl, sessionId) {
    this.baseUrl = baseUrl;
    this.sessionId = sessionId;
    this.client = axios.create({
      baseURL: baseUrl,
      headers: {
        'Content-Type': 'application/json'
      }
    });
  }
  
  async sendText(to, text, replyTo = null) {
    const payload = { to, text };
    
    if (replyTo) {
      payload.replyTo = replyTo;
    }
    
    try {
      const response = await this.client.post(
        `/sessions/${this.sessionId}/messages/send/text`,
        payload
      );
      return response.data;
    } catch (error) {
      throw new Error(`Erro ao enviar mensagem: ${error.response?.data?.error || error.message}`);
    }
  }
}

// Uso
const sender = new WhatsAppTextSender('http://localhost:8080', 'mySession');

// Mensagem simples
sender.sendText('5511987654321@s.whatsapp.net', 'Olá!')
  .then(result => console.log('Sucesso:', result))
  .catch(error => console.error('Erro:', error));

// Com reply
sender.sendText(
  '5511987654321@s.whatsapp.net', 
  'Esta é uma resposta!',
  { messageId: '3EB0C431C26A1916E07A' }
).then(result => console.log('Reply enviado:', result));
```

## 🐍 Python

### Usando requests
```python
import requests
import json
from typing import Optional, Dict, Any

class WhatsAppTextSender:
    def __init__(self, base_url: str, session_id: str):
        self.base_url = base_url
        self.session_id = session_id
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})
    
    def send_text(self, to: str, text: str, reply_to: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        """
        Envia mensagem de texto com reply opcional
        
        Args:
            to: Destinatário (JID)
            text: Texto da mensagem
            reply_to: Dict com messageId e participant (opcional)
        
        Returns:
            Resposta da API
        """
        payload = {
            'to': to,
            'text': text
        }
        
        if reply_to:
            payload['replyTo'] = reply_to
        
        url = f"{self.base_url}/sessions/{self.session_id}/messages/send/text"
        
        try:
            response = self.session.post(url, json=payload)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f"Erro ao enviar mensagem: {e}")

# Exemplos de uso
if __name__ == "__main__":
    sender = WhatsAppTextSender('http://localhost:8080', 'mySession')
    
    try:
        # Mensagem simples
        result1 = sender.send_text('5511987654321@s.whatsapp.net', 'Olá do Python!')
        print(f"Mensagem enviada: {result1}")
        
        # Reply individual
        result2 = sender.send_text(
            '5511987654321@s.whatsapp.net',
            'Obrigado pela mensagem!',
            {'messageId': '3EB0C431C26A1916E07A'}
        )
        print(f"Reply enviado: {result2}")
        
        # Reply em grupo
        result3 = sender.send_text(
            '120363025343298765@g.us',
            'Concordo com você!',
            {
                'messageId': '3EB0C431C26A1916E07A',
                'participant': '5511987654321@s.whatsapp.net'
            }
        )
        print(f"Reply em grupo: {result3}")
        
    except Exception as e:
        print(f"Erro: {e}")
```

### Usando aiohttp (Async)
```python
import aiohttp
import asyncio
from typing import Optional, Dict, Any

class AsyncWhatsAppTextSender:
    def __init__(self, base_url: str, session_id: str):
        self.base_url = base_url
        self.session_id = session_id
    
    async def send_text(self, to: str, text: str, reply_to: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        payload = {'to': to, 'text': text}
        
        if reply_to:
            payload['replyTo'] = reply_to
        
        url = f"{self.base_url}/sessions/{self.session_id}/messages/send/text"
        
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=payload) as response:
                if response.status != 200:
                    raise Exception(f"Erro HTTP {response.status}: {await response.text()}")
                return await response.json()

# Uso async
async def main():
    sender = AsyncWhatsAppTextSender('http://localhost:8080', 'mySession')
    
    # Enviar múltiplas mensagens em paralelo
    tasks = [
        sender.send_text('5511987654321@s.whatsapp.net', 'Mensagem 1'),
        sender.send_text('5511987654321@s.whatsapp.net', 'Mensagem 2'),
        sender.send_text('5511987654321@s.whatsapp.net', 'Reply', {'messageId': '3EB0C431C26A1916E07A'})
    ]
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    for i, result in enumerate(results):
        if isinstance(result, Exception):
            print(f"Erro na mensagem {i+1}: {result}")
        else:
            print(f"Mensagem {i+1} enviada: {result}")

# Executar
# asyncio.run(main())
```

## ☕ Java

### Usando OkHttp
```java
import okhttp3.*;
import com.google.gson.Gson;
import com.google.gson.JsonObject;
import java.io.IOException;
import java.util.concurrent.TimeUnit;

public class WhatsAppTextSender {
    private final OkHttpClient client;
    private final String baseUrl;
    private final String sessionId;
    private final Gson gson;
    
    public WhatsAppTextSender(String baseUrl, String sessionId) {
        this.baseUrl = baseUrl;
        this.sessionId = sessionId;
        this.gson = new Gson();
        this.client = new OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build();
    }
    
    public class ReplyTo {
        public String messageId;
        public String participant;
        
        public ReplyTo(String messageId) {
            this.messageId = messageId;
        }
        
        public ReplyTo(String messageId, String participant) {
            this.messageId = messageId;
            this.participant = participant;
        }
    }
    
    public class TextMessageRequest {
        public String to;
        public String text;
        public ReplyTo replyTo;
        
        public TextMessageRequest(String to, String text) {
            this.to = to;
            this.text = text;
        }
        
        public TextMessageRequest(String to, String text, ReplyTo replyTo) {
            this.to = to;
            this.text = text;
            this.replyTo = replyTo;
        }
    }
    
    public JsonObject sendText(String to, String text, ReplyTo replyTo) throws IOException {
        TextMessageRequest request = new TextMessageRequest(to, text, replyTo);
        String json = gson.toJson(request);
        
        RequestBody body = RequestBody.create(
            json, 
            MediaType.get("application/json; charset=utf-8")
        );
        
        Request httpRequest = new Request.Builder()
            .url(baseUrl + "/sessions/" + sessionId + "/messages/send/text")
            .post(body)
            .build();
        
        try (Response response = client.newCall(httpRequest).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Erro HTTP: " + response.code() + " - " + response.body().string());
            }
            
            return gson.fromJson(response.body().string(), JsonObject.class);
        }
    }
    
    public JsonObject sendText(String to, String text) throws IOException {
        return sendText(to, text, null);
    }
    
    // Exemplo de uso
    public static void main(String[] args) {
        WhatsAppTextSender sender = new WhatsAppTextSender("http://localhost:8080", "mySession");
        
        try {
            // Mensagem simples
            JsonObject result1 = sender.sendText("5511987654321@s.whatsapp.net", "Olá do Java!");
            System.out.println("Mensagem enviada: " + result1);
            
            // Reply individual
            ReplyTo reply = new ReplyTo("3EB0C431C26A1916E07A");
            JsonObject result2 = sender.sendText("5511987654321@s.whatsapp.net", "Obrigado!", reply);
            System.out.println("Reply enviado: " + result2);
            
            // Reply em grupo
            ReplyTo groupReply = new ReplyTo("3EB0C431C26A1916E07A", "5511987654321@s.whatsapp.net");
            JsonObject result3 = sender.sendText("120363025343298765@g.us", "Concordo!", groupReply);
            System.out.println("Reply em grupo: " + result3);
            
        } catch (IOException e) {
            System.err.println("Erro: " + e.getMessage());
        }
    }
}
```

## 🐘 PHP

### Usando cURL
```php
<?php

class WhatsAppTextSender {
    private $baseUrl;
    private $sessionId;
    
    public function __construct($baseUrl, $sessionId) {
        $this->baseUrl = $baseUrl;
        $this->sessionId = $sessionId;
    }
    
    public function sendText($to, $text, $replyTo = null) {
        $payload = [
            'to' => $to,
            'text' => $text
        ];
        
        if ($replyTo !== null) {
            $payload['replyTo'] = $replyTo;
        }
        
        $url = $this->baseUrl . "/sessions/" . $this->sessionId . "/messages/send/text";
        
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json'
        ]);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, 30);
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        
        if (curl_errno($ch)) {
            throw new Exception('Erro cURL: ' . curl_error($ch));
        }
        
        curl_close($ch);
        
        if ($httpCode !== 200) {
            throw new Exception("Erro HTTP $httpCode: $response");
        }
        
        return json_decode($response, true);
    }
}

// Exemplos de uso
try {
    $sender = new WhatsAppTextSender('http://localhost:8080', 'mySession');
    
    // Mensagem simples
    $result1 = $sender->sendText('5511987654321@s.whatsapp.net', 'Olá do PHP!');
    echo "Mensagem enviada: " . json_encode($result1) . "\n";
    
    // Reply individual
    $result2 = $sender->sendText(
        '5511987654321@s.whatsapp.net',
        'Obrigado pela mensagem!',
        ['messageId' => '3EB0C431C26A1916E07A']
    );
    echo "Reply enviado: " . json_encode($result2) . "\n";
    
    // Reply em grupo
    $result3 = $sender->sendText(
        '120363025343298765@g.us',
        'Concordo!',
        [
            'messageId' => '3EB0C431C26A1916E07A',
            'participant' => '5511987654321@s.whatsapp.net'
        ]
    );
    echo "Reply em grupo: " . json_encode($result3) . "\n";
    
} catch (Exception $e) {
    echo "Erro: " . $e->getMessage() . "\n";
}
?>
```

## 🔧 Bash/Shell Script

```bash
#!/bin/bash

# Função para enviar texto simples
send_text() {
    local session_id=$1
    local to=$2
    local text=$3
    
    curl -s -X POST "http://localhost:8080/sessions/$session_id/messages/send/text" \
        -H "Content-Type: application/json" \
        -d "{
            \"to\": \"$to\",
            \"text\": \"$text\"
        }"
}

# Função para enviar texto com reply
send_text_with_reply() {
    local session_id=$1
    local to=$2
    local text=$3
    local message_id=$4
    local participant=$5
    
    local payload="{
        \"to\": \"$to\",
        \"text\": \"$text\",
        \"replyTo\": {
            \"messageId\": \"$message_id\""
    
    if [ -n "$participant" ]; then
        payload="$payload,
            \"participant\": \"$participant\""
    fi
    
    payload="$payload
        }
    }"
    
    curl -s -X POST "http://localhost:8080/sessions/$session_id/messages/send/text" \
        -H "Content-Type: application/json" \
        -d "$payload"
}

# Exemplos de uso
echo "Enviando mensagem simples..."
send_text "mySession" "5511987654321@s.whatsapp.net" "Olá do Bash!"

echo "Enviando reply..."
send_text_with_reply "mySession" "5511987654321@s.whatsapp.net" "Obrigado!" "3EB0C431C26A1916E07A"

echo "Enviando reply em grupo..."
send_text_with_reply "mySession" "120363025343298765@g.us" "Concordo!" "3EB0C431C26A1916E07A" "5511987654321@s.whatsapp.net"
```

## 📝 Notas Importantes

1. **Tratamento de Erros**: Sempre implemente tratamento adequado de erros
2. **Timeouts**: Configure timeouts apropriados para as requisições
3. **Rate Limiting**: Respeite os limites de taxa do WhatsApp
4. **Logs**: Implemente logging para debugging
5. **Validação**: Valide os dados antes de enviar
6. **Segurança**: Nunca exponha tokens ou credenciais no código
