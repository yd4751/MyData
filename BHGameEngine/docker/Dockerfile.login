FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o login cmd/login/main.go

EXPOSE 8081

CMD ["./login", "-config", "config/config.toml"]