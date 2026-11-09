FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-service ./payment_service/cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /payment-service .

EXPOSE 50053

CMD ["./payment-service"]