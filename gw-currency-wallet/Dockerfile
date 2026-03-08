FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY proto-exchange ./proto-exchange
COPY gw-currency-wallet ./gw-currency-wallet

WORKDIR /build/gw-currency-wallet

RUN go mod tidy && go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN $(go env GOPATH)/bin/swag init -g cmd/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/gw-currency-wallet/main .
COPY --from=builder /build/gw-currency-wallet/docs ./docs

EXPOSE 8080

CMD ["./main"]