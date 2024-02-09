NAME ?= prldevops
export PACKAGE_NAME ?= $(NAME)
export VERSION=$(shell ./scripts/workflows/current-version.sh -f VERSION)

COBERTURA = cobertura

GOX = gox

GOLANGCI_LINT = golangci-lint

GOSEC = gosec

SWAG = swag

DEVELOPMENT_TOOLS = $(GOX) $(COBERTURA) $(GOLANGCI_LINT) $(SWAG)
SECURITY_TOOLS = $(GOSEC)
GET_CURRENT_LINT_CONTAINER = $(shell docker ps -a -q -f "name=$(PACKAGE_NAME)-linter")

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
	@cd src && go test -coverprofile coverage.txt -covermode count -v ./...
	@gocov convert coverage.txt | gocov-xml >../"$COVERAGE_DIR"/cobertura-coverage.xml
	@rm coverage.txt

.PHONY: lint
lint:
	@echo "Running linter..."
ifeq ($(GET_CURRENT_LINT_CONTAINER),)
	@echo "Linter container does not exist, creating it..."
	@-docker run --name $(PACKAGE_NAME)-linter -e RUN_LOCAL=true -e VALIDATE_ALL_CODEBASE=true -e VALIDATE_JSCPD=false -e CREATE_LOG_FILE=true -v .:/tmp/lint ghcr.io/super-linter/super-linter:slim-v5
else
	@echo "Linter container already exists, starting it..."
	@-docker start $(PACKAGE_NAME)-linter --attach
endif
	@docker cp $(PACKAGE_NAME)-linter:/tmp/lint/super-linter.log ./super-linter.log
	@echo "Linter report saved to super-linter.log"
	@echo "Linter finished."

.PHONY: security
security-check:
	@echo "Running Security Check..."
	@scripts/security-check -d ./src

.PHONY: build
build:
	@echo "Building..."
ifneq ("$(wildcard out)","")
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && go build -o ../out/binaries/$(PACKAGE_NAME)
	@echo "Build finished."

.PHONY: clean
clean:
	@echo "Cleaning..."
ifneq ("$(wildcard bin)","")
	@echo "Removing bin directory..."
	@rm -rf bin
endif
ifneq ("$(wildcard out)","")
	@echo "Removing out directory..."
	@rm -rf out
endif
ifneq ("$(wildcard coverage)","")
	@echo "Removing coverage directory..."
	@rm -rf out
endif
ifneq ("$(wildcard tmp)","")
	@echo "Removing tmp directory..."
	@rm -rf out
endif
	@echo "Clean finished."

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