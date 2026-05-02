# Blazeginx

Blazeginx is a small Go HTTP gateway: it receives requests, matches them by
path prefix, proxies them to upstream services, and can also serve static files.

## Production Notice

Do not use Blazeginx in production. You probably didnt intend to.

## Quick Start

Create a config file or use the example:

```bash
cp config/config_example.yml config/config.yml
```

Run Blazeginx with `CONFIG_PATH` pointing to the config:

```bash
CONFIG_PATH=config/config.yml go run ./cmd/blazeginx
```

By default, the public server listens on:

```text
127.0.0.1:8888
```

The admin server listens on:

```text
127.0.0.1:9999
```

## Public And Admin Servers

The public server handles user traffic:

```yaml
addr: "127.0.0.1:8888"
```

The admin server exposes internal endpoints:

```yaml
admin_addr: "127.0.0.1:9999"
```

`addr` and `admin_addr` must be different.

Currently the admin server exposes:

```text
GET /healthz
GET /healthz/
```

## Reverse Proxy

Routes define which upstream should receive requests for a path prefix.

```yaml
routes:
  - name: api
    path: "/api/"
    url: "http://127.0.0.1:8080"
    health_path: "/healthz"
    strip_prefix: false
```

With this config:

```text
GET /api/users -> http://127.0.0.1:8080/api/users
```

Fields:

- `name`: route name for logs and healthcheck output.
- `path`: public path prefix.
- `url`: upstream service URL.
- `health_path`: upstream healthcheck path. Default is `/healthz`.
- `strip_prefix`: removes `path` before sending the request upstream.

Example with `strip_prefix: true`:

```text
GET /api/users -> http://127.0.0.1:8080/users
```

## Storage

Blazeginx has a small in-memory storage with expiration time and background GC.
It is used by the rate limiter to keep per-client token state.

## Rate Limiting

Rate limiting uses a token bucket.

```yaml
rate-limit:
  enabled: true
  max_tokens: 100
  refill_rate: 30s
  default_expiration_time: 30s
  cleanup_interval: 30s
```

Each client has a bucket with up to `max_tokens` tokens. Every request spends
one token. If the bucket is empty, Blazeginx returns:

```text
429 Too Many Requests
```

Tokens are restored over time: one token every `refill_rate`. Client state is
stored for `default_expiration_time`, and old entries are removed every
`cleanup_interval`.

Set `enabled: false` to disable rate limiting.

## Timeouts

Timeouts protect both upstream calls and total request handling time.

```yaml
timeout:
  upstream: 2s
  server: 5s
  idle: 60s
```

Fields:

- `upstream`: maximum time for an upstream request.
- `server`: maximum time for request processing inside Blazeginx.
- `idle`: HTTP server idle timeout.

When a timeout is exceeded, Blazeginx returns:

```text
504 Gateway Timeout
```

## Healthcheck

The admin healthcheck endpoint checks every configured upstream.

```text
GET http://127.0.0.1:9999/healthz
```

For each route, Blazeginx requests:

```text
<route.url><route.health_path>
```

Example response:

```json
{
  "message": "OK",
  "upstream_responses": [
    {
      "name": "api",
      "status": 200,
      "response_time_ms": 3,
      "response": "OK"
    }
  ]
}
```

If at least one upstream check fails, the endpoint returns `500`.

## Static Files

Blazeginx can serve a frontend build directory after proxy routes are checked.

```yaml
static:
  enabled: true
  root: "./web/dist"
```

Behavior:

```text
/assets/app.js      -> ./web/dist/assets/app.js
/favicon.ico        -> ./web/dist/favicon.ico
/dashboard/settings -> ./web/dist/index.html
/assets/missing.js  -> 404
```

If `static.enabled` is `false`, unmatched public requests return `404`.

