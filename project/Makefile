BROKER_APP_BINARY=brokerApp
SEARCH_APP_BINARY=searchApp

postgres-up:
	@echo Starting Postgresql
	docker-compose up -d postgres
	@echo Postgresql started

postgres_up_build:
	@echo Starting Postgresql
	docker-compose up -d postgres
	@echo Postgresql started

up:
	@echo Starting docker containers...
	docker-compose up -d
	@echo Docker containers started

up_build: down build_broker_service build_search_service
	@echo Starting docker containers...
	docker-compose up -d --build
	@echo Docker containers started

down:
	@echo Stopping containers
	docker-compose down
	@echo Container stopped

search_service_start:
	@echo Starting search searvice
	cd ../search-service && DSN="host=localhost user=admin password=admin dbname=advance-search port=5432 sslmode=disable" PORT="6001" OLLAMA_EMBEDDING_URL="http://localhost:11434/api/embeddings" go run ./cmd/api/
	@echo Stopping search searvice

broker_start: 
	@echo Starting search searvice
	cd ../broker-service && PORT="6000" go run ./cmd/api/
	@echo Stopping search searvice

build_search_service:
	@echo Building search searvice
	cd ../search-service && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ${SEARCH_APP_BINARY} ./cmd/api/
	@echo Search service built

build_broker_service:
	@echo Building broker searvice
	cd ../broker-service && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ${BROKER_APP_BINARY} ./cmd/api/
	@echo Search broker built