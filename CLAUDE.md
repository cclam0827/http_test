# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

http_test is a small httpbin-style echo service (Go, gin) that aims to behave like httpbin and httpbun. The priority is lightweight and efficient. README.md covers endpoints, env vars, Docker and the CI secret setup; this file only holds what you need to change code safely.

## Commands

```bash
go run main.go start                                # run (LOG_FORMAT=color for readable dev logs)
go vet ./... && go test -race ./...                 # full check
go test -run TestRootBody ./apis/app/controllers    # one test, any package
go test -run TestCORSPreflight .                    # integration tests live in the root package
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go run golang.org/x/vuln/cmd/govulncheck@latest -tags nomsgpack ./...
```

- Run staticcheck through `go run ...@latest`: the `staticcheck` on PATH on this machine was built with an older Go and cannot read go1.27 export data.
- Point govulncheck at the source (or an unstripped build). The shipped binary is built with `-s -w`, and scanning that in binary mode reports a false positive (GO-2026-5932, `x/crypto/openpgp`, which is not linked).
- Image: `docker buildx build --platform linux/arm64 --load -t httptest:test .` (`--load` takes one platform; for both, use a docker-container builder with `--push` or `--output`).

## Architecture

Request path: `main.go` (urfave/cli `start`) → `logger.Setup` → `apis.NewServer().StartAPI()` → gin engine from `apis.CreateGinServer` → routes in `apis/app/router.go` → `controllers.RootController`.

- **The catch-all is the product.** Only `/info` (GET/PUT/POST/DELETE) is routed. Every other path and method, even `PATCH /info`, falls to `NoRoute` and `RootController.Root`, which echoes the request with 200. Do not add 404/405 handling.
- **Echo rules** (`controllers/root.go`): the body is read up to 1 MiB (`maxEchoBodyBytes`; the excess is dropped and `body_truncated` is true). It is embedded as JSON only when the Content-Type is JSON and the payload is valid, otherwise it is a string. An empty body stays `{}` (the historical value, kept for compatibility). `form` is null unless the request is an urlencoded POST/PUT/PATCH. `headers` maps each name to one string, with a repeated header joined by `, ` as httpbin does, and includes `Host` (Go keeps it in `Request.Host`, not `Request.Header`); `query` and `form` stay arrays. Responses use `c.IndentedJSON` (4 spaces).
- **Middleware order is deliberate** (`CreateGinServer`): AccessLogger, Recoverer, then CORS, so preflights are logged and a recovered panic shows up as a logged 500. `StartAPI` serves through its own `http.Server` instead of `engine.Run`, to get graceful shutdown on SIGINT/SIGTERM (10s) and `ReadHeaderTimeout`.
- **Config** (`core/config`) loads JSON in `init()` from `/etc/app/config.json`, `./etc/app/config.json` or `../etc/app/config.json`, and panics if none exists; env vars override it. Consequence: any package that imports it, directly or transitively (`apis`, `apis/app`, `main`), can only be tested from a directory where one of those paths resolves, which means the repo root. That is why the end-to-end tests are in `main_test.go`. `controllers`, `middlewares` and `logger` deliberately do not import config (the hostname and the `*slog.Logger` are passed in); keep them that way so they stay unit-testable in place.
- **Logging** is `log/slog` (`core/logger`). `Setup` installs the default logger: json when `APP_ENV=production`, text otherwise, and `LOG_FORMAT=color` selects the dev-only handler in `color.go` (its `WithAttrs` must not share a slice with its parent; a test covers that). Each request logs one `http_request` line with `method, path, status, duration, bytes, client_ip` and `error` when set.
- **CORS** is our own middleware (`core/api/middlewares/cors.go`), modelled on httpbin and httpbun: it reflects `Origin`, `Access-Control-Request-Method` and `Access-Control-Request-Headers` instead of sending `*`, because a wildcard fails on credentialed requests and never covers `Authorization`. Allowed origins come from config.json (`["*"]` by default; a specific list makes every other origin get a 403).
- `core/api/errors` is currently not imported anywhere.

## Build and release constraints

- The image is built with `-tags nomsgpack`, which drops gin's unused msgpack support (about 5.7 MB). Tests pass with and without it.
- The Dockerfile cross-compiles in a `--platform=$BUILDPLATFORM` builder, and the final distroless stage has no `RUN`. Keep it that way: a `RUN` in the final stage would need QEMU for the foreign architecture, and distroless has no shell anyway.
- The `go` line in `go.mod` must not be newer than the `golang:` tag in the Dockerfile. The official image sets `GOTOOLCHAIN=local`, so the build fails instead of downloading a toolchain.
- `.github/workflows/docker.yml` is intentionally minimal (one job, no test gate, no PR trigger). A push to `main` publishes `ghcr.io/cclam0827/http_test:latest` and `:sha-<sha>` (amd64 and arm64) to GHCR with the built-in `GITHUB_TOKEN`, so no secret is involved. A new package starts private and must be made public once in its settings.

## Decisions

Stay on gin (chi was considered and rejected), no Swagger/swag, and logging follows the conventions of the sibling `xyztradings-api` repo. Weigh binary size and memory before adding a dependency.
