###
# BUILD STAGE
###

# Get golang image, set AS builder
FROM golang:1.27.3-alpine AS builder

# Define workdir
WORKDIR /sim

# Copy go.mod and go.sum
# go.mod defines dependencies, go.sum has checksums for dependencies
COPY go.mod go.sum ./

# Run download to get dependencies
RUN go mod download

# Copy source (simulator.go)
COPY . .

# Tells alpine to make statically linked binary with no deps
# CGo disables the use of native C libraries.
ENV CGO_ENABLED=0

# Build binary. 
RUN go build -o main ./cmd/simulator 

###
# RUN STAGE
###

# Build alpine docker image
FROM alpine:3.24.4

# Add group simgroup with UID 10001, add user simuser with UID 10001
RUN addgroup -g 10001 simgroup && \
    adduser -u 10001 -G simgroup -h /home/simuser -D simuser

# Set workdir as /sim
WORKDIR /sim

# Copy simulator binary from build stage, simuser owns it
COPY --from=builder --chown=simuser:simgroup /sim/main .

# Set simuser: non-root user.
USER simuser

# Run "./main", which will start the simulator.
CMD ["./main"]