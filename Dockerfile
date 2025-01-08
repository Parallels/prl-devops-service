# syntax=docker/dockerfile:1
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git coreutils musl-utils ca-certificates curl

WORKDIR /go/src

COPY ./src .

# Using go get.
RUN go get -d -v

# Build the binary.
ARG BUILD_ENV=production
ARG VERSION
ARG OS=linux
ARG ARCHITECTURE=amd64

# Update version in main.go for swagger documentation
RUN sed -i "/@version/c\//\t@version\t\t$VERSION" ./main.go

# Install swag for swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go mod tidy && swag fmt && swag init -g main.go

RUN --mount=type=secret,id=amplitude_api_key \
  export AMPLITUDE_API_KEY=$(cat /run/secrets/amplitude_api_key) && \
  if [ "$BUILD_ENV" = "production" ]; then \
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCHITECTURE go build -ldflags="-s -w -X main.ver=$VERSION -X 'github.com/Parallels/prl-devops-service/constants.AmplitudeApiKey=$AMPLITUDE_API_KEY'" -o /go/bin/prl-devops-service; \
  else \
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCHITECTURE go build -ldflags="-s -w -X main.ver=$VERSION" -o /go/bin/prl-devops-service; \
  fi

############################
# STEP 2 build a small image
############################
FROM alpine:latest

RUN apk update && apk add curl coreutils bash musl-utils ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /tmp/

COPY --from=builder /go/bin/prl-devops-service /go/bin/prl-devops-service

EXPOSE 80

ENTRYPOINT ["/go/bin/prl-devops-service"]