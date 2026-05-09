import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const projectRoot = path.resolve(__dirname, '..');
const gatewayTemplatePath = path.join(projectRoot, 'k8s/helm/go-mall/templates/configmap.yaml');
const frontendApiPath = path.join(projectRoot, 'frontend/src/services/api.js');
const outputDir = path.join(projectRoot, 'test/apifox');
const outputPath = path.join(outputDir, 'go-mall-gateway.openapi.json');

const source = fs.readFileSync(gatewayTemplatePath, 'utf8');
const frontendSource = fs.readFileSync(frontendApiPath, 'utf8');
const mappingRegex = /Method:\s*(\w+)\s*\n\s*Path:\s*([^\n]+)\s*\n\s*RpcPath:\s*([^\n]+)/g;
const frontendApiRegex = /api\.(get|post|put|delete)\((`|'|")([^`'"]+)\2/g;

const routeOverrides = {
  'get /api/v1/system/time': {
    summary: '获取系统时间',
    responses: {
      '200': {
        description: '成功',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/SystemTimeResponse' },
            example: { now: 1778336885665 },
          },
        },
      },
    },
  },
  'get /api/v1/activity/token': {
    summary: '获取秒杀动态令牌',
    parameters: [
      {
        name: 'activity_id',
        in: 'query',
        required: true,
        schema: { type: 'integer', format: 'int64' },
        example: 1,
      },
    ],
    responses: {
      '200': {
        description: '成功',
        content: {
          'application/json': {
            schema: {
              type: 'object',
              properties: {
                path_key: { type: 'string' },
                expires_at: { type: 'integer', format: 'int64' },
              },
            },
            example: { path_key: 'seckill-token-demo', expires_at: 1778339999999 },
          },
        },
      },
    },
  },
  'get /api/v1/activity/status': {
    summary: '查询秒杀购买状态',
    parameters: [
      {
        name: 'activity_id',
        in: 'query',
        required: true,
        schema: { type: 'integer', format: 'int64' },
        example: 1,
      },
    ],
  },
  'post /api/v1/order/seckill': {
    summary: '提交秒杀订单',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: { $ref: '#/components/schemas/SeckillRequest' },
          example: {
            path_key: 'seckill-token-demo',
            product_id: 1,
            user_id: 1001,
            address_id: 1,
          },
        },
      },
    },
  },
  'get /api/v1/order/detail': {
    summary: '查询订单详情',
    parameters: [
      {
        name: 'order_id',
        in: 'query',
        required: true,
        schema: { type: 'string' },
        example: 'ORDER-001',
      },
      {
        name: 'user_id',
        in: 'query',
        required: true,
        schema: { type: 'integer', format: 'int32' },
        example: 1001,
      },
    ],
  },
  'post /api/v1/order/cancel': {
    summary: '取消订单',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              order_id: { type: 'string' },
              user_id: { type: 'integer', format: 'int32' },
              cancel_reason: { type: 'string' },
              initiative: { type: 'boolean' },
            },
            required: ['order_id', 'user_id'],
          },
          example: {
            order_id: 'ORDER-001',
            user_id: 1001,
            cancel_reason: 'manual test',
            initiative: true,
          },
        },
      },
    },
  },
  'post /api/v1/checkout/prepare': {
    summary: '准备结算',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              user_id: { type: 'integer', format: 'int32' },
              address_id: { type: 'integer', format: 'int64' },
              coupon_id: { type: 'string' },
            },
          },
          example: { user_id: 1001, address_id: 1, coupon_id: '' },
        },
      },
    },
  },
  'post /api/v1/auth/token': {
    summary: '签发令牌',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              user_id: { type: 'integer', format: 'int32' },
              device_id: { type: 'string' },
            },
            required: ['user_id', 'device_id'],
          },
        },
      },
    },
  },
  'post /api/v1/auth/validate': {
    summary: '校验令牌',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              token: { type: 'string' },
            },
            required: ['token'],
          },
        },
      },
    },
  },
  'post /douyin/user/login': {
    summary: '用户登录',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              username: { type: 'string' },
              password: { type: 'string' },
              device_id: { type: 'string' },
            },
            required: ['username', 'password', 'device_id'],
          },
          example: {
            username: 'demo',
            password: 'demo123456',
            device_id: 'device-demo-001',
          },
        },
      },
    },
  },
  'post /douyin/user/register': {
    summary: '用户注册',
    requestBody: {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            properties: {
              username: { type: 'string' },
              password: { type: 'string' },
              email: { type: 'string', format: 'email' },
            },
            required: ['username', 'password'],
          },
        },
      },
    },
  },
};

