import { createApp, h } from 'vue'
import { ConfigProvider } from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'

// All colors of the UI are defined in this file.

// Status colors used by tags and badges (preset names or any css color)
export const colors = { ok: 'green', warn: 'orange', bad: 'red', off: '#bfbfbf' }

// Top bar colors, applied as inline style in AppLayout.vue
export const header = { background: '#1c1c1c', color: '#fff'}

const appTheme = {
  token: {
    colorPrimary: '#0078d4', // accent: buttons, links, selected menu item, switches, focus
    colorBorderSecondary: '#f0f0f0', // thin lines: side menu border, line under page titles
  },
  components: {
    Layout: {
      bodyBg: '#fff', // page background
      lightSiderBg: '#fff', // side menu background
    },
    // Table: { headerBg: '#fafafa', headerColor: '#1f1f1f', headerSortHoverBg: '#f0f0f0', headerSortActiveBg: '#f5f5f5' },
  },
}

// Mounts a page inside ConfigProvider (shared theme)
export default (page) =>
    createApp({
      render: () => h(ConfigProvider, { theme: appTheme }, () => h(page)),
    }).mount('#app')
