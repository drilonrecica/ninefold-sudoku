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
	pnpm --filter @ninefold/web test

lint:
	@test -z "$$(gofmt -l apps/server)" || { echo "Go files require formatting:"; gofmt -l apps/server; exit 1; }
	cd apps/server && go vet ./...
	pnpm format:check
	pnpm --filter @ninefold/web lint
	pnpm --filter @ninefold/web check

generate:
	@echo "Contract generation is not implemented until Phase 2." >&2
	@exit 1

migrate:
	@echo "Database migrations are not implemented until Phase 4." >&2
	@exit 1

puzzles:
	@echo "Puzzle catalog tooling is not implemented until Phase 3." >&2
	@exit 1

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
