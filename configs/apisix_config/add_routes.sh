# 通过 Admin API 添加路由
curl -X POST http://127.0.0.1:9180/apisix/admin/routes/1 \
  -H "X-API-KEY: 12345678901234567890" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "/douyin/user/login",
    "methods": ["POST"],
    "upstream": {
      "type": "roundrobin",
      "nodes": {
        "host.docker.internal:9000": 1
      }
    }
  }'

curl -X POST http://127.0.0.1:9180/apisix/admin/routes/2 \
  -H "X-API-KEY: 12345678901234567890" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "/douyin/user/register",
    "methods": ["POST"],
    "upstream": {
      "type": "roundrobin",
      "nodes": {
        "host.docker.internal:9000": 1
      }
    }
  }'
