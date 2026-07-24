.PHONY: dev test lint generate migrate puzzles e2e build tla server-dev web-dev

dev:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) server-dev & \
	$(MAKE) web-dev & \
	wait

server-dev:
	cd apps/server && go run -ldflags="-X main.buildVersion=development" ./cmd/api

web-dev:
	pnpm --filter @ninefold/web dev

test:
	cd apps/server && go test ./...
	cd contracts/generated/go && go test ./...
	pnpm --filter @ninefold/web test

lint:
	@test -z "$$(gofmt -l apps/server)" || { echo "Go files require formatting:"; gofmt -l apps/server; exit 1; }
	cd apps/server && go vet ./...
	cd contracts/generated/go && go vet ./...
	pnpm format:check
	pnpm --filter @ninefold/web lint
	pnpm --filter @ninefold/web check

generate:
	mkdir -p contracts/generated/go/http contracts/generated/go/realtime contracts/generated/go/replay
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 \
		-generate types,skip-prune -package http \
		-o contracts/generated/go/http/types.gen.go contracts/openapi/ninefold.openapi.yaml
	go run github.com/atombender/go-jsonschema@v0.23.1 \
		--only-models --min-sized-ints --struct-name-from-title --tags json \
		-p realtime -o contracts/generated/go/realtime/types.gen.go \
		contracts/websocket/client-message.schema.json \
		contracts/websocket/server-message.schema.json
	go run github.com/atombender/go-jsonschema@v0.23.1 \
		--only-models --min-sized-ints --struct-name-from-title --tags json \
		-p replay -o contracts/generated/go/replay/types.gen.go \
		contracts/replay/replay.schema.json contracts/replay/proof.schema.json
	cd apps/server && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
	pnpm generate:contracts
	gofmt -w contracts/generated/go apps/server/internal/persistence/gen

migrate:
	cd apps/server && go run ./cmd/maintenance migrate

puzzles:
	cd apps/server && go run ./cmd/puzzle-generator \
		-validate internal/puzzle/catalog/catalog.jsonl

e2e:
	@echo "End-to-end tests are not implemented until Phase 10." >&2
	@exit 1

build:
	mkdir -p apps/server/bin
	cd apps/server && go build -trimpath -ldflags="-s -w -X main.buildVersion=development" -o bin/ninefold-api ./cmd/api
	pnpm --filter @ninefold/web build

tla:
	@echo "TLA+ verification is not implemented until Phase 6." >&2
	@exit 1
