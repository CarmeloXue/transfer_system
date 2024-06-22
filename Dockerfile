# syntax=docker/dockerfile:1

FROM golang:1.20-alpine as builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Build the services
RUN go build -o main_account ./applications/account/main.go
RUN go build -o main_transaction ./applications/transaction/main.go

EXPOSE 8080
EXPOSE 8081

# Run both services in a single container
CMD ["sh", "-c", "./main_account & ./main_transaction"]