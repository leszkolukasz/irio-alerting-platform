#!/bin/bash
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'


echo -e "${BLUE}[INFO] Loading configuration...${NC}"
if [ -f .env ]; then
    set -a
    source .env
    set +a
else
    echo -e "${RED}[WARN] .env file not found.${NC}"
fi

echo -e "${BLUE}[INFO] Cleaning up stale Go processes...${NC}"
pkill -f "go-build" || true
pkill -f "exe/api" || true
pkill -f "exe/scheduler" || true
pkill -f "exe/worker" || true
pkill -f "exe/incident-manager" || true
pkill -f "exe/logger" || true
sleep 1

echo -e "${BLUE}[INFO] Resetting infrastructure (fresh state)...${NC}"
docker compose down -v 
docker compose up -d db redis pubsub firebase

wait_for_port() {
    local port=$1
    local name=$2
    local retries=30
    echo -n "   Waiting for $name (port $port)... "
    while ! nc -z localhost $port; do
        sleep 1
        retries=$((retries-1))
        if [ $retries -le 0 ]; then
            echo -e "${RED}TIMEOUT${NC}"
            exit 1
        fi
    done
    echo -e "${GREEN}OK${NC}"
}

wait_for_port 5432 "Postgres"
wait_for_port 6379 "Redis"
wait_for_port 8085 "Pub/Sub Emulator"
wait_for_port 8086 "Firestore Emulator"

echo -e "${BLUE}[INFO] Starting microservices...${NC}"
cd backend

echo -e "${BLUE}[INFO] Initializing Pub/Sub topics...${NC}"
PROJECT_ID=${PROJECT_ID} \
PUBSUB_EMULATOR_HOST=localhost:8085 \
go run cmd/setup_infra/main.go

run_service() {
    svc=$1
    var_name=${svc//-/_}
    
    PROJECT_ID=${PROJECT_ID} \
    API_HOST=127.0.0.1 \
    RPC_PORT=9092 \
    PUBSUB_EMULATOR_HOST=localhost:8085 \
    FIRESTORE_EMULATOR_HOST=localhost:8086 \
    go run services/$svc/main.go > "/tmp/alerting_$svc.log" 2>&1 &
    eval "${var_name}_PID=$!"
}

run_service "logger"
run_service "worker"
run_service "incident-manager"

echo "   Starting API service..."
PROJECT_ID=${PROJECT_ID} \
API_HOST=127.0.0.1 \
RPC_PORT=9092 \
PUBSUB_EMULATOR_HOST=localhost:8085 \
FIRESTORE_EMULATOR_HOST=localhost:8086 \
go run services/api/main.go > /tmp/alerting_api.log 2>&1 &
api_PID=$!

echo -n "   Waiting for API HTTP (8080)..."
MAX_WAIT=120
count=0
while ! curl -s "http://localhost:8080/health" > /dev/null; do
    sleep 1
    count=$((count+1))
    if [ $count -ge $MAX_WAIT ]; then
        echo -e "\n${RED}[ERROR] API failed to start.${NC}"
        tail -n 20 /tmp/alerting_api.log
        kill $api_PID $worker_PID $incident_manager_PID $logger_PID 2>/dev/null
        exit 1
    fi
    echo -n "."
done
echo -e "${GREEN} OK${NC}"

wait_for_port 9092 "API gRPC"


echo "   Starting services dependent on API..."
run_service "incident-manager"
run_service "scheduler"

cleanup() {
    echo -e "\n${BLUE}[INFO] Teardown: Stopping services...${NC}"
    kill $api_PID $scheduler_PID $worker_PID $incident_manager_PID $logger_PID 2>/dev/null || true
}
trap cleanup EXIT


echo -e "${BLUE}[INFO] Running E2E Integration Tests...${NC}"
if go test -v -count=1 tests/e2e/main_test.go; then
    echo -e "${GREEN}[SUCCESS] All tests passed!${NC}"
else
    echo -e "${RED}[FAILURE] Tests failed.${NC}"
    echo "--- SERVICE LOGS (TAIL) ---"
    echo ">> API:"
    tail -n 10 /tmp/alerting_api.log
    echo ">> SCHEDULER:"
    tail -n 10 /tmp/alerting_scheduler.log
    echo ">> WORKER:"
    tail -n 10 /tmp/alerting_worker.log
    echo ">> INCIDENT MANAGER:"
    tail -n 10 /tmp/alerting_incident-manager.log
    exit 1
fi