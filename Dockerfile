# Lint
FROM golang:1.20-alpine AS lint
RUN apk add --no-cache make git
WORKDIR /app
COPY . .
# Assuming you use golangci-lint for linting
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN golangci-lint run
RUN make test

# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o bin/api ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/api /usr/local/bin/api
RUN mkdir -p /var/log/api


CMD ["api"]