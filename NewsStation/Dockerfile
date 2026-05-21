FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o geonews ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/geonews .
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./geonews"]