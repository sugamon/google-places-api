BIN = $(abspath ./bin)

.PHONY: help
help: ## list of operations
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: generate
generate: ## run generate command
	go generate ./...

.PHONY: test
test: ## run test command
	go test ./... -v

.PHONY: cov
cov: ## run test command with coverage
	go test -covermode=atomic

.PHONY: race
race: ## run race command
	go test --race

gotests=$(BIN)/gotests
$(gotests):
	GOBIN=$(BIN) go get -u github.com/cweill/gotests/...

generate-test: $(gotests)
	$(gotests) -w -all ./
