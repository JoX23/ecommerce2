#!/bin/bash
set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== E-Commerce API Demo ==="

echo -e "\n1. Register user"
REGISTER=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","name":"Demo User","password":"password123"}')
echo $REGISTER | python3 -m json.tool 2>/dev/null || echo $REGISTER

TOKEN=$(echo $REGISTER | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "")

echo -e "\n2. Login"
LOGIN=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"password123"}')
echo $LOGIN | python3 -m json.tool 2>/dev/null || echo $LOGIN

TOKEN=$(echo $LOGIN | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo $TOKEN)

echo -e "\n3. Get current user profile"
curl -s $BASE_URL/auth/me \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo -e "\n4. Create a product"
PRODUCT=$(curl -s -X POST $BASE_URL/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"sku":"PROD-001","name":"Test Product","price":29.99,"stock":100,"status":"published","description":"A great product"}')
echo $PRODUCT | python3 -m json.tool 2>/dev/null || echo $PRODUCT

PRODUCT_ID=$(echo $PRODUCT | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")

echo -e "\n5. List published products"
curl -s $BASE_URL/products | python3 -m json.tool 2>/dev/null

echo -e "\n6. Get product detail"
if [ -n "$PRODUCT_ID" ]; then
  curl -s $BASE_URL/products/$PRODUCT_ID | python3 -m json.tool 2>/dev/null
fi

echo -e "\n7. Create order"
if [ -n "$PRODUCT_ID" ]; then
  ORDER=$(curl -s -X POST $BASE_URL/orders \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"items\":[{\"productId\":\"$PRODUCT_ID\",\"qty\":2}]}")
  echo $ORDER | python3 -m json.tool 2>/dev/null || echo $ORDER
  ORDER_ID=$(echo $ORDER | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")
fi

echo -e "\n8. List user orders"
curl -s $BASE_URL/orders \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo -e "\n9. Get order detail"
if [ -n "$ORDER_ID" ]; then
  curl -s $BASE_URL/orders/$ORDER_ID \
    -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
fi

echo -e "\n=== Demo Complete ==="
