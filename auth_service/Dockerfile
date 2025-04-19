# Build stage
FROM golang:alpine AS builder

# Metadata label
LABEL stage=gobuilder

# Set environment variables for Go
ENV CGO_ENABLED=0 GOOS=linux

# Install dependencies
RUN apk add --no-cache tzdata

# Set working directory and copy Go modules
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build the app
COPY . .
RUN go build -o /out/build .

# Final stage
FROM alpine

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata
RUN apk add dumb-init

# Set timezone
ENV TZ=Asia/Jakarta

# Set working directory
WORKDIR /out

RUN touch .env

# Copy the built binary from the builder stage
COPY --from=builder /out/build /out/build

EXPOSE 3000

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Command to run the application
CMD ["/out/build"]