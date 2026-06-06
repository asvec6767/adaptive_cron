FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all services
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task-manager ./cmd/task-manager
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler ./cmd/scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metric-agent ./cmd/metric-agent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./cmd/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taskctl ./cmd/taskctl

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/task-manager .
COPY --from=builder /app/scheduler .
COPY --from=builder /app/metric-agent .
COPY --from=builder /app/worker .
COPY --from=builder /app/api-gateway .
COPY --from=builder /app/taskctl .

# Copy config files
COPY --from=builder /app/.env.example .env.example

EXPOSE 8080 8081 8082 8083 8084

CMD ["./api-gateway"]