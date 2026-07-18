# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download

# Install swag CLI for swagger doc generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

# Generate swagger docs before building
RUN swag init -g ./cmd/api/main.go -o ./docs --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# Development stage with hot-reload
FROM golang:1.25-alpine AS development

RUN apk add --no-cache git curl

ENV GOTOOLCHAIN=local

# Install air for hot reload
RUN curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Install swag CLI for swagger doc generation on file changes
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate swagger docs
RUN swag init -g ./cmd/api/main.go -o ./docs --parseDependency --parseInternal

CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.20 AS production

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/api /app/api
COPY migrations ./migrations

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app/uploads

USER appuser

EXPOSE 8080

CMD ["/app/api"]
