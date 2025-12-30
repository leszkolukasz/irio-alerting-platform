SERVICE_ID=$1
SERVICE_STATUS=$2

DATA_BASE64=$(echo -n "{\"service_id\": $SERVICE_ID}" | base64)

curl -X POST http://localhost:8085/v1/projects/test-project/topics/service-$SERVICE_STATUS:publish \
-H "Content-Type: application/json" \
-d '{
    "messages": [
        {
            "data": "'$DATA_BASE64'"
        }
    ]
}'

