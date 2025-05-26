FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gophkeeper ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /gophkeeper .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
COPY docs/ ./docs/

EXPOSE 8080

CMD ["./gophkeeper"] 