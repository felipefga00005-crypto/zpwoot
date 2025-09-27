#!/bin/bash

# Script para testar a compatibilidade dos endpoints padronizados
# Testa tanto o campo 'body' (padrão) quanto 'text' (deprecated)

BASE_URL="http://localhost:8080"
SESSION_ID="testSession"
API_KEY="dev-api-key-12345"
TO="559981769536@s.whatsapp.net"

echo "🧪 Testando Compatibilidade dos Endpoints de Mensagem"
echo "=================================================="

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

# Usando 'body' (padrão)
make_request "/sessions/$SESSION_ID/messages/send" \
    '{"to":"'$TO'","type":"text","body":"Teste usando body (padrão)"}' \
    "Mensagem de texto usando 'body'"

# Usando 'text' (deprecated)
make_request "/sessions/$SESSION_ID/messages/send" \
    '{"to":"'$TO'","type":"text","text":"Teste usando text (deprecated)"}' \
    "Mensagem de texto usando 'text'"

# Usando ambos (body deve ter prioridade)
make_request "/sessions/$SESSION_ID/messages/send" \
    '{"to":"'$TO'","type":"text","body":"Body tem prioridade","text":"Text ignorado"}' \
    "Mensagem com ambos campos (body deve ter prioridade)"

# 2. Teste do endpoint específico /messages/send/text
echo ""
echo "2️⃣ Endpoint Específico (/messages/send/text)"

# Usando 'body' (padrão)
make_request "/sessions/$SESSION_ID/messages/send/text" \
    '{"to":"'$TO'","body":"Teste endpoint específico com body"}' \
    "Endpoint específico usando 'body'"

# Usando 'text' (deprecated)
make_request "/sessions/$SESSION_ID/messages/send/text" \
    '{"to":"'$TO'","text":"Teste endpoint específico com text"}' \
    "Endpoint específico usando 'text'"

# 3. Teste de mensagens de botão
echo ""
echo "3️⃣ Mensagens de Botão (/messages/send/button)"

# Usando 'body' (padrão)
make_request "/sessions/$SESSION_ID/messages/send/button" \
    '{"to":"'$TO'","body":"Escolha uma opção:","buttons":[{"id":"1","text":"Opção 1"}]}' \
    "Mensagem de botão usando 'body'"

# Usando 'text' (deprecated)
make_request "/sessions/$SESSION_ID/messages/send/button" \
    '{"to":"'$TO'","text":"Escolha uma opção:","buttons":[{"id":"1","text":"Opção 1"}]}' \
    "Mensagem de botão usando 'text'"

# 4. Teste de mensagens de lista
echo ""
echo "4️⃣ Mensagens de Lista (/messages/send/list)"

# Usando 'body' (padrão)
make_request "/sessions/$SESSION_ID/messages/send/list" \
    '{"to":"'$TO'","body":"Selecione uma opção:","buttonText":"Ver opções","sections":[{"title":"Seção 1","rows":[{"id":"1","title":"Item 1"}]}]}' \
    "Mensagem de lista usando 'body'"

# 5. Teste de edição de mensagem
echo ""
echo "5️⃣ Edição de Mensagem (/messages/edit)"

# Usando 'newBody' (padrão)
make_request "/sessions/$SESSION_ID/messages/edit" \
    '{"to":"'$TO'","messageId":"fake-id","newBody":"Mensagem editada com newBody"}' \
    "Edição usando 'newBody'"

# Usando 'newText' (deprecated)
make_request "/sessions/$SESSION_ID/messages/edit" \
    '{"to":"'$TO'","messageId":"fake-id","newText":"Mensagem editada com newText"}' \
    "Edição usando 'newText'"

echo ""
echo "✅ Testes de compatibilidade concluídos!"
echo ""
echo "📋 Resumo da Padronização:"
echo "- Campo padrão: 'body' (alinhado com WhatsApp)"
echo "- Campo deprecated: 'text'"
echo "- Prioridade: 'body' > 'text'"
echo "- Compatibilidade: Ambos aceitos durante transição"
echo "- Endpoints afetados: /send, /send/text, /send/button, /send/list, /edit"
