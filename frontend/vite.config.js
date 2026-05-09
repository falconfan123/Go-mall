import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const douyinProxyTarget = env.VITE_API_PROXY_TARGET || 'http://localhost:30088';
  const apiProxyTarget = env.VITE_API_V1_PROXY_TARGET || douyinProxyTarget;
  const port = Number(env.VITE_PORT || 3000);

  return {
    plugins: [react()],
    server: {
      port,
      proxy: {
        '/douyin': {
          target: douyinProxyTarget,
          changeOrigin: true,
        },
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
