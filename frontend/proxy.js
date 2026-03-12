const http = require('http');
const httpProxy = require('http-proxy');
const path = require('path');
const fs = require('fs');

// 创建代理服务器
const proxy = httpProxy.createProxyServer({});

// 处理代理错误
proxy.on('error', (err, req, res) => {
    console.error('Proxy error:', err);
    res.writeHead(500);
    res.end('Proxy Error');
});

// 创建HTTP服务器
const server = http.createServer((req, res) => {
    // 解析 URL，去掉查询参数
    const urlObj = new URL(req.url, `http://localhost:${PORT}`);
    const pathname = urlObj.pathname;

    // 如果请求的是API路径，转发到网关
    if (pathname.startsWith('/douyin/')) {
        proxy.web(req, res, {
            target: 'http://localhost:8888',
            changeOrigin: true,
            // 添加CORS头
            onProxyRes: (proxyRes) => {
                proxyRes.headers['Access-Control-Allow-Origin'] = '*';
                proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
                proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Access-Token, Refresh-Token';
            }
        });
    } else {
        // 否则返回静态文件
        let filePath = path.join(__dirname, pathname === '/' ? 'index.html' : pathname);

        // 检查文件是否存在
        fs.access(filePath, fs.constants.F_OK, (err) => {
            if (err) {
                // 文件不存在，返回index.html（SPA路由）
                filePath = path.join(__dirname, 'index.html');
            }

            // 读取并返回文件
            fs.readFile(filePath, (err, data) => {
                if (err) {
                    res.writeHead(404);
                    res.end('Not Found');
                    return;
                }

                // 设置正确的Content-Type
                const ext = path.extname(filePath);
                let contentType = 'text/html';
                if (ext === '.css') contentType = 'text/css';
                if (ext === '.js') contentType = 'application/javascript';

                res.writeHead(200, { 'Content-Type': contentType });
                res.end(data);
            });
        });
    }
});

const PORT = 8090;
server.listen(PORT, () => {
    console.log(`前端代理服务器运行在 http://localhost:${PORT}`);
    console.log('API请求将自动转发到 http://localhost:8888');
});
