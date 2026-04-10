# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go module files first for better layer caching
COPY go.mod ./
COPY go.sum* ./

# Copy source code
COPY . .

# Resolve dependencies and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /taskflow-api ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /taskflow-api .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Set migrations path for runtime
ENV MIGRATIONS_PATH=file://migrations

# Expose the API port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./taskflow-api"]
