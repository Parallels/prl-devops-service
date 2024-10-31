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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/prl-devops-service

############################
# STEP 2 build a small image
############################
FROM alpine:latest

RUN apk update && apk add curl coreutils bash musl-utils ca-certificates
# Copy our static executable.
# COPY --from=builder /bin/cat /bin/cat
# COPY --from=builder /usr/bin/whoami /usr/bin/whoami
# COPY --from=builder /usr/bin/getent /usr/bin/getent
# COPY --from=builder /bin/uname /usr/bin/uname
# COPY --from=builder /usr/bin/id /usr/bin/id
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /tmp/

COPY --from=builder /go/bin/prl-devops-service /go/bin/prl-devops-service

EXPOSE 80

ENTRYPOINT ["/go/bin/prl-devops-service"]