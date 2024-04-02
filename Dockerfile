############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git 
WORKDIR . 
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN go build -ldflags='-s -w -extldflags "-static"' -o /go/bin/data-layer
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/data-layer /go/bin/data-layer
EXPOSE 8080
# Run the hello binary.
ENTRYPOINT ["/go/bin/data-layer"]

