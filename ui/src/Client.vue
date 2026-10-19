<script setup>
import { ref } from 'vue'
import { api, list, act, download } from './lib.js'
import AppLayout from "./components/AppLayout.vue";

const page = ref()
const view = ref(null)
const columns = [
  { title: 'Name', key: 'name', dataIndex: 'name' },
  { title: 'Actions', key: 'actions', align: 'right' },
]
const load = () => list('/api/v1/client/server', 'servers')
const url = (s) => `/api/v1/client/server/${s.id}/rendered`

// force_new=true issues a new client certificate
const show = (s) =>
    act(async () => {
      view.value = { title: s.name, text: await (await api('POST', url(s), { force_new: false })).text() }
    })
const save = (s, force_new) => act(() => download('POST', url(s), { force_new }, `${s.name}.ovpn`))
</script>

<template>
  <AppLayout area="client">
    <a-layout-content style="max-width: 1000px; width: 100%; margin: 24px auto">
      <ClientApps />
      <DataPage ref="page" title="Available servers" :columns="columns" :load="load">
        <template #cell="{ column, record }">
          <a-space v-if="column.key === 'actions'" size="middle">
            <a-tooltip title="View config"><a @click="show(record)"><EyeOutlined /></a></a-tooltip>
            <a-tooltip title="Download config"><a @click="save(record, false)"><DownloadOutlined /></a></a-tooltip>
            <a-popconfirm title="Issue a new certificate and download the config?" @confirm="save(record, true)">
              <a-tooltip title="New certificate"><a><SyncOutlined /></a></a-tooltip>
            </a-popconfirm>
          </a-space>
        </template>
      </DataPage>
    </a-layout-content>
    <ViewModal v-if="view" v-bind="view" @close="view = null" />
  </AppLayout>
</template>
