FROM golang:1.23.0-alpine AS builder

LABEL maintainer="neznaika1337@list.ru"

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]