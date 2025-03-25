FROM golang:1.24.1-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./bin/server ./cmd/server


FROM alpine:3.19


RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates


RUN adduser -D appuser
USER appuser

WORKDIR /app


COPY --from=builder /app/bin/server .

COPY --from=builder /app/.env* ./


EXPOSE 8080


HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1


CMD ["./server"]