const genericResponse = {
  '200': {
    description: '成功',
    content: {
      'application/json': {
        schema: { $ref: '#/components/schemas/GenericResponse' },
      },
    },
  },
};

const paths = {};
const knownRoutes = new Set();

function buildTagFromPath(rawPath) {
  const segments = rawPath.split('/').filter(Boolean);
  if (segments[0] === 'api') {
    return segments[2] || 'api';
  }
  return segments[1] || segments[0] || 'gateway';
}

for (const match of source.matchAll(mappingRegex)) {
  const method = match[1].toLowerCase();
  const rawPath = match[2].trim();
  const rpcPath = match[3].trim();
  const openapiPath = rawPath.replace(/:([A-Za-z0-9_]+)/g, '{$1}');
  const routeKey = `${method} ${rawPath}`;
  const override = routeOverrides[routeKey] || {};

  const pathParams = [...openapiPath.matchAll(/\{([^}]+)\}/g)].map((paramMatch) => ({
    name: paramMatch[1],
    in: 'path',
    required: true,
    schema: { type: 'string' },
  }));
  const tag = buildTagFromPath(rawPath);

  if (!paths[openapiPath]) {
    paths[openapiPath] = {};
  }

  paths[openapiPath][method] = {
    tags: [tag],
    summary: override.summary || rpcPath.split('/').pop(),
    description: `Generated from gateway mapping ${rpcPath}.`,
    operationId: rpcPath.replace(/[/.]/g, '_'),
    parameters: [...pathParams, ...(override.parameters || [])],
    responses: override.responses || genericResponse,
  };

  if (override.requestBody) {
    paths[openapiPath][method].requestBody = override.requestBody;
  }

  knownRoutes.add(`${method} ${openapiPath}`);
}

for (const match of frontendSource.matchAll(frontendApiRegex)) {
  const method = match[1].toLowerCase();
  const rawUrl = match[3];
  const [rawPath, rawQuery = ''] = rawUrl.split('?');
  const openapiPath = rawPath.trim();
  const routeKey = `${method} ${openapiPath}`;
  if (knownRoutes.has(routeKey)) {
    continue;
  }

  const parameters = [];
  for (const querySegment of rawQuery.split('&').filter(Boolean)) {
    const [name] = querySegment.split('=');
    if (!name) {
      continue;
    }
    parameters.push({
      name,
      in: 'query',
      required: true,
      schema: { type: 'string' },
    });
  }

  if (!paths[openapiPath]) {
    paths[openapiPath] = {};
  }

  paths[openapiPath][method] = {
    tags: [`frontend-${buildTagFromPath(openapiPath)}`],
    summary: `Frontend contract for ${openapiPath}`,
    description: 'Derived from frontend/src/services/api.js. Use this path in Apifox Mock when the backend route is not implemented yet.',
    operationId: `frontend_${method}_${openapiPath.replace(/[/?=&:{}.-]/g, '_')}`,
    parameters,
    responses: genericResponse,
  };

  if (method !== 'get' && method !== 'delete') {
    paths[openapiPath][method].requestBody = {
      required: true,
      content: {
        'application/json': {
          schema: {
            type: 'object',
            additionalProperties: true,
          },
        },
      },
    };
  }
}

const openapi = {
  openapi: '3.0.3',
  info: {
    title: 'Go Mall Test Contract API',
    version: '0.1.0',
    description: 'Generated from gateway route mappings and frontend API calls for Apifox import, mock testing, and gateway smoke testing.',
  },
  servers: [
    {
      url: '{baseUrl}',
      description: 'Gateway base URL',
      variables: {
        baseUrl: {
          default: 'http://127.0.0.1:18888',
        },
      },
    },
  ],
  paths,
  components: {
    schemas: {
      GenericResponse: {
        type: 'object',
        additionalProperties: true,
        properties: {
          status_code: { type: 'integer' },
          status_msg: { type: 'string' },
          message: { type: 'string' },
        },
      },
      SystemTimeResponse: {
        type: 'object',
        properties: {
          now: { type: 'integer', format: 'int64' },
        },
        required: ['now'],
      },
      SeckillRequest: {
        type: 'object',
        properties: {
          path_key: { type: 'string' },
          product_id: { type: 'integer', format: 'int64' },
          user_id: { type: 'integer', format: 'int32' },
          address_id: { type: 'integer', format: 'int64' },
        },
        required: ['path_key', 'product_id', 'user_id', 'address_id'],
      },
    },
  },
};

fs.mkdirSync(outputDir, { recursive: true });
fs.writeFileSync(outputPath, `${JSON.stringify(openapi, null, 2)}\n`);
console.log(`Generated ${path.relative(projectRoot, outputPath)}`);
