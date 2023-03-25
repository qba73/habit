.PHONY: dox test vet check cover tidy

default: dox

dox: ## Run tests with gotestdox
	@gotestdox  >/dev/null 2>&1 || go install github.com/bitfield/gotestdox/cmd/gotestdox@latest ;
	gotestdox

test: ## Run tests
	go test -race ./...

vet: ## Run go vet
	go vet ./...

check: ## Run staticcheck analyzer
	@staticcheck -version >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@2022.1;
	staticcheck ./...

cover: ## Run unit tests and generate test coverage report
	go test -race -v ./... -count=1 -cover -covermode=atomic -coverprofile=coverage.out
	go tool cover -html coverage.out

tidy: ## Run go mod tidy
	go mod tidy

