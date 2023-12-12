.DEFAULT_GOAL := help

# AutoDoc
# -------------------------------------------------------------------------
.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.DEFAULT_GOAL := help

.PHONY: gotidy
gotidy: ## Run golangci-lint, goimports and gofmt
	./scripts/golinter.sh

.PHONY: update
update: ## Update all dependencies
	./scripts/update_go_mod.sh