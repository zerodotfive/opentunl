<script setup>
import { ref, reactive, computed } from 'vue'
import { A, api, json, list, act, sure, col, download } from '../lib.js'
import { colors } from '../mount.js'
// default templates for new servers, inlined at build time
import serverTemplate from '../templates/server.conf?raw'
import clientTemplate from '../templates/client.ovpn?raw'

const page = ref()
const form = ref(null) // server create/edit dialog
const client = ref(null) // client config dialog
const view = ref(null) // rendered config viewer
const opts = reactive({ certs: [], dh: [] })
const nodes = reactive({ server: null, rows: [] })

const columns = [
  col('ID', 'id'),
  col('Name', 'name'),
  col('Root certificate', 'root'),
  col('Server certificate', 'cert'),
  col('DH parameters', 'dh_name'),
  { title: 'Nodes', key: 'nodes' },
  { title: '', key: 'actions', align: 'right' },
]
const nodeColumns = [
  { title: 'Node', dataIndex: 'name' },
  { title: 'Online', key: 'online' },
  { title: 'Running', key: 'running' },
  { title: 'Enabled', key: 'enabled' },
]
const fields = computed(() => [
  { key: 'name', label: 'Name' },
  { key: 'server_certificate_id', label: 'Server certificate', type: 'select', options: opts.certs.map((c) => ({ value: c.id, label: c.name })) },
  { key: 'dh_id', label: 'DH parameters', type: 'select', options: opts.dh.map((d) => ({ value: d.id, label: d.name })) },
  { key: 'groups_allowed', label: 'Allowed groups', type: 'tags' },
  { key: 'server_config_template', label: 'Server config template', type: 'area' },
  { key: 'client_config_template', label: 'Client config template', type: 'area' },
])
const clientFields = [
  { key: 'name', label: 'Client name' },
  { key: 'valid_days', label: 'Validity, days', type: 'number' },
  { key: 'force_new', label: 'Issue a new certificate', type: 'switch' },
]

const nameOf = (arr, id) => arr.find((x) => x.id === id)?.name ?? id
const clientDefaultValidDaysOf = (arr, id) => arr.find((x) => x.id === id)?.client_default_valid_days ?? 0

// Servers with referenced objects resolved to names and per-node states attached
async function load() {
  const [servers, roots, dh] = await Promise.all([list(`${A}/server`, 'servers'), list(`${A}/ca`, 'certificates'), list(`${A}/dh`, 'dh')])
  const certs = (await Promise.all(roots.map((r) => list(`${A}/ca/${r.id}/server`, 'certificates')))).flat()
  Object.assign(opts, { certs: certs.filter((c) => !c.is_revoked), dh })
  return Promise.all(
      servers.map(async (s) => ({
        ...s,
        root: nameOf(roots, s.root_certificate_id),
        cert: nameOf(certs, s.server_certificate_id),
        dh_name: nameOf(dh, s.dh_id),
        client_default_valid_days: clientDefaultValidDaysOf(roots, s.root_certificate_id),
        states: await list(`${A}/server/${s.id}/state`, 'state'),
      })),
  )
}

const edit = (id) =>
    act(async () => {
      const s = id ? (await json('GET', `${A}/server/${id}`)).server : { name: '', groups_allowed: [], server_config_template: serverTemplate, client_config_template: clientTemplate }
      form.value = { ...s, groups_allowed: s.groups_allowed ?? [] }
    })
// id/root_certificate_id are read-only and rejected by the API
const saveServer = ({ id, root_certificate_id, ...body }) => api(id ? 'PATCH' : 'POST', id ? `${A}/server/${id}` : `${A}/server`, body)

async function more(key, s) {
  const url = `${A}/server/${s.id}/rendered`
  if (key === 'delete') sure(`Delete server ${s.name}?`, () => api('DELETE', `${A}/server/${s.id}`), page.value)
  else if (key === 'download') act(() => download('GET', url, null, `${s.name}.conf`))
  else act(async () => (view.value = { title: s.name, text: await (await api('GET', url)).text() }))
}

