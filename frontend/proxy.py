#!/usr/bin/env python3
import http.server
import requests
import os
from urllib.parse import urlparse

class ProxyHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type, Access-Token, Refresh-Token')
        self.end_headers()

    def do_GET(self):
        if self.path.startswith('/douyin/'):
            # 转发API请求到网关
            url = f'http://localhost:8888{self.path}'
            headers = {k: v for k, v in self.headers.items() if k.lower() != 'host'}

            try:
                response = requests.get(url, headers=headers)
                self.send_response(response.status_code)
                for k, v in response.headers.items():
                    if k.lower() not in ['content-length', 'transfer-encoding', 'connection']:
                        self.send_header(k, v)
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(response.content)
            except Exception as e:
                self.send_error(500, f'Proxy Error: {str(e)}')
        else:
            # 处理静态文件
            super().do_GET()

    def do_POST(self):
        if self.path.startswith('/douyin/'):
            # 转发API请求到网关
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)

            url = f'http://localhost:8888{self.path}'
            headers = {k: v for k, v in self.headers.items() if k.lower() != 'host'}

            try:
                response = requests.post(url, data=post_data, headers=headers)
                self.send_response(response.status_code)
                for k, v in response.headers.items():
                    if k.lower() not in ['content-length', 'transfer-encoding', 'connection']:
                        self.send_header(k, v)
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(response.content)
            except Exception as e:
                self.send_error(500, f'Proxy Error: {str(e)}')
        else:
            self.send_error(405, 'Method Not Allowed')

    def do_PUT(self):
        if self.path.startswith('/douyin/'):
            # 转发API请求到网关
            content_length = int(self.headers['Content-Length'])
            put_data = self.rfile.read(content_length)

            url = f'http://localhost:8888{self.path}'
            headers = {k: v for k, v in self.headers.items() if k.lower() != 'host'}

            try:
                response = requests.put(url, data=put_data, headers=headers)
                self.send_response(response.status_code)
                for k, v in response.headers.items():
                    if k.lower() not in ['content-length', 'transfer-encoding', 'connection']:
                        self.send_header(k, v)
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(response.content)
            except Exception as e:
                self.send_error(500, f'Proxy Error: {str(e)}')
        else:
            self.send_error(405, 'Method Not Allowed')

    def do_DELETE(self):
        if self.path.startswith('/douyin/'):
            # 转发API请求到网关
            url = f'http://localhost:8888{self.path}'
            headers = {k: v for k, v in self.headers.items() if k.lower() != 'host'}

            try:
                response = requests.delete(url, headers=headers)
                self.send_response(response.status_code)
                for k, v in response.headers.items():
                    if k.lower() not in ['content-length', 'transfer-encoding', 'connection']:
                        self.send_header(k, v)
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(response.content)
            except Exception as e:
                self.send_error(500, f'Proxy Error: {str(e)}')
        else:
            self.send_error(405, 'Method Not Allowed')

if __name__ == '__main__':
    import sys
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8090
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    http.server.ThreadingHTTPServer(('', port), ProxyHTTPRequestHandler).serve_forever()
    print(f'代理服务器运行在 http://localhost:{port}')
