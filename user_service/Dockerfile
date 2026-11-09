FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./user_service/cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /user-service .

EXPOSE 50056

CMD ["./user-service"]