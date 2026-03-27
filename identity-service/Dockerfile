# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./server/

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/www ./www

EXPOSE 3031 9301

CMD ["./server", "grpc-gw-server", "--port1", "9301", "--port2", "3031", "--grpc-endpoint", ":9301"]
