# Gateway 镜像 (预编译二进制 + 配置)
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY gateway-binary /app/gateway
COPY etc /app/etc
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3   CMD curl -f http://localhost:8080/health || exit 1
CMD ["/app/gateway"]