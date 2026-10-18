# opentunl

**opentunl** is a self-hosted management service for OpenVPN. It runs its own certificate authority, issues server and
client certificates, renders OpenVPN configs from templates, supervises the `openvpn` processes on one or many nodes,
and gives your users a self-service page where they sign in with your identity provider (OIDC) and download a ready-to-use
client config.

Everything is stored in PostgreSQL and shipped as a single Go binary with an embedded Vue UI, so scaling out is just
starting more containers against the same database.

## Features

- **Built-in certificate authority**
    - Multiple root CAs (ECDSA P-384), each with its own certificate revocation list (CRL)
    - Server and client certificates issued and revoked from the UI or the API
    - Configurable CA lifetime and per-CA default client certificate validity
    - CRL download per root CA
- **Diffie-Hellman parameters** generated (4096-bit) and stored in the database
- **OpenVPN server definitions** built from a root CA, a server certificate, DH parameters and two Go templates
  (server config and client config)
- **Multi-node operation** – every running instance registers itself as a node; each server can be enabled or
  disabled per node, and a supervisor on every node starts, stops and reloads `openvpn` processes to match
- **Single sign-on** via any OpenID Connect provider (tested with Authentik); access is controlled by group
  membership
    - members of `ADMIN_GROUP` get the admin console
    - other users see only the servers whose *allowed groups* include one of their groups
- **Self-service client portal** – users view or download their `.ovpn` config; a client certificate is issued
  on first download and can be rotated on demand (at most two valid certificates per user at a time)
- **Admin console** with auto-refreshing tables, filters, sorting and live node status
- **REST API** with OpenAPI documentation (optional)
- **Health endpoint** for container orchestrators
- **Automatic database migrations** on start-up
- Multi-arch container image (`linux/amd64`, `linux/arm64`) and standalone binaries

## Screenshots

### Client portal

Signed-in users see the servers they are allowed to use and can view, download or re-issue their config.

![Client portal](docs/images/client-portal.png)

### Admin console

Servers, their certificates and the nodes they run on:

![Servers](docs/images/admin-servers.png)

Each server is defined by a server certificate, DH parameters, the groups that may use it, and two config templates:

![Edit server](docs/images/admin-server-edit.png)

Enable a server on any subset of nodes and see at a glance which nodes are online and running it:

![Server nodes](docs/images/admin-server-nodes.png)

### Certificate authority

Root certificates:

![Root certificates](docs/images/admin-root-certificates.png)

Server certificates issued by a root:

![Server certificates](docs/images/admin-server-certificates.png)

Client certificates issued by a root, with their revocation status:

![Client certificates](docs/images/admin-client-certificates.png)

## How it works

```
                 ┌────────────────────────── PostgreSQL ──────────────────────────┐
                 │ CAs · certificates · CRLs · DH · server definitions · node state │
                 └───────────────▲───────────────────────────▲─────────────────────┘
                                 │                           │
                      ┌──────────┴─────────┐       ┌─────────┴──────────┐
 browser ──OIDC──────▶│  opentunl (node A) │       │  opentunl (node B) │
 (admin / client UI)  │  API + UI          │       │  API + UI          │
                      │  supervisor ──▶ openvpn     │  supervisor ──▶ openvpn
                      └────────────────────┘       └────────────────────┘
```

Each node runs the same binary. Every 10 seconds it records a heartbeat in the `nodes` table. Every 30 seconds the
supervisor compares the servers enabled for *this* node with the `openvpn` processes it is running, renders the
config for any server that needs to start, launches `openvpn`, kills processes of servers that were disabled or
deleted, and reports the PID back to the database. Nodes are considered online (and servers running) if they have
reported within the last minute.

Because the node name is the hostname, give every container or host a unique one.

## Quick start

The repository ships a `compose.yml` with PostgreSQL and the opentunl image.

1. Create a `.env` file next to `compose.yml`:

   ```ini
   # Members of this group get the admin console
   ADMIN_GROUP=vpn-admins

   # OpenID Connect
   OPENID_CONFIGURATION_ISSUER=https://idp.example.com/application/o/opentunl/
   AUTH_CLIENT_ID=opentunl
   AUTH_CLIENT_SECRET=change-me
   AUTH_REDIRECT_URL=https://opentunl.example.com/auth/callback

   # Session signing key, at least 32 bytes, e.g. `openssl rand -hex 32`
   JWT_KEY=change-me-to-a-long-random-string

   # Optional: serve OpenAPI docs at /docs
   DOCS_ENABLED=false
   ```

2. Register an OIDC client at your identity provider with the redirect URL above. The provider must return a
   `groups` claim (the `openid profile email groups` scopes are requested).

3. Start the stack:

   ```sh
   docker compose up -d
   ```

4. Open the UI. By default the service listens on `127.0.0.1:8080`; put a reverse proxy with TLS in front of it,
   or set `LISTEN` to another address.

   | URL | Who | What |
      | --- | --- | --- |
   | `/` | any signed-in user | client portal |
   | `/admin` | `ADMIN_GROUP` members | admin console |
   | `/health` | anyone | liveness / database check |

The `server` container needs `NET_ADMIN` and `/dev/net/tun` to run OpenVPN, and `network_mode: host` so that the
OpenVPN ports are reachable directly. The sample `compose.yml` uses a default database password and is meant as a
starting point – change it before exposing anything.

### Without Docker

