curl -X POST http://localhost:8085/v1/projects/test-project/topics/ServiceDown:publish \
-H "Content-Type: application/json" \
-d '{
    "messages": [
        {
            "data": "eyJzZXJ2aWNlX2lkIjogMTIzfQo="
        }
    ]
}'

