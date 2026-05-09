# ==============================================
# checkout Service Dockerfile
# ==============================================

FROM golang:1.21-alpine AS builder
WORKDIR /build
RUN apk add --no-cache git make

ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /checkout .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /build/checkout /app/
COPY etc/*.yaml /app/etc/ 2>/dev/null || true

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3   CMD curl -f http://localhost:8080/health || exit 1

CMD ["/app/checkout"]
