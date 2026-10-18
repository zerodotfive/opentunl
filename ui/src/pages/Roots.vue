<script setup>
import { ref } from 'vue'
import { A, api, list, col, sure } from '../lib.js'

const page = ref()
const create = ref(null)
const edit = ref(null)
const columns = [
  col('ID', 'id'),
  col('Name', 'name'),
  col('Valid from', 'not_before', { date: true }),
  col('Valid until', 'not_after', { date: true }),
  col('Client validity, days', 'client_default_valid_days'),
  { title: '', key: 'actions', align: 'right' },
]
const fields = [
  { key: 'name', label: 'Name' },
  { key: 'ca_valid_days', label: 'CA validity, days', type: 'number' },
  { key: 'client_default_valid_days', label: 'Default client validity, days', type: 'number' },
]
const load = () => list(`${A}/ca`, 'certificates')
</script>

<template>
  <DataPage ref="page" title="Root certificates" subtitle="Certificate authorities" creatable :columns="columns" :load="load" @create="create = { name: '', ca_valid_days: 3650, client_default_valid_days: 365 }">
    <template #cell="{ column, record }">
      <a-space v-if="column.key === 'actions'" size="middle">
        <a-tooltip title="Edit"><a @click="edit = { ...record }"><EditOutlined /></a></a-tooltip>
        <a-tooltip title="Download CRL"><a :href="`${A}/ca/${record.id}/crl`" download><DownloadOutlined /></a></a-tooltip>
        <a @click="sure(`Delete root certificate ${record.name}?`, () => api('DELETE', `${A}/ca/${record.id}`), page)"><DeleteOutlined /></a>
      </a-space>
    </template>
  </DataPage>
  <FormModal v-if="create" title="Create root certificate" :fields="fields" :init="create" :submit="(d) => api('POST', `${A}/ca`, d)" @close="create = null; page.refresh()" />
  <FormModal v-if="edit" :title="`Edit ${edit.name}`" :fields="fields.slice(2)" :init="edit" :submit="(d) => api('PUT', `${A}/ca/${d.id}`, { client_default_valid_days: d.client_default_valid_days })" @close="edit = null; page.refresh()" />
</template>
