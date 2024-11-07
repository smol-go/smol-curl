FROM golang:1.23-bullseye AS build-base
ENV GO111MODULE=on

# Create and set working directory
WORKDIR /smolcurl/src

# Copy go.mod and go.sum first for better caching
COPY go.mod ./

# Try downloading dependencies first
RUN go mod download

# Then copy the rest of the source code
COPY . .

# Build image
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/smolcurl

## Image creation stage
FROM scratch

# Copy user from build stage
COPY --from=build-base /etc/passwd /etc/passwd

# Copy smolcurl
COPY --from=build-base /go/bin/smolcurl /go/bin/smolcurl
WORKDIR /go/bin

# Set entrypoint
ENTRYPOINT ["./smolcurl"]