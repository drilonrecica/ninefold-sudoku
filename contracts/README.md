# Contracts

OpenAPI 3.1 and JSON Schema Draft 2020-12 files are the transport contract sources.

Run `make generate` after changing a schema. Generated Go and TypeScript packages are committed and
must remain transport DTOs; domain packages must not import them.

Pinned generators:

- `oapi-codegen` 2.8.0
- `go-jsonschema` 0.23.1
- `openapi-typescript` 7.13.0
- `json-schema-to-typescript` 15.0.4

## Evolution rules

Additive optional fields and new event types are minor changes. Consumers must ignore unknown event
types only after preserving the event number and requesting a snapshot when authoritative state
cannot be reduced safely.

Removing or renaming a field, changing its meaning or type, tightening an accepted value, or
reinterpreting an existing event is a major protocol change. Breaking changes require a new schema
version and an explicit compatibility window.

Schema changes land with producer handling, consumer handling, regenerated packages, and shared
fixtures in the same phase. Unsupported schema versions are rejected; payload interpretation is
never guessed.
