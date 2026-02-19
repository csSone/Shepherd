import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  // 公共目录，用于存放 config.yaml
  publicDir: 'public',
  server: {
    port: 3000,
    // 前端独立运行，直接连接后端，不需要代理
    // 如需跨域，在后端配置 CORS
    host: true, // 监听所有地址，方便局域网访问
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': [
            'react',
            'react-dom',
            'react-router-dom'
          ],
          'query-vendor': [
            '@tanstack/react-query',
            '@tanstack/react-virtual'
          ],
          'ui-vendor': [
            'zustand'
          ]
        }
      }
    }
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom'],
  },
});
