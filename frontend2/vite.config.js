// Plugins
import vue from '@vitejs/plugin-vue'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import forwardToTrailingSlashPlugin from './forward-to-trailing-slash-plugin.js'

// Utilities
import { defineConfig } from 'vite'
import { fileURLToPath, URL } from 'node:url'
import { resolve } from 'path';

// https://vitejs.dev/config/
const base = "/front2";

export default defineConfig({
  base: base,
  build: {
    rollupOptions: {
      input: {
        appMain: resolve(__dirname, 'index.html'),
        appBlog: resolve(__dirname, 'blog', 'index.html'),
      },
    },
  },
  plugins: [
    // workaround https://github.com/vitejs/vite/issues/6596
    forwardToTrailingSlashPlugin(base),
    vue({
      template: { transformAssetUrls }
    }),
    // https://github.com/vuetifyjs/vuetify-loader/tree/next/packages/vite-plugin
    vuetify({
      autoImport: true,
      styles: {
        configFile: 'src/styles/settings.scss',
      },
    }),
  ],
  define: { 'process.env': {} },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
    extensions: [
      '.js',
      '.json',
      '.jsx',
      '.mjs',
      '.ts',
      '.tsx',
      '.vue',
    ],
  },
  server: {
    port: 3000,
    strictPort: true,
  },
})
