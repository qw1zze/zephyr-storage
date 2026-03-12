# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:latest

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]
