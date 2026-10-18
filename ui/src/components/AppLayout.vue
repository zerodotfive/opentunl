<script setup>
import { ref, onMounted } from 'vue'
import { json } from '../lib.js'
import { header } from '../mount.js'

const props = defineProps({ title: { default: 'opentunl' }, anon: Boolean })
const user = ref('')
const logout = '/auth/logout?return=' + encodeURIComponent(location.pathname + location.search)

// Current user name for the header (not available on the unauthorized page)
onMounted(async () => {
  if (props.anon) return
  try {
    user.value = (await json('GET', '/auth/info')).user_info.name
  } catch {}
})
</script>

<template>
  <a-layout style="min-height: 100vh">
    <a-layout-header :style="{ display: 'flex', alignItems: 'center', gap: '12px', padding: '0 16px', height: '48px', lineHeight: '48px', ...header }">
      <span style="font-size: calc(1em + 4px);"><SafetyOutlined /></span>
      <span style="font-size: calc(1em + 4px);">{{ title }}</span>
      <span style="flex: 1" />
      <template v-if="user">
        <span>{{ user }}</span>
        <a :href="logout" style="color: inherit"><LogoutOutlined /></a>
      </template>
    </a-layout-header>
    <slot />
  </a-layout>
</template>