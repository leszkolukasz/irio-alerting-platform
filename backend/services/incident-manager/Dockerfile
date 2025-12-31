FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common/go.mod common/go.sum ./common/
COPY services/api/go.mod services/api/go.sum ./services/api/
COPY services/logger/go.mod services/logger/go.sum ./services/logger/
COPY services/incident-manager/go.mod services/incident-manager/go.sum ./services/incident-manager/

RUN go mod download

COPY common/ ./common/
COPY services/incident-manager/ ./services/incident-manager/

WORKDIR /app/services/incident-manager
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/incident-manager .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/incident-manager .
COPY --from=builder /app/services/incident-manager/.env .

CMD ["./incident-manager"]