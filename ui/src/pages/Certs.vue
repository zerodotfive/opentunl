<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { A, api, list, col, sure } from '../lib.js'
import { colors } from '../mount.js'

// Server or client certificates of the selected root certificate
const props = defineProps({ kind: String }) // 'server' | 'client'
const page = ref()
const roots = ref([])
const root = ref()
const create = ref(null)
const rootOptions = computed(() => roots.value.map((r) => ({ value: r.id, label: r.name })))
const columns = [
  col('ID', 'id'),
  col('Name', 'name'),
  col('Valid from', 'not_before', { date: true }),
  col('Valid until', 'not_after', { date: true }),
  { title: 'Status', key: 'status' },
  { title: '', key: 'actions', align: 'right' },
]

// [label, tag color]: revoked takes precedence over expired
const status = (c) => (c.is_revoked ? ['revoked', colors.bad] : new Date(c.not_after) < new Date() ? ['expired', colors.warn] : ['valid', colors.ok])

onMounted(async () => {
  roots.value = await list(`${A}/ca`, 'certificates')
  root.value = roots.value[0]?.id
})
const load = async () => (root.value ? list(`${A}/ca/${root.value}/${props.kind}`, 'certificates') : [])
watch(root, () => page.value?.refresh())
</script>

<template>
  <DataPage
    ref="page"
    :title="kind === 'server' ? 'Server certificates' : 'Client certificates'"
    subtitle="Issued by the selected root certificate"
    :creatable="kind === 'server'"
    :columns="columns"
    :load="load"
    @create="create = { name: '', valid_days: 365 }"
  >
    <template #filters>
      <a-select v-model:value="root" :options="rootOptions" placeholder="Root certificate" style="width: 220px" />
    </template>
    <template #cell="{ column, record }">
      <a-tag v-if="column.key === 'status'" :color="status(record)[1]">{{ status(record)[0] }}</a-tag>
      <a v-else-if="column.key === 'actions' && !record.is_revoked" @click="sure(`Revoke certificate ${record.name}?`, () => api('DELETE', `${A}/ca/${root}/${kind}/${record.id}`), page)">
        <StopOutlined />
      </a>
    </template>
  </DataPage>
  <FormModal
    v-if="create"
    title="Create server certificate"
    :fields="[{ key: 'name', label: 'Name' }, { key: 'valid_days', label: 'Validity, days', type: 'number' }]"
    :init="create"
    :submit="(d) => api('POST', `${A}/ca/${root}/server`, d)"
    @close="create = null; page.refresh()"
  />
</template>
