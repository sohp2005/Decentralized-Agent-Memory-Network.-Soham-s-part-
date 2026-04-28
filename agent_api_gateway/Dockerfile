# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o agent-gateway .

# ---- Run Stage ----
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/agent-gateway .

EXPOSE 8081
EXPOSE 50051

CMD ["./agent-gateway"]
