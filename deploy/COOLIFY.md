# Production deployment

Ninefold deploys as one Compose resource. Only the Caddy `gateway` is public; `web` and `server`
remain on the private Compose network.

## Generate the production configuration

From the repository root:

```sh
go run ./apps/server/cmd/deployment-config \
  -public-url https://ninefold.recica.dev \
  -version 0.3.2 \
  -output .env.production
```

The command creates a new mode-`0600` file and refuses to overwrite an existing path. It prints no
secret values. Keep the file outside backups and password managers unless those systems are
approved for production secrets. The replay public key is deliberately non-secret; the private
PKCS#8 key, cookie secret, and proxy secret are secrets.

## Coolify

Create one resource from this repository with exactly these fields:

| Field          | Value                  |
| -------------- | ---------------------- |
| Build pack     | `Docker Compose`       |
| Base directory | `/`                    |
| Compose file   | `/compose.yaml`        |
| Git reference  | immutable tag `v0.3.2` |

Paste every value from `.env.production` into the resource environment. Mark and lock these
sensitive values:

- `NINEFOLD_COOKIE_SECRET`
- `NINEFOLD_REPLAY_SIGNING_KEY`
- `NINEFOLD_PROXY_SECRET`

After Coolify parses the Compose file:

1. Assign only service `gateway` the URL `https://ninefold.recica.dev:8080`.
2. Leave `web` and `server` without domains.
3. Confirm `ninefold-data` is mounted only at `server:/app/data`.
4. Confirm the server replica count is one and its stop grace period is 60 seconds.
5. Deploy and wait for all three health checks.

The outer Coolify Caddy terminates TLS. The bundled gateway uses HTTP only on its private port 8080,
routes same-origin browser traffic, strips forged admin/proxy headers, and blocks `/admin`,
`/internal/*`, and `/health/status`. Both gateway configurations are embedded in the image; do not
add a Coolify file or bind mount for `/etc/caddy/Caddyfile`.

Do not remove the existing manually created backend resource until the Compose resource passes the
root, API, WebSocket, replay, reconnect, and persistent-database smoke checks. Remove that old
resource manually afterward; repository automation does not mutate it.

## Upgrade

1. Generate a new environment file for the target version. Preserve existing cookie and proxy
   secrets when rotating them is not intended.
2. For replay-key rotation, deploy a web image that trusts the new public key before the server
   starts signing with the corresponding private key. A normally paired Compose build injects the
   matching generated key automatically.
3. Select an immutable qualified Git tag.
4. Preserve the `ninefold-data` volume.
5. Deploy, allow the 60-second shutdown window, and verify readiness, reconnect, and replay
   verification.

Never copy only the SQLite main file while WAL is active.

## Rollback

Retain the previous Compose file and web/server image pair. Before rollback, confirm the previous
server supports the current migration version. Then:

1. stop new traffic;
2. allow graceful snapshot and WAL checkpoint;
3. select the previous immutable tag;
4. keep `ninefold-data` attached;
5. deploy and verify readiness, reconnect, and replay verification.

Do not roll back across an incompatible database migration.

## Standalone Docker host

Point `NINEFOLD_DOMAIN` at the host and allow inbound TCP 80/443 plus UDP 443. Then run:

```sh
docker compose \
  --env-file .env.production \
  -f compose.yaml \
  -f compose.standalone.yaml \
  up -d --build
```

The override publishes 80/443, lets Caddy obtain and renew public certificates, and persists
certificate/configuration state in `caddy-data` and `caddy-config`. `web` and `server` remain
private.

Use the same upgrade and rollback sequence, always combining both Compose files. Back up SQLite
through a WAL-safe snapshot procedure and preserve all three named volumes.