// Node selection dialog: enable/disable the server on each node
async function loadNodes() {
  const [all, states] = await Promise.all([list(`${A}/node`, 'nodes'), list(`${A}/server/${nodes.server.id}/state`, 'state')])
  nodes.rows = all.map((n) => ({ ...n, ...states.find((s) => s.node_name === n.name) }))
}
const openNodes = (s) => act(async () => {
  nodes.server = s
  await loadNodes()
})
async function toggle(n) {
  await act(() => api('POST', `${A}/server/${nodes.server.id}/state/${n.name}/${n.enabled ? 'disable' : 'enable'}`))
  await loadNodes()
}
</script>

<template>
  <DataPage ref="page" title="Servers" subtitle="OpenVPN servers" creatable :columns="columns" :load="load" @create="edit()">
    <template #cell="{ column, record, text }">
      <a v-if="column.key === 'name'" @click="edit(record.id)">{{ text }}</a>
      <div v-else-if="column.key === 'nodes'">
        <a-tag v-if="!record.states.some((s) => s.enabled)">disabled</a-tag>
        <template v-else>
          <a-tag v-for="s in record.states.filter((s) => s.enabled)" :key="s.node_name" :color="s.running ? colors.ok : colors.warn">{{ s.node_name }}</a-tag>
        </template>
        <div><a @click="openNodes(record)">select nodes</a></div>
      </div>
      <a-space v-else-if="column.key === 'actions'" size="middle">
        <a-tooltip title="Client config"><a @click="client = { id: record.id, name: '', valid_days: record.client_default_valid_days, force_new: false }"><UserOutlined /></a></a-tooltip>
        <a-tooltip title="Edit"><a @click="edit(record.id)"><EditOutlined /></a></a-tooltip>
        <a-tooltip title="Reload"><a @click="sure(`Reload server ${record.name}?`, () => api('POST', `${A}/server/${record.id}/reload`), page)"><SyncOutlined /></a></a-tooltip>
        <a-dropdown>
          <a><EllipsisOutlined /></a>
          <template #overlay>
            <a-menu @click="(e) => more(e.key, record)">
              <a-menu-item key="view">View config</a-menu-item>
              <a-menu-item key="download">Download config</a-menu-item>
              <a-menu-item key="delete" danger>Delete</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </a-space>
    </template>
  </DataPage>

  <FormModal v-if="form" :title="form.id ? `Edit ${form.name}` : 'Create server'" :fields="fields" :init="form" :submit="saveServer" @close="form = null; page.refresh()" />
  <FormModal
      v-if="client"
      title="Client configuration"
      :fields="clientFields"
      :init="client"
      :submit="({ id, name, valid_days, force_new }) => download('POST', `${A}/server/${id}/client/${encodeURIComponent(name)}/rendered`, { valid_days, force_new }, `${name}.ovpn`)"
      @close="client = null"
  />
  <ViewModal v-if="view" v-bind="view" @close="view = null" />

  <a-modal v-if="nodes.server" open :title="`Nodes of ${nodes.server.name}`" :footer="null" @cancel="nodes.server = null; page.refresh()">
    <a-table :data-source="nodes.rows" :columns="nodeColumns" row-key="name" :pagination="false" size="small">
      <template #bodyCell="{ column, record }">
        <a-badge v-if="column.key === 'online'" :color="record.online ? colors.ok : colors.off" :text="record.online ? 'online' : 'offline'" />
        <a-badge v-else-if="column.key === 'running'" :color="record.running ? colors.ok : colors.off" :text="record.running ? 'running' : 'stopped'" />
        <a-switch v-else-if="column.key === 'enabled'" :checked="!!record.enabled" @change="toggle(record)" />
      </template>
    </a-table>
  </a-modal>
</template>