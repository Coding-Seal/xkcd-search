BINARY_NAME=xkcd
PORT=8090

.PHONY: build
build:
	@echo Building...
	@go mod tidy
	CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o build/${BINARY_NAME} ./cmd/xkcd

.PHONY: download 
download:
	@echo Downloading dependencies...
	go mod download

.PHONY: format
format:
	@gofumpt -w .
	@wsl -fix ./...
.PHONY: sec
sec:
	@govulncheck
	@trivy fs .
.PHONY: lint
lint:
	@echo Linting...
	@golangci-lint run
.PHONY: bench
bench:
	@echo Running benchmark ...
	@go test -bench=. ./...
.PHONY: test
test:
	@echo Running tests ...
	@go test -race -coverprofile build/cover.out ./...
	# @go tool cover -html=test/out/cover.out
.PHONY: e2e
e2e: build
	@echo Running e2e tests...
	@build/${BINARY_NAME} -p=${PORT} 2>&1 1>/dev/null &
	@sleep 10s;
	@python3 test/e2e/update.py ${PORT};
	@python3 test/e2e/pics.py ${PORT};
	@kill $$(lsof -t -i:${PORT})
.PHONY: web
web:
	@echo Building...
	@go mod tidy
	go build -o build/web-server ./cmd/web
