FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/orchestrator ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/frontend ./frontend
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/orchestrator /app/orchestrator

CMD ["/app/orchestrator"]