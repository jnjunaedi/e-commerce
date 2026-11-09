FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./order_service/cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /order-service .

EXPOSE 50052

CMD ["./order-service"]