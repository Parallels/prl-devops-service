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

RUN --mount=type=secret,id=amplitude_api_key,env=AMPLITUDE_API_KEY

RUN if [ "$BUILD_ENV" = "production" ]; then \
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCHITECTURE go build -ldflags="-s -w -X main.ver=$VERSION -X 'github.com/Parallels/prl-devops-service/telemetry.AmplitudeApiKey=$AMPLITUDE_API_KEY'" -o /go/bin/prl-devops-service; \
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