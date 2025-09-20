# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /srv
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o fibra-ctrl ./cmd/main.go

# Development stage
FROM golang:1.23-alpine AS development
WORKDIR /srv
COPY . .
COPY .air.toml .air.toml
RUN go install github.com/cosmtrek/air@v1.30.0  # Install Air for live reloading
CMD ["air", "-c", "/srv/.air.toml"]

# Production stage
FROM alpine:latest AS production
WORKDIR /srv
RUN apk --no-cache add ca-certificates
COPY --from=builder /srv/fibra-ctrl /srv/fibra-ctrl
COPY --from=builder /srv/.env /srv/.env
RUN chmod +x /srv/fibra-ctrl
CMD ["/srv/fibra-ctrl"]
EXPOSE 8080

# Command to run the application
CMD ["./fibra-ctrl"]
