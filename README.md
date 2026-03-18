# Go Demo

This is a minimal demo project in Go for learning purposes.

## Local run

```bash
go run .
go test ./...
```

## Pre-commit

Install the git hooks:

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

Run all checks manually:

```bash
pre-commit run --all-files
```

Configured hooks:

- `check-byte-order-marker`
- `check-case-conflict`
- `check-merge-conflict`
- `check-symlinks`
- `check-yaml`
- `end-of-file-fixer`
- `mixed-line-ending`
- `trailing-whitespace`
- `typos`
- `gofmt`
- `go vet ./...`
- `go test ./...`

## GitHub Actions

The workflow file is at `.github/workflows/ci.yml`.

It will run on every push to `main` or `master`, and on every pull request.

Checks included:

- `gofmt`
- `go test -v ./...`
- `go test -v -cover ./...`
- `go test -race ./...`
- `go build ./...`

## Project files

- `.github/workflows/ci.yml`: GitHub Actions CI
- `.github/dependabot.yml`: dependency update automation
- `.editorconfig`: editor formatting defaults
- `Makefile`: common local commands
- `LICENSE`: MIT license

## Common commands

```bash
make fmt
make test
make testv
make cover
make race
make build
make check
make ci
```
