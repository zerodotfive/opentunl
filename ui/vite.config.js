import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
import { viteSingleFile } from 'vite-plugin-singlefile'

const backend = 'http://localhost:8080' // dev server proxy target

export default defineConfig(({ command }) => ({
  plugins: [
    vue(),
    // auto-import ant-design-vue components + icons, and everything from src/components
    Components({ resolvers: [AntDesignVueResolver({ importStyle: false, resolveIcons: true })], dts: false }),
    // inline JS/CSS into the html so every page is a single file
    command === 'build' && viteSingleFile(),
  ],
  server: { proxy: { '/api': backend, '/auth': backend } },
}))
