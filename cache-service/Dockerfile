# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o cache-service .

# ---- Run Stage ----
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/cache-service .

EXPOSE 8080

CMD ["./cache-service"]
