FROM golang:1.25 AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common/go.mod common/go.sum ./common/
COPY services/logger/go.mod services/logger/go.sum ./services/logger/

RUN go mod download

COPY common/ ./common/
COPY services/logger/ ./services/logger/

WORKDIR /app/services/logger
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/logger ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/logger .

COPY --from=builder /app/services/logger/.env .

CMD ["./logger"]