<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'

// Generic form dialog, open while mounted (use v-if). Emits `close` on cancel or success.
// fields: [{ key, label, type?: 'number' | 'select' | 'tags' | 'area' | 'switch', options? }]
// submit: async (data) => ...; throws to keep the dialog open and show the error
const props = defineProps({ title: String, fields: Array, init: Object, submit: Function })
const emit = defineEmits(['close'])

const data = ref(JSON.parse(JSON.stringify(props.init)))
const busy = ref(false)

async function ok() {
  busy.value = true
  try {
    await props.submit(data.value)
    emit('close')
  } catch (e) {
    message.error(e.message)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <a-modal open :title="title" :confirm-loading="busy" width="640px" @ok="ok" @cancel="emit('close')">
    <a-form layout="vertical">
      <a-form-item v-for="f in fields" :key="f.key" :label="f.label">
        <a-select v-if="f.type === 'select'" v-model:value="data[f.key]" :options="f.options" />
        <a-select v-else-if="f.type === 'tags'" v-model:value="data[f.key]" mode="tags" :open="false" :token-separators="[',', ' ']" placeholder="Type a group and press Enter" />
        <a-input-number v-else-if="f.type === 'number'" v-model:value="data[f.key]" :min="1" style="width: 100%" />
        <a-switch v-else-if="f.type === 'switch'" v-model:checked="data[f.key]" />
        <a-textarea v-else-if="f.type === 'area'" v-model:value="data[f.key]" :rows="8" style="font-family: monospace" />
        <a-input v-else v-model:value="data[f.key]" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>
