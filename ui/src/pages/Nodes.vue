<script setup>
import { ref } from 'vue'
import { A, api, list, col, sure } from '../lib.js'
import { colors } from '../mount.js'

const page = ref()
const columns = [
  col('Name', 'name'),
  col('Status', 'online'),
  col('Last seen', 'last_seen', { date: true }),
  { title: '', key: 'actions', align: 'right' },
]
const load = () => list(`${A}/node`, 'nodes')
</script>

<template>
  <DataPage ref="page" title="Nodes" subtitle="Servers running OpenVPN" row-key="name" :columns="columns" :load="load">
    <template #cell="{ column, record }">
      <a-badge v-if="column.key === 'online'" :color="record.online ? colors.ok : colors.off" :text="record.online ? 'online' : 'offline'" />
      <a v-else-if="column.key === 'actions'" @click="sure(`Delete node ${record.name}?`, () => api('DELETE', `${A}/node/${record.name}`), page)">
        <DeleteOutlined />
      </a>
    </template>
  </DataPage>
</template>
