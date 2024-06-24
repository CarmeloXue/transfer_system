.PHONY: all build clean run

all: build

build:
	@echo "Building api server..."
	@go build -o bin/api cmd/api/main.go

clean:
	@echo "Cleaning build..."
	@rm -rf bin/*

run:
	@golangci-lint run
	@docker-compose up --build

down: 
	@docker-compose down

clean-volume:
	@docker volume rm transfer_system_db_data

test:
	@go test ./...

lint:
	@golangci-lint run