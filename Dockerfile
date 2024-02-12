############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git coreutils musl-utils ca-certificates 

WORKDIR /go/src

COPY ./src .

# Build the binary.
RUN go get -d -v && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/pd-api-service

############################
# STEP 2 build a small image
############################
FROM scratch

# Copy our static executable.
COPY --from=builder /bin/cat /bin/cat
COPY --from=builder /usr/bin/whoami /usr/bin/whoami
COPY --from=builder /usr/bin/getent /usr/bin/getent
COPY --from=builder /usr/bin/id /usr/bin/id
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /tmp/


COPY --from=builder /go/bin/pd-api-service /go/bin/pd-api-service

EXPOSE 80

ENTRYPOINT ["/go/bin/pd-api-service"]