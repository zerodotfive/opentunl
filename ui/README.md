# opentunl

UI for the opentunl API on Vue 3 + [ant-design-vue](https://antdv.com) 4, built with Vite.

```
npm install
npm run dev     # dev server, proxies /api and /auth to the backend (see vite.config.js)
npm run build   # dist/admin.html, client.html, unauthorized.html, forbidden.html
```

Each page is built separately into one self-contained html file (JS and CSS are inlined).

## Project layout

| Path | What it is |
| --- | --- |
| `admin.html`, `client.html`, `unauthorized.html`, `forbidden.html` | Entry points; each one only mounts a root component |
| `build.mjs` | Runs the Vite build once per entry |
| `vite.config.js` | Plugins, dev proxy (`backend` constant = backend address for `npm run dev`) |
| `src/mount.js` | Mounts a page inside ant-design-vue `ConfigProvider`; **all colors are defined here** |
| `src/lib.js` | API helpers (`api`, `list`, `download`, `act`, `sure`), table column helper `col` |
| `src/Admin.vue` | Admin shell: side menu and hash-based page switching |
| `src/Client.vue` | Client page (`/api/v1/client`) |
| `src/Unauthorized.vue`, `src/Forbidden.vue` | Authorize button and 403 pages |
| `src/pages/*.vue` | Admin pages: Servers, Nodes, Roots, Certs (server and client), Dh |
| `src/templates/server.conf`, `src/templates/client.ovpn` | Default config templates that pre-fill the "Create server" form |
| `src/components/DataPage.vue` | Header, filter, auto-refresh and table shared by all list pages |
| `src/components/FormModal.vue` | Generic create/edit dialog built from a list of fields |
| `src/components/ViewModal.vue` | Read-only text dialog (rendered configs) |
| `src/components/AppLayout.vue` | Top bar with user name and logout |
| `src/components/ClientApps.vue` | Recommended clients (Windows, macOS, Linux install commands) shown on the client page |

Components from `src/components` and all `a-*` components and icons are imported automatically
(`unplugin-vue-components`), so there are no imports for them in `.vue` files.

## Colors

There are no hard-coded colors in `.vue` files: everything is configured in **`src/mount.js`**.
It has two parts.

### 1. Theme (`appTheme`)

Passed to ant-design-vue `ConfigProvider`. Library components take their colors from it.

| Setting | What it controls |
| --- | --- |
| `token.colorPrimary` | Accent color: primary buttons, links, selected menu item, switches, focus rings. Hover/active shades are derived automatically |
| `token.colorBorderSecondary` | Thin lines: border of the side menu, line under the page title |
| `components.Layout.bodyBg` | Page background |
| `components.Layout.lightSiderBg` | Side menu background |

The top bar is configured separately by the exported `header` object
(`{ background, color, fontSize }`; `fontSize` sets the size of the title, user name and icons):
it is applied as an inline style in `AppLayout.vue`, because the `Layout.headerBg`/`headerColor`
theme tokens are not reliably supported by all ant-design-vue versions.
The same applies to `Layout.bodyBg` and `Layout.lightSiderBg` above: if the page or side menu background
does not change, set it with an inline style instead.

Anything not listed uses the library default. For example, table headers are `#fafafa` by default.
To change them, uncomment and edit the prepared `Table` line in `appTheme.components`:

```js
Table: { headerBg: '#fafafa', headerColor: '#1f1f1f', headerSortHoverBg: '#f0f0f0', headerSortActiveBg: '#f5f5f5' },
```

Change `headerBg`, `headerSortHoverBg` and `headerSortActiveBg` together, otherwise a hovered or sorted
column header will look different from the rest. This affects all tables.

Other tokens (`colorText`, `colorBgContainer`, `colorSuccess`, `colorError`, `borderRadius`, ...) and
per-component tokens are listed in the
[ant-design-vue theme docs](https://antdv.com/docs/vue/customize-theme) and on each component's page.

Dark theme: `import { theme } from 'ant-design-vue'` and add `algorithm: theme.darkAlgorithm` to `appTheme`.
Then review the `Layout` values above, they are light-theme colors.

### 2. Status colors (`colors`)

Exported object used by tags and badges:

```js
export const colors = { ok: 'green', warn: 'orange', bad: 'red', off: '#bfbfbf' }
```

| Key | Used for |
| --- | --- |
| `ok` | Valid certificate, running node, online |
| `warn` | Expired certificate, enabled node that is not running |
| `bad` | Revoked certificate |
| `off` | Offline or stopped (badge dots) |

Values may be preset names (`red`, `orange`, `green`, `blue`, `purple`, ...) or any css color.
Where they are used: `src/pages/Certs.vue` (`status()`), `src/pages/Servers.vue` (node tags, node dialog),
`src/pages/Nodes.vue` (online badge).

`src/components/AppLayout.vue`, `src/Admin.vue` and `src/components/DataPage.vue` read the theme
(`theme.useToken()` or the `Layout` tokens), so they follow `appTheme` automatically.

## Default config templates

The "Server config template" and "Client config template" fields of the **Create server** form are
pre-filled from `src/templates/server.conf` and `src/templates/client.ovpn`. Put your templates into
these files as plain text and rebuild; the content is inlined into `admin.html` at build time
(Vite `?raw` imports, see the top of `src/pages/Servers.vue`). Editing an existing server is not
affected: it shows the templates stored on the server.

## Conventions

- List pages use `DataPage` (`columns` + `load`); custom cells go into the `#cell` slot, and a column
  without a `#cell` branch falls back to the plain value (`id` columns get a copy button,
  columns with `date: true` are formatted as dates).
- Create/edit dialogs use `FormModal` with a `fields` array; the dialog is open while it is mounted (`v-if`).
- Destructive actions go through `sure(title, fn, page)`: confirm, run, refresh the table.
  Titles starting with `Delete` or `Revoke` get a red confirm button.
