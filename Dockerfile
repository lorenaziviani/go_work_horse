# syntax=docker/dockerfile:1
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/worker /app/worker
COPY configs ./configs
ENV REDIS_ADDR=redis:6379
ENV WORKER_COUNT=5
EXPOSE 2112
CMD ["/app/worker"] 