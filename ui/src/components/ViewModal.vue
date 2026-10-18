<script setup>
// Read-only text dialog, open while mounted (use v-if)
defineProps({ title: String, text: String })
defineEmits(['close'])
</script>

<template>
  <a-modal open :title="title" width="720px" :footer="null" @cancel="$emit('close')">
    <pre style="max-height: 60vh; overflow: auto; margin: 0">{{ text }}</pre>
  </a-modal>
</template>
