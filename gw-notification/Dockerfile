FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY gw-notification ./gw-notification

WORKDIR /build/gw-notification

RUN go mod tidy
RUN go mod download
RUN go mod verify

RUN GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/gw-notification/main .

CMD ["./main"]