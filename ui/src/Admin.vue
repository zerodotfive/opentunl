<script setup>
import { ref } from 'vue'
import { theme } from 'ant-design-vue'
import Servers from './pages/Servers.vue'
import Nodes from './pages/Nodes.vue'
import Roots from './pages/Roots.vue'
import Certs from './pages/Certs.vue'
import Dh from './pages/Dh.vue'

// Hash-based navigation: page component + its props
const pages = {
  servers: [Servers],
  nodes: [Nodes],
  roots: [Roots],
  'server-certs': [Certs, { kind: 'server' }],
  'client-certs': [Certs, { kind: 'client' }],
  dh: [Dh],
}
const { token } = theme.useToken() // colors come from the theme (src/mount.js)
const cur = ref('servers')
const sync = () => (cur.value = pages[location.hash.slice(1)] ? location.hash.slice(1) : 'servers')
sync()
addEventListener('hashchange', sync)
const go = (e) => (location.hash = e.key)
</script>

<template>
  <AppLayout title="opentunl admin">
    <a-layout>
      <a-layout-sider theme="light" :width="240" :style="{ borderRight: '1px solid ' + token.colorBorderSecondary }">
        <a-menu :selected-keys="[cur]" style="border: 0" @click="go">
          <a-menu-item-group title="OpenVPN">
            <a-menu-item key="servers"><template #icon><DatabaseOutlined /></template>Servers</a-menu-item>
            <a-menu-item key="nodes"><template #icon><SettingOutlined /></template>Nodes</a-menu-item>
          </a-menu-item-group>
          <a-menu-item-group title="Certificate authority">
            <a-menu-item key="roots"><template #icon><SafetyOutlined /></template>Root certificates</a-menu-item>
            <a-menu-item key="server-certs"><template #icon><TrophyOutlined /></template>Server certificates</a-menu-item>
            <a-menu-item key="client-certs"><template #icon><UserOutlined /></template>Client certificates</a-menu-item>
            <a-menu-item key="dh"><template #icon><KeyOutlined /></template>DH parameters</a-menu-item>
          </a-menu-item-group>
        </a-menu>
      </a-layout-sider>
      <a-layout-content>
        <component :is="pages[cur][0]" v-bind="pages[cur][1]" :key="cur" />
      </a-layout-content>
    </a-layout>
  </AppLayout>
</template>
