############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /go/src

COPY . .

# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/users.service

############################
# STEP 2 build a small image
############################
FROM scratch

# Copy our static executable.
COPY --from=builder /go/bin/users.service /go/bin/users.service

ENV PORT=80

# Run the hello binary.
EXPOSE 80

ENTRYPOINT ["/go/bin/users.service"]