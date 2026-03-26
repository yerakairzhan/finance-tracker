FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

ENV PORT=8080
ENV JWT_SECRET=dev-jwt-secret-change-me

WORKDIR /root/

COPY --from=builder /app/api .

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 \
  CMD wget -qO- "http://127.0.0.1:${PORT}/health/ready" >/dev/null || exit 1

CMD ["./api"]
