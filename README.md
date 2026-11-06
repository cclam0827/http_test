# http_test

Lightweight HTTP request testing service in the spirit of httpbin: it echoes back whatever you send it.

## Endpoints

| Path | Methods | Response |
|---|---|---|
| `/info` | GET, PUT, POST, DELETE | server hostname |
| anything else | any | echo of the request: method, path, URI, headers, query, body, form |

About the echoed `body`:

- A JSON payload (JSON `Content-Type`) is embedded as JSON, anything else is returned as a string, and a request without a body gives `{}`.
- Bodies larger than 1 MiB are cut at 1 MiB and flagged with `"body_truncated": true`.
- Fields of urlencoded POST/PUT/PATCH bodies are returned in `form`.

CORS works like httpbin and httpbun: any origin, method and request header is accepted, with credentials, by echoing back the request's `Origin`, `Access-Control-Request-Method` and `Access-Control-Request-Headers`. To restrict origins, list them in `API.AllowedOriginsConfig.app` in `etc/app/config.json` (default `["*"]`); any other origin then gets a 403.

## go version

- go1.27.3 (see `go.mod`)

## run on local

```bash
go run main.go start
```

```bash
# vet and test
go vet ./... && go test -race ./...
```

## configuration

`etc/app/config.json` is loaded first. These environment variables override it:

| Variable | Default | Meaning |
|---|---|---|
| `API_APP_PORT` | `4000` | listen port |
| `APP_ENV` | | `production` switches gin to release mode and makes logs JSON by default |
| `LOG_FORMAT` | `json` if `APP_ENV=production`, else `text` | `json`, `text`, or `color` (colorized, for local development) |
| `LOG_LEVEL` | `info` | `info` or `debug` |

Logging uses the standard library `log/slog`; each request produces one `http_request` line with `method`, `path`, `status`, `duration`, `bytes`, `client_ip` (and `error` when a handler recorded one).

```bash
LOG_FORMAT=color go run main.go start
```

The server drains in-flight requests on SIGINT/SIGTERM (up to 10s) before exiting.

## run with docker

```
# pull from the GitHub Container Registry (linux/amd64 and linux/arm64, the right one is picked automatically)
docker pull ghcr.io/cclam0827/http_test:latest
# run on local
docker run --name httptest -p 80:4000 --rm ghcr.io/cclam0827/http_test:latest
```

The image is a static binary on distroless (non-root, no shell). It used to be published on Docker Hub as `cclam0827/httptest`; that image is no longer updated.

## docker build

The Dockerfile cross-compiles on the build host's own architecture, so a multi-arch build needs no QEMU. CI normally does this; to push by hand (after `docker login ghcr.io`):

```
docker buildx build --platform linux/amd64,linux/arm64 --push -t ghcr.io/cclam0827/http_test:latest .
```

## ci: build and push to ghcr.io

`.github/workflows/docker.yml` builds `linux/amd64` and `linux/arm64` in one go on every push to `main` (and on manual runs from the Actions tab) and pushes a single multi-arch image to `ghcr.io/cclam0827/http_test` as `latest` and `sha-<commit sha>`. Pushes that only touch `**.md` are skipped.

No token or secret is needed: the workflow logs in with the `GITHUB_TOKEN` that GitHub provides, and the first push creates the package.

One-time step after the first successful run: a new package is private by default, so make it public to allow anonymous `docker pull`. GitHub: your profile, Packages, `http_test`, Package settings, Change visibility, Public.
