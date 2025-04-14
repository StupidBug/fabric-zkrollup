curl -sS -X GET 'http://127.0.0.1:8080/api/v1/proof/status?hash=ad57366865126e55649ecb23ae1d48887544976efea46a48eb5d85a6eeb4d306' \
    -H "Content-Type: application/json" \
    -d "$json"
echo