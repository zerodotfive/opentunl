<script setup>
import { ref } from 'vue'
import { A, api, list, col, sure } from '../lib.js'

const page = ref()
const create = ref(null)
const columns = [col('ID', 'id'), col('Name', 'name'), { title: '', key: 'actions', align: 'right' }]
const load = () => list(`${A}/dh`, 'dh')
</script>

<template>
  <DataPage ref="page" title="DH parameters" subtitle="Diffie-Hellman parameters for servers" creatable :columns="columns" :load="load" @create="create = { name: '' }">
    <template #cell="{ column, record }">
      <a v-if="column.key === 'actions'" @click="sure(`Delete DH parameters ${record.name}?`, () => api('DELETE', `${A}/dh/${record.id}`), page)">
        <DeleteOutlined />
      </a>
    </template>
  </DataPage>
  <FormModal v-if="create" title="Create DH parameters" :fields="[{ key: 'name', label: 'Name' }]" :init="create" :submit="(d) => api('POST', `${A}/dh`, d)" @close="create = null; page.refresh()" />
</template>
