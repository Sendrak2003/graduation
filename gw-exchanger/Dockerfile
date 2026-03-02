FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./

COPY . .

RUN go mod tidy && go mod download && go mod verify

RUN GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

EXPOSE 50051

CMD ["./main"]
