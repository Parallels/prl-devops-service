NAME ?= "prldevops"
export PACKAGE_NAME ?= $(NAME)
export VERSION=$(shell ./scripts/workflows/current-version.sh -f VERSION)

COBERTURA = cobertura

GOX = gox

GOLANGCI_LINT = golangci-lint

GOSEC = gosec

SWAG = swag

DEVELOPMENT_TOOLS = $(GOX) $(COBERTURA) $(GOLANGCI_LINT) $(SWAG)

SECURITY_TOOLS = $(GOSEC)

.PHONY: help
help:
  # make version:
	# make test
	# make lint

.PHONY: version
version:
	@echo Version: $(VERSION)

.PHONY: test
test:
	@echo "Running tests..."
	@scripts/test -d ./src

.PHONY: coverage
coverage:
	@echo "Running coverage report..."
	@scripts/coverage -d ./src

.PHONY: lint
lint:
	@echo "Running linter..."
	@scripts/lint

.PHONY: security
security-check:
	@echo "Running Security Check..."
	@scripts/security-check -d ./src

.PHONY: build
build:
	@echo "Building..."
	@scripts/build -d ./src -p $(PACKAGE_NAME)

.PHONY: clean
clean:
	@echo "Cleaning..."
	@scripts/clean

.PHONY: generate-swagger
generate-swagger:
	@echo "Generating Swagger..."
	@scripts/generate-swagger -d ./src

.PHONY: deps
deps: $(DEVELOPMENT_TOOLS) $(SECURITY_TOOLS)

.PHONY: release-check
release-check: test lint generate-swagger coverage security-check
	

$(COBERTURA):
	@echo "Installing cobertura..."
	@go install github.com/axw/gocov/gocov@latest
	@go install github.com/AlekSi/gocov-xml@latest
	@go install github.com/matm/gocov-html/cmd/gocov-html@latest

$(GOX):
	@echo "Installing gox..."
	@go install github.com/mitchellh/gox@latest


$(GOLANGCI_LINT):
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

$(GOSEC):
	@echo "Installing gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest

$(SWAG):
	@echo "Installing swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest