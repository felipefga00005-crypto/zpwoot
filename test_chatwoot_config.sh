#!/bin/bash

# Configuração do Chatwoot para a sessão my-session
# Execute este script para configurar a integração Chatwoot

SESSION_ID="73b71a13-fb97-4ff9-bed0-16f28137c255"
BASE_URL="http://localhost:8080"
CHATWOOT_URL="http://127.0.0.1:3001"
CHATWOOT_TOKEN="tMo1XbvJWXM1V4BcJtFjkMXr"
ACCOUNT_ID="1"

echo "🚀 Configurando Chatwoot para sessão: my-session"
echo "📱 Número: 554988989314"
echo "🔗 Chatwoot URL: $CHATWOOT_URL"
echo "🎯 Account ID: $ACCOUNT_ID"
echo ""

# Criar configuração com auto-criação de inbox
echo "📝 Criando configuração Chatwoot..."
response=$(curl -s -X POST "$BASE_URL/sessions/$SESSION_ID/chatwoot/set" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "url": "'$CHATWOOT_URL'",
    "token": "'$CHATWOOT_TOKEN'",
    "accountId": "'$ACCOUNT_ID'",
    "autoCreate": true,
    "inboxName": "WhatsApp zpwoot - my-session",
    "enabled": true,
    "signMsg": false,
    "signDelimiter": "\n\n",
    "reopenConv": true,
    "convPending": false,
    "importContacts": false,
    "importMessages": false,
    "importDays": 60,
    "mergeBrazil": true,
    "number": "554988989314"
  }')

echo "📄 Resposta da API:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# Verificar se a configuração foi criada
echo "🔍 Verificando configuração criada..."
config_response=$(curl -s -X GET "$BASE_URL/sessions/$SESSION_ID/chatwoot/find" \
  -H "X-API-Key: your-api-key-here")

echo "📋 Configuração atual:"
echo "$config_response" | jq '.' 2>/dev/null || echo "$config_response"
echo ""

# Verificar status da sessão
echo "📊 Status da sessão:"
session_response=$(curl -s -X GET "$BASE_URL/sessions/$SESSION_ID" \
  -H "X-API-Key: your-api-key-here")

echo "$session_response" | jq '.data.session | {id, name, deviceJid, isConnected}' 2>/dev/null || echo "$session_response"

echo ""
echo "✅ Configuração concluída!"
echo "🎉 O Chatwoot agora está integrado com a sessão WhatsApp"
echo ""
echo "📌 Próximos passos:"
echo "   1. Verifique se o inbox foi criado no Chatwoot"
echo "   2. Teste enviando uma mensagem para o número WhatsApp"
echo "   3. Verifique se a conversa aparece no Chatwoot"
