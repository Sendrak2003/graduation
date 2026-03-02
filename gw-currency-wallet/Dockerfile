FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./

COPY . .

RUN go mod tidy && go mod download && go mod verify

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init -g cmd/main.go -o docs

RUN GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
