// Builds each page separately (vite-plugin-singlefile supports one entry per build)
// Result: dist/{admin,client,unauthorized,forbidden}.html
import { build } from 'vite'
import { rmSync } from 'node:fs'

rmSync('dist', { recursive: true, force: true })
for (const page of ['admin', 'client', 'unauthorized', 'forbidden']) {
  await build({ build: { rollupOptions: { input: `${page}.html` }, emptyOutDir: false } })
}
