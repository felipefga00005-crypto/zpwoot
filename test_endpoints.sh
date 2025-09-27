#!/bin/bash

# Script para testar os endpoints padronizados de mensagem
# Todos os endpoints usam 'body' como campo padrão

BASE_URL="http://localhost:8080"
SESSION_ID="testSession"
API_KEY="dev-api-key-12345"
TO="559981769536@s.whatsapp.net"

echo "🧪 Testando Endpoints de Mensagem Padronizados"
echo "=============================================="

# Função para fazer requisições
make_request() {
    local endpoint="$1"
    local data="$2"
    local description="$3"
    
    echo ""
    echo "📝 Testando: $description"
    echo "Endpoint: $endpoint"
    echo "Payload: $data"
    
    response=$(curl -s -X POST "$BASE_URL$endpoint" \
        -H "Content-Type: application/json" \
        -H "Authorization: $API_KEY" \
        -d "$data")
    
    echo "Resposta: $response"
    echo "---"
}

# 1. Teste do endpoint genérico /messages/send
echo "1️⃣ Endpoint Genérico (/messages/send)"

make_request "/sessions/$SESSION_ID/messages/send" \
    '{"to":"'$TO'","type":"text","body":"Teste endpoint genérico"}' \
    "Mensagem de texto genérica"

# 2. Teste do endpoint específico /messages/send/text
echo ""
echo "2️⃣ Endpoint Específico (/messages/send/text)"

make_request "/sessions/$SESSION_ID/messages/send/text" \
    '{"to":"'$TO'","body":"Teste endpoint específico"}' \
    "Endpoint específico de texto"

# 3. Teste de mensagens de botão
echo ""
echo "3️⃣ Mensagens de Botão (/messages/send/button)"

make_request "/sessions/$SESSION_ID/messages/send/button" \
    '{"to":"'$TO'","body":"Escolha uma opção:","buttons":[{"id":"1","text":"Opção 1"},{"id":"2","text":"Opção 2"}]}' \
    "Mensagem de botão"

# 4. Teste de mensagens de lista
echo ""
echo "4️⃣ Mensagens de Lista (/messages/send/list)"

make_request "/sessions/$SESSION_ID/messages/send/list" \
    '{"to":"'$TO'","body":"Selecione uma opção:","buttonText":"Ver opções","sections":[{"title":"Seção 1","rows":[{"id":"1","title":"Item 1","description":"Descrição do item 1"}]}]}' \
    "Mensagem de lista"

# 5. Teste de edição de mensagem
echo ""
echo "5️⃣ Edição de Mensagem (/messages/edit)"

make_request "/sessions/$SESSION_ID/messages/edit" \
    '{"to":"'$TO'","messageId":"fake-id","newBody":"Mensagem editada"}' \
    "Edição de mensagem"

# 6. Teste de mídia
echo ""
echo "6️⃣ Mensagens de Mídia"

make_request "/sessions/$SESSION_ID/messages/send/image" \
    '{"to":"'$TO'","file":"https://example.com/image.jpg","caption":"Legenda da imagem"}' \
    "Mensagem de imagem"

make_request "/sessions/$SESSION_ID/messages/send/document" \
    '{"to":"'$TO'","file":"https://example.com/doc.pdf","filename":"documento.pdf","caption":"Documento anexo"}' \
    "Mensagem de documento"

# 7. Teste de localização
echo ""
echo "7️⃣ Mensagem de Localização"

make_request "/sessions/$SESSION_ID/messages/send/location" \
    '{"to":"'$TO'","latitude":-23.5505,"longitude":-46.6333,"body":"São Paulo, SP"}' \
    "Mensagem de localização"

echo ""
echo "✅ Testes concluídos!"
echo ""
echo "📋 Resumo da Padronização:"
echo "- Campo padrão: 'body' (alinhado com WhatsApp)"
echo "- Sem compatibilidade: apenas 'body' é aceito"
echo "- Endpoints padronizados: /send, /send/text, /send/button, /send/list, /edit"
echo "- Mídia usa 'caption' para legendas"
echo "- Localização usa 'body' para endereço"
