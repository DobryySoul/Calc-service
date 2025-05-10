FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/agent ./cmd/agent/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/config/config.yaml /app/config/config.yaml
COPY --from=builder /app/agent /app/agent

CMD ["/app/agent"]