# Coolify deployment

Deploy two applications from the same immutable Git commit.

## Applications

| Application       | Dockerfile               | Public routing                                      | Resources              |
| ----------------- | ------------------------ | --------------------------------------------------- | ---------------------- |
| `ninefold-web`    | `apps/web/Dockerfile`    | `/` except protected/server paths                   | 0.5–1 CPU, 256–512 MiB |
| `ninefold-server` | `apps/server/Dockerfile` | `/api/*`, `/ws`, `/health/*`; private `/internal/*` | 1–2 CPU, 512 MiB–1 GiB |

Deploy image tags `sha-<commit>` or `v0.3.0`; do not deploy `latest`. Mount a persistent volume
only on the server at `/app/data` with at least 5 GiB. Both containers run as non-root. Make their
root filesystems read-only and allow `/tmp` as tmpfs.

The reverse proxy must strip `NINEFOLD_ADMIN_PROXY_HEADER` from public requests. It may set that
header only for authenticated private-network requests from a source included in
`NINEFOLD_ADMIN_TRUSTED_PROXIES`. Keep `/admin`, `/internal/*`, and `/health/status` private.
Redact Room, Match, and replay paths in proxy access logs.

## Environment contract

Copy `.env.example` into Coolify and replace all development values through its secret store.
`NINEFOLD_COOKIE_SECRET` and `NINEFOLD_REPLAY_SIGNING_KEY` are owner-supplied secrets. Production
URLs must use HTTPS. Set the shutdown timeout to `60s`; configure Coolify's stop grace period to at
least the same value. Ordinary application logs are retained 14 days, security logs 30 days, and
admin audit records one year.

## Startup and rollback

The server opens the single SQLite writer, applies forward migrations, recovers active matches, and
then becomes ready. Configure the server health check as `/health/ready` and the web check as `/`.
Deploy the server before the web application when contracts change.

Retain the previous web/server image pair. Before rollback, confirm the previous server supports
the current migration version. Stop readiness, allow the 60-second graceful snapshot/checkpoint,
deploy both previous images, and verify readiness plus reconnect. Never copy only the live SQLite
main file while WAL is active.

No DNS, secret creation, registry publication, or remote deployment is performed by repository
automation.
