# ==============================================
# Go-mall 通用构建 Dockerfile
# 从仓库根目录构建单个微服务镜像
# ==============================================

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct \
    GOTOOLCHAIN=auto

WORKDIR /workspace

ARG SERVICE_NAME

COPY . .

RUN test -n "$SERVICE_NAME"
WORKDIR /workspace/services/${SERVICE_NAME}
RUN go mod download
RUN go build -o /out/${SERVICE_NAME} .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

ENV TZ=Asia/Shanghai \
    GIN_MODE=release

WORKDIR /app

ARG SERVICE_NAME
ENV SERVICE_NAME=$SERVICE_NAME

COPY --from=builder /out/${SERVICE_NAME} /app/${SERVICE_NAME}
COPY --from=builder /workspace/services/${SERVICE_NAME}/etc/ /app/etc/

RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser && \
    chown -R appuser:appgroup /app

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["/bin/sh", "-lc", "exec /app/$SERVICE_NAME"]
