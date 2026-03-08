FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY proto-exchange ./proto-exchange
COPY gw-exchanger/go.mod gw-exchanger/go.sum ./gw-exchanger/
COPY gw-exchanger/cmd ./gw-exchanger/cmd
COPY gw-exchanger/internal ./gw-exchanger/internal
COPY gw-exchanger/migrations ./gw-exchanger/migrations

WORKDIR /build/gw-exchanger

RUN go mod tidy && go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/gw-exchanger/main .

EXPOSE 50051

CMD ["./main"]