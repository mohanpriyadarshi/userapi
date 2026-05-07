#!/bin/bash

BASE="http://localhost:8080"

echo "=== CREATE USERS ==="
IDS=()
USERS=(
  '{"name":"Alice","email":"alice@example.com","password":"pass123"}'
  '{"name":"Bob","email":"bob@example.com","password":"pass456"}'
  '{"name":"Charlie","email":"charlie@example.com","password":"pass789"}'
)

for user in "${USERS[@]}"; do
  response=$(curl -s -X POST "$BASE/users" \
    -H "Content-Type: application/json" \
    -d "$user")
  echo "$response" | python3 -m json.tool
  id=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
  IDS+=("$id")
done

echo ""
echo "=== LIST ALL USERS ==="
curl -s "$BASE/users" | python3 -m json.tool

echo ""
echo "=== DELETE USERS ==="
for id in "${IDS[@]}"; do
  echo "Deleting user id=$id"
  curl -s -X DELETE "$BASE/users/$id" | python3 -m json.tool
done

echo ""
echo "=== LIST USERS AFTER DELETE ==="
curl -s "$BASE/users" | python3 -m json.tool
