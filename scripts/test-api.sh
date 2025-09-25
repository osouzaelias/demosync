#!/bin/bash

# Endpoint da API
API_URL="http://localhost:8080/api/v1/payments/capture"

# Dados da requisição
REQUEST_DATA='{
  "transaction_id": "tx_'$(date +%s)'",
  "amount": 100.50,
  "currency": "BRL",
  "metadata": {
    "order_id": "order_12345",
    "customer_id": "cust_6789"
  }
}'

echo "Enviando requisição para $API_URL"
echo "Dados: $REQUEST_DATA"
echo "-------------------------------------"

# Envia a requisição e guarda a resposta
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "$REQUEST_DATA" \
  $API_URL)

echo "Resposta recebida:"
echo $RESPONSE | jq '.'
