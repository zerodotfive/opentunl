import { message, Modal } from 'ant-design-vue'
import dayjs from 'dayjs'

export const A = '/api/v1/admin'
export const fmt = (d) => dayjs(d).format('YYYY-MM-DD HH:mm')

// Table column with client-side sorting; `extra` may set `date: true` to format values as dates
export const col = (title, key, extra) => ({
  title,
  key,
  dataIndex: key,
  sorter: (a, b) => String(a[key] ?? '').localeCompare(String(b[key] ?? '')),
  ...extra,
})

// fetch wrapper: throws Error with the problem details returned by the server
export async function api(method, url, body) {
  const r = await fetch(url, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body && JSON.stringify(body),
  })
  if (r.status === 401) {
    window.location.reload()
    return
  }
  if (!r.ok) {
    const t = await r.text()
    let e
    try { e = JSON.parse(t) } catch { e = { detail: t } }
    throw new Error([e.detail || e.title || r.statusText, ...(e.errors ?? []).map((x) => `${x.location}: ${x.message}`)].join('\n'))
  }
  return r
}

export const json = async (method, url, body) => (await api(method, url, body)).json()

// GET a JSON response and return one of its (possibly null) array fields
export const list = async (url, key) => (await json('GET', url))[key] ?? []

// Fetch a file and hand it to the browser as a download
export async function download(method, url, body, filename) {
  const response = await api(method, url, body)
  const disposition = response.headers.get('Content-Disposition')

  const match = disposition?.match(/filename\*?=(?:UTF-8'')?["']?([^"';\n]+)["']?/i)
  if (match) {
    filename = decodeURIComponent(match[1].trim())
  }

  const a = document.createElement('a')
  a.href = URL.createObjectURL(await response.blob())
  a.download = filename
  a.click()
  setTimeout(() => URL.revokeObjectURL(a.href), 10000)
}

// Run an action, show an error toast on failure; returns true on success
export async function act(fn, ok) {
  try {
    await fn()
    if (ok) message.success(ok)
    return true
  } catch (e) {
    message.error(e.message)
    return false
  }
}

// Confirm dialog, then run the action and refresh the DataPage instance
export const sure = (title, fn, page) =>
  Modal.confirm({
    title,
    okButtonProps: { danger: /^(Delete|Revoke)/.test(title) },
    onOk: async () => {
      if (await act(fn)) page.refresh()
    },
  })
