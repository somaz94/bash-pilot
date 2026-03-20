.PHONY: build clean test test-unit cover cover-html fmt vet install demo demo-clean demo-all branch pr help

APP_NAME=bash-pilot
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/somaz94/bash-pilot/cmd/cli.Version=$(VERSION) -X github.com/somaz94/bash-pilot/cmd/cli.GitCommit=$(GIT_COMMIT) -X github.com/somaz94/bash-pilot/cmd/cli.BuildDate=$(BUILD_DATE)"

## Build

build: ## Build the binary
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/

clean: ## Remove build artifacts and coverage files
	rm -rf bin/ coverage.out coverage.html

## Test

test: test-unit ## Run unit tests (alias)

test-unit: ## Run unit tests with coverage
	go test ./... -v -race -cover

## Coverage

cover: ## Generate coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html: cover ## Open coverage report in browser
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

## Quality

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

## Install

install: build ## Install to /usr/local/bin
	cp bin/$(APP_NAME) /usr/local/bin/$(APP_NAME)

## Demo

demo: build ## Run demo (create temp SSH config, test all commands)
	@./scripts/demo.sh

demo-clean: ## Clean up demo resources
	@./scripts/demo-clean.sh

demo-all: demo demo-clean ## Run demo and clean up automatically

## Workflow

branch: ## Create feature branch (usage: make branch name=git-module)
	@if [ -z "$(name)" ]; then echo "Usage: make branch name=<module-name>"; exit 1; fi
	git checkout main
	git pull origin main
	git checkout -b feat/$(name)
	@echo "\033[32m✓ Branch feat/$(name) created\033[0m"

pr: ## Run tests, push, and create PR (usage: make pr title="Add git module")
	@if [ -z "$(title)" ]; then echo "Usage: make pr title=\"PR title\""; exit 1; fi
	go test ./... -race -cover
	go vet ./...
	git push -u origin $$(git branch --show-current)
	gh pr create --title "$(title)" --body "## Summary"$$'\n\n'"Branch: $$(git branch --show-current)"$$'\n\n'"## Test plan"$$'\n\n'"- [ ] Unit tests pass"$$'\n'"- [ ] Coverage maintained"
	@echo "\033[32m✓ PR created\033[0m"

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
