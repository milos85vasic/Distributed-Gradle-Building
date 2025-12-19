# Agent Instructions for Distributed Gradle Building

## Build/Test Commands
- **All Go tests**: `cd go && go test ./... -v`
- **Single Go test**: `cd go && go test -run TestName -v`
- **Unit tests**: `cd tests && make test-unit`
- **Integration tests**: `cd tests && make test-integration`
- **All tests**: `cd tests && make test-all`
- **Lint Go**: `cd go && gofmt -d . && go vet ./...`
- **Lint shell**: `shellcheck scripts/*.sh tests/**/*.sh`

## Code Style Guidelines
- **Go**: Use `gofmt` formatting, PascalCase for types/structs, snake_case for JSON tags
- **Shell**: `set -e` for error handling, snake_case functions, stderr for errors
- **Imports**: Group standard library, then third-party, then local packages
- **Error handling**: Return errors explicitly, use `if err != nil` pattern
- **Naming**: Descriptive names, avoid abbreviations except common ones (e.g., `id`, `url`)
- **Types**: Use struct tags for JSON serialization, include validation where needed
- **Comments**: Document exported functions/types, keep implementation comments minimal