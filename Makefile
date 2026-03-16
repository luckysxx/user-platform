.PHONY: init-networks local-infra-up local-infra-down local-run local-run-http local-run-grpc local-test proto-gen docker-up docker-down docker-logs ps health

NETWORK_EXTERNAL = go-net
NETWORK_INTERNAL = platform-internal
COMPOSE_FILES = -f docker-compose-infra.yaml -f docker-compose-service.yaml

init-networks:
	@docker network inspect $(NETWORK_EXTERNAL) >/dev/null 2>&1 || docker network create $(NETWORK_EXTERNAL)
	@docker network inspect $(NETWORK_INTERNAL) >/dev/null 2>&1 || docker network create $(NETWORK_INTERNAL)

local-infra-up: init-networks
	docker compose -f docker-compose-infra.yaml up -d postgres redis

local-infra-down:
	docker compose -f docker-compose-infra.yaml stop postgres redis

local-run:
	go run ./cmd/http

local-run-http:
	go run ./cmd/http

local-run-grpc:
	go run ./cmd/grpc

local-test:
	go test ./...

proto-gen:
	protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/user/user_api.proto proto/auth/auth_api.proto

docker-up: init-networks
	docker compose $(COMPOSE_FILES) up -d --build

docker-down:
	docker compose $(COMPOSE_FILES) down

docker-logs:
	docker compose $(COMPOSE_FILES) logs -f user-http user-grpc postgres redis

ps:
	docker compose $(COMPOSE_FILES) ps

health:
	docker compose $(COMPOSE_FILES) ps --format "table {{.Name}}\t{{.State}}\t{{.Health}}"
	docker compose $(COMPOSE_FILES) logs --tail=40 user-http user-grpc postgres redis
