.PHONY: all build clean run

all: build

build:
	@echo "Building api server..."
	@go build -o bin/api cmd/api/main.go

clean:
	@echo "Cleaning build..."
	@rm -rf bin/*

run:
	@docker-compose up --build

down: 
	@docker-compose down