Download the `opentunl-linux-amd64` or `opentunl-linux-arm64` binary from the
[releases page](https://github.com/zerodotfive/opentunl/releases), install the `openvpn` package on the host,
point `POSTGRES_URI` at a PostgreSQL database, and run it as root (or with `CAP_NET_ADMIN`):

```sh
chmod +x opentunl-linux-amd64
POSTGRES_URI=postgres://user:pass@localhost:5432/opentunl ./opentunl-linux-amd64
```

## Configuration

All settings are environment variables.

| Variable | Default | Description |
| --- | --- | --- |
| `LISTEN` | `127.0.0.1:8080` | Address the HTTP server listens on |
| `POSTGRES_URI` | `postgres://postgres:password@localhost:5432/opentunl?sslmode=prefer` | PostgreSQL connection string |
| `MIGRATE` | `true` | Apply database migrations on start-up |
| `AUTH_ENABLED` | `true` | Require OIDC authentication. **Never disable it outside local development** – the admin console and API become open to everyone |
| `OPENID_CONFIGURATION_ISSUER` | – | OIDC issuer URL (discovery is used) |
| `AUTH_CLIENT_ID` | – | OIDC client ID |
| `AUTH_CLIENT_SECRET` | – | OIDC client secret |
| `AUTH_REDIRECT_URL` | – | OIDC redirect URL, `https://<host>/auth/callback` |
| `ADMIN_GROUP` | – | Group whose members may use the admin console |
| `JWT_KEY` | – | Key used to sign session cookies (HS256); at least 32 bytes. Sessions last one hour |
| `DOCS_ENABLED` | `false` | Serve the OpenAPI document and interactive API docs |
| `CONFIG_OUTPUT_DIR` | `/tmp` | Directory where rendered OpenVPN server configs are written (mode `0600`) |

## Setting up a VPN server

In the admin console:

1. **Root certificates → Create.** Pick a name, the CA lifetime and the default validity of client certificates.
2. **Server certificates → Create.** Select the root and issue a certificate for the server.
3. **DH parameters → Create.** Generating a 4096-bit prime can take a while.
4. **Servers → Create.** Choose the server certificate and DH parameters, list the groups allowed to connect, and
   adjust the config templates (they are pre-filled with sensible defaults).
5. **Servers → select nodes.** Switch the server on for the nodes that should run it. Within about 30 seconds the
   node starts `openvpn`; the dialog shows whether each node is online and running the server.

Users in the allowed groups can now open the client portal and download their config. Admins can also render a
client config for any name from the *Client config* action of a server.

### Config templates

Server and client configs are [Go `text/template`](https://pkg.go.dev/text/template) templates. The following
values are available:

| Template | Variables |
| --- | --- |
| Server config | `{{ .CA }}`, `{{ .Cert }}`, `{{ .Key }}`, `{{ .DH }}`, `{{ .CRL }}` |
| Client config | `{{ .CA }}`, `{{ .Cert }}`, `{{ .Key }}` |

All values are PEM-encoded. Typical usage is to inline them:

```
<ca>
{{ .CA -}}
</ca>
<cert>
{{ .Cert -}}
</cert>
<key>
{{ .Key -}}
</key>
```

The default templates live in [`ui/src/templates`](ui/src/templates) and are compiled into the admin page, so you can
change the defaults offered for new servers and rebuild.

### Revoking certificates

Revoking a server or client certificate adds it to the CRL of its root CA. The CRL is available from the root
certificates page (download icon) and as `{{ .CRL }}` in server templates (`crl-verify`).

## REST API

Set `DOCS_ENABLED=true` to get the OpenAPI document and interactive docs. Endpoints are versioned under `/api/v1`:

- `/api/v1/admin/…` – requires membership in `ADMIN_GROUP`:
  root CAs (`/ca`), server and client certificates (`/ca/{root_id}/server`, `/ca/{root_id}/client`), DH parameters
  (`/dh`), server definitions (`/server`, including rendering, per-node enable/disable and reload) and nodes (`/node`)
- `/api/v1/client/…` – any signed-in user: list allowed servers and render the user's client config
- `/health` and the `/auth/*` login endpoints are unauthenticated

Requests are authenticated with the `session` cookie issued after OIDC login.

## Building from source

Requirements: Go, Node.js and a PostgreSQL database.

```sh
# UI (outputs ui/dist/*.html, embedded into the Go binary)
cd ui && npm install && npm run build && cd ..

# Server
go build -o bin/opentunl-server ./cmd/server
```

Or build the container image, which does both:

```sh
docker build -t opentunl .
```

For UI development run `npm run dev` inside `ui/` – it proxies `/api` and `/auth` to a locally running backend. See
[`ui/README.md`](ui/README.md) for the UI layout and theming.

## Releases

Pushing a tag named `vX.Y.Z` triggers a GitHub Actions workflow that builds `linux/amd64` and `linux/arm64` binaries
and a multi-arch image, publishes the image as `ghcr.io/zerodotfive/opentunl:X.Y.Z` and `:latest`, and creates a
GitHub release with the binaries and a `SHA256SUMS` file attached.

## Security notes

- Private keys of root, server and client certificates are stored in PostgreSQL. Restrict access to the database,
  encrypt its storage and back it up accordingly.
- Run opentunl behind TLS; session cookies are `HttpOnly` and `SameSite=Lax`.
- Keep `JWT_KEY` and `AUTH_CLIENT_SECRET` out of version control (`.env` is git-ignored).
- Rendered server configs contain private keys; they are written with mode `0600` to `CONFIG_OUTPUT_DIR`.