# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o persistence-service .

# ---- Run Stage ----
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/persistence-service .

# SQLite file will be created inside container
VOLUME ["/data"]

EXPOSE 8090

CMD ["./persistence-service"]
