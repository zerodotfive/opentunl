<script setup>
</script>
<template>
  <AppLayout area="forbidden">
    <a-layout-content style="display: flex; align-items: center; justify-content: center">
      <a-result title="403" sub-title="You don't have access to this page">
        <template #icon><span style="font-size: 72px">⛔</span></template>
      </a-result>
    </a-layout-content>
  </AppLayout>
</template>
