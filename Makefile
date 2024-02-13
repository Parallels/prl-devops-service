NAME ?= prldevops
export PACKAGE_NAME ?= $(NAME)
ifeq ($(OS),Windows_NT)
	export VERSION=$(shell type VERSION)
else
	export VERSION=$(shell cat VERSION)
endif

COBERTURA = cobertura

GOX = gox

GOLANGCI_LINT = golangci-lint

GOSEC = gosec

SWAG = swag

START_SUPER_LINTER_CONTAINER = start_super_linter_container

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
ifeq ("$(wildcard coverage)","")
	@echo "Creating coverage directory..."
	@mkdir coverage
endif
	@cd src && go test -coverprofile coverage.txt -covermode count -v ./...
	@cd src && gocov convert coverage.txt | gocov-xml >../coverage/cobertura-coverage.xml
	@cd src && rm coverage.txt

.PHONY: lint
lint: $(START_SUPER_LINTER_CONTAINER)
	@echo "Running linter..."
	@docker cp $(PACKAGE_NAME)-linter:/tmp/lint/super-linter.log .
	@echo "Linter report saved to super-linter.log"
	@docker stop $(PACKAGE_NAME)-linter
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

$(START_SUPER_LINTER_CONTAINER):
ifeq ($(OS), Windows_NT)
	$(eval CONTAINER_ID := $(shell  docker ps -a -q -f "name=$(PACKAGE_NAME)-linter"))
	$(eval REPO := $(shell git rev-parse --abbrev-ref HEAD))
	@IF "$(CONTAINER_ID)" EQU "" (\
	docker run --name $(PACKAGE_NAME)-linter -e DEFAULT_BRANCH=main -e RUN_LOCAL=true -e VALIDATE_ALL_CODEBASE=true -e VALIDATE_JSCPD=false -e CREATE_LOG_FILE=true -e VALIDATE_GO=false -v .:/tmp/lint ghcr.io/super-linter/super-linter:latest \
	) \
	ELSE (\
	docker start $(PACKAGE_NAME)-linter --attach \
	);
else
	$(eval CONTAINER_ID := $(shell docker ps -a | grep $(PACKAGE_NAME)-linter | awk '{print $$1}'))
	$(eval REPO := $(shell git rev-parse --abbrev-ref HEAD))
	@if [ -z $(CONTAINER_ID) ]; then \
	echo "Linter container does not exist, creating it..."; \
	docker run --platform linux/amd64 --name $(PACKAGE_NAME)-linter -e DEFAULT_BRANCH=$(REPO) -e RUN_LOCAL=true -e VALIDATE_ALL_CODEBASE=true -e VALIDATE_JSCPD=false -e CREATE_LOG_FILE=true -e VALIDATE_GO=false -v .:/tmp/lint ghcr.io/super-linter/super-linter:slim-latest; \
	else \
	echo "Linter container already exists $(CONTAINER_ID), starting it..."; \
	docker start $(PACKAGE_NAME)-linter --attach; \
	fi
endif