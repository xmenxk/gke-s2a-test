# Use the offical Go image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.22-rc as builder
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code.
COPY *.go ./

# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux go build -o /gke-s2a-test

FROM ubuntu:22.04

# Change the working directory.
WORKDIR /

RUN apt-get update
RUN apt-get install ca-certificates -y

# Copy the binary to the production image from the builder stage.
COPY --from=builder /gke-s2a-test /gke-s2a-test

# Run the web service on container startup.
ENTRYPOINT ["/gke-s2a-test"]