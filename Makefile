NAME ?= prldevops
export PACKAGE_NAME ?= $(NAME)
export DOCKER_PACKAGE_NAME ?= "prl-devops-service"
ifeq ($(OS),Windows_NT)
	export VERSION:=$(shell type VERSION)
else
	export VERSION:=$(shell cat VERSION)
	export BUILD_ID:=$(shell date +%s)
	export SHORT_VERSION:=$(shell echo $(VERSION) | cut -d'.' -f1,2)
	export BUILD_VERSION:=$(shell echo $(SHORT_VERSION).$(BUILD_ID))
endif

COBERTURA = cobertura

GOX = gox

GOLANGCI_LINT = golangci-lint

GOSEC = gosec

SWAG = swag

START_SUPER_LINTER_CONTAINER = start_super_linter_container

DEVELOPMENT_TOOLS = $(GOX) $(COBERTURA) $(GOLANGCI_LINT) $(SWAG) $(bundler)
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
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && go build -o ../out/binaries/$(PACKAGE_NAME)
	@echo "Build finished."

.PHONY: build-canary
build-canary:
	@echo "Building Canary..."
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && go build -o ../out/binaries/$(PACKAGE_NAME) -ldflags="-X 'github.com/Parallels/prl-devops-service/config.canaryBuildFlag=true'"
	@echo "Build finished."

.PHONY: build-beta
build-beta:
	@echo "Building Beta..."
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && go build -o ../out/binaries/$(PACKAGE_NAME) -ldflags="-X 'github.com/Parallels/prl-devops-service/config.betaBuildFlag=true'"
	@echo "Build finished."

.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "Building..."
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o ../out/binaries/$(PACKAGE_NAME)-linux-amd64
	@echo "Build finished."

.PHONY: build-windows-amd64
build-windows-amd64:
	@echo "Building..."
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && CGO_ENABLED=0 GOOS="windows" GOARCH="amd64" go build -o ../out/binaries/$(PACKAGE_NAME)-linux-amd64
	@echo "Build finished."

.PHONY: build-alpine
build-alpine:
	@echo "Building..."
ifeq ($(wildcard ./out/.*),)
	@echo "Creating out directory..."
	@mkdir out
	@mkdir out/binaries
endif
	@cd src && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ../out/binaries/$(PACKAGE_NAME)-alpine
	@echo "Build finished."

.PHONY: push-alpha-container
push-alpha-container:
	@echo "Building $(BUILD_VERSION) Alpha Container..."
	@docker build --platform linux/amd64,linux/arm64 \
		-t cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-alpha \
		-t cjlapao/$(DOCKER_PACKAGE_NAME):latest-alpha \
		--build-arg VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_ENV=canary \
		--build-arg OS=linux \
		--build-arg ARCHITECTURE=amd64 \
		-f Dockerfile .
	@echo "Pushing $(BUILD_VERSION) Container..."
	@echo "Pushing cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-alpha tag..."
	@docker push cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-alpha
	@echo "Pushing cjlapao/$(DOCKER_PACKAGE_NAME):latest-alpha tag..."
	@docker push cjlapao/$(DOCKER_PACKAGE_NAME):latest-alpha
	@echo "Build finished. Pushed to cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)_alpha and cjlapao/$(DOCKER_PACKAGE_NAME):latest_alpha."

.PHONY: clean-alpha
clean-alpha-container:
	@echo "Removing all alpha versions from Docker Hub..."
	@./.github/workflow_scripts/remove-docker-images.sh rm --filter '.*alpha.*$$' 
	@echo "All alpha versions removed."

.PHONY: push-beta-container
push-beta-container:
	@echo "Building $(BUILD_VERSION) Beta Container..."
	@docker build --platform linux/amd64,linux/arm64 \
		-t cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-beta \
		-t cjlapao/$(DOCKER_PACKAGE_NAME):unstable \
		--build-arg VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_ENV=canary \
		--build-arg OS=linux \
		--build-arg ARCHITECTURE=amd64 \
		-f Dockerfile .
	@echo "Pushing $(BUILD_VERSION) Container..."
	@echo "Pushing cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-beta tag..."
	@docker push cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-beta
	@echo "Pushing cjlapao/$(DOCKER_PACKAGE_NAME):unstable tag..."
	@docker push cjlapao/$(DOCKER_PACKAGE_NAME):unstable
	@echo "Build finished. Pushed to cjlapao/$(DOCKER_PACKAGE_NAME):$(BUILD_VERSION)-beta and cjlapao/$(DOCKER_PACKAGE_NAME):unstable."

.PHONY: clean-beta
clean-beta-container:
	@echo "Removing all beta versions from Docker Hub..."
	@./.github/workflow_scripts/remove-docker-images.sh rm --filter ".*beta.*$$" 
	@echo "All alpha versions removed."

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
	@cd src && swag fmt
	@cd src && swag init -g main.go

.PHONY: deps
deps: $(DEVELOPMENT_TOOLS) $(SECURITY_TOOLS)

.PHONY: release-check
release-check: test lint generate-swagger coverage security-check

.PHONY: build-docs
build-docs:
	@echo "Building Documentation..."
	@cd docs && bundle exec jekyll build ./docs

.PHONY: serve-docs
serve-docs:
	@echo "Serving Documentation..."
	@cd docs && bundle exec jekyll serve ./docs

.PHONY: build-helm-chart
build-helm-chart:
	@echo "Building Helm Chart..."
	@helm package ./helm -d ./docs/charts
	@helm repo index ./docs/charts --url https://parallels.github.io/prl-devops-service/charts --merge ./docs/charts/index.yaml

sign-macos-app:
	@echo "Building App App..."
	@if [ -f ./out/binaries/$(PACKAGE_NAME) ]; then \
		echo "Removing previous MacOS App..."; \
		rm ./out/binaries/$(PACKAGE_NAME); \
	fi
	@if [ -f ./out/binaries/$(PACKAGE_NAME).zip ]; then \
		echo "Removing previous MacOS App Bundle..."; \
		rm ./out/binaries/$(PACKAGE_NAME).zip; \
	fi
	@make build
	@echo "Signing MacOS App..."
	@codesign --force --deep --strict --verbose --options=runtime,library --sign "Developer ID Application: Carlos Lapao (KXLX56937Q)" --entitlements prldevops.entitlements ./out/binaries/$(PACKAGE_NAME)
	@echo "MacOS App signed."
	@echo "Notarizing MacOS App..."
	@cd ./out/binaries && ditto -c -k --sequesterRsrc $(PACKAGE_NAME) $(PACKAGE_NAME).zip
	@xcrun notarytool submit ./out/binaries/$(PACKAGE_NAME).zip --keychain-profile "notary-credentials" --wait
	@echo "Verifying MacOS App..."
	@codesign --verify --verbose ./out/binaries/$(PACKAGE_NAME)
	@spctl -t open --context context:primary-signature -a -vvv ./out/binaries/$(PACKAGE_NAME)

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

$(bundler):
	@echo "Installing bundler..."
	@bundler install

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