# syntax=docker/dockerfile:1

# Build stage. It always runs on the build host's own architecture and
# cross-compiles for the requested target, so multi-arch builds are fast and
# need no QEMU. Keep the final stage free of RUN instructions for the same reason.
FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Modules first: this layer is rebuilt only when go.mod or go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static, stripped binary. nomsgpack drops gin's unused msgpack support (-5.7 MB).
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -buildvcs=false -tags nomsgpack -ldflags="-s -w" -o /out/main .


# Final stage: minimal distroless image (ships tzdata, so TZ=... works).
FROM gcr.io/distroless/static-debian13:nonroot

# Links the ghcr.io package to this repository.
LABEL org.opencontainers.image.source="https://github.com/cclam0827/http_test"

ENV APP_ENV=production
EXPOSE 4000

COPY etc/app /etc/app
COPY --from=builder /out/main /app/main

USER nonroot:nonroot
ENTRYPOINT ["/app/main", "start"]
