<script setup>
// Return to the page the user was trying to open after login
const login = '/auth/login?return=' + encodeURIComponent(location.pathname + location.search + location.hash)
</script>

<template>
  <AppLayout anon>
    <a-layout-content style="display: flex; align-items: center; justify-content: center">
      <a-button type="primary" :href="login">Authorize</a-button>
    </a-layout-content>
  </AppLayout>
</template>
