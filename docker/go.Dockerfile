#
# Parametric build for kungal's PURE-GO binaries: the API server (cmd/server)
# and every one-off cmd/migrate / sync-moemoepoint / backfill-* tool.
#
#   docker build -f docker/go.Dockerfile --build-arg CMD=server  -t kungal-api .
#   docker build -f docker/go.Dockerfile --build-arg CMD=migrate -t kungal-migrate .
#
# kungal imports NO cgo (image processing is image_service's job, called over
# HTTP), so everything is a CGO_ENABLED=0 static binary on distroless — there
# is no cgo.Dockerfile here (unlike oauth/image, which embed libwebp).
#
# Build context MUST be the repo root.
ARG GO_VERSION=1.26

# ---- build ----
FROM golang:${GO_VERSION}-trixie AS build
WORKDIR /src
# Manifests first → module-download layer is cached until go.mod/sum change.
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
ARG CMD=server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
        -o /out/app ./cmd/${CMD}

# ---- run ----
# distroless/static: ~2MB base, no shell, nonroot (uid 65532). Bundles
# ca-certificates (outbound HTTPS: OAuth, image_service, wiki, B2, SMTP TLS)
# + tzdata (the daily reset cron / admin stats pin Asia/Shanghai).
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
# The container HEALTHCHECK is set in docker-compose.yml as:
#   test: ["CMD", "/app", "healthcheck"]   # → GETs /healthz, exits 0/1
