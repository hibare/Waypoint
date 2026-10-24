# syntax=docker/dockerfile:1.7

ARG GOLANG_VERSION=1.27.3
ARG NODE_VERSION=26

ARG VERSION=unknown
ARG BUILD_TIMESTAMP=unknown
ARG COMMIT_HASH=unknown

# =============================================================================
# Frontend
# =============================================================================

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS frontend-build

ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

WORKDIR /app

ENV CI=true \
    PNPM_CONFIG_FETCH_MIN_SPEED_KIBPS=0

# Copy dependency manifests first for maximum cache reuse.
COPY web/package*.json \
     web/pnpm-lock.yaml* \
     ./

# Install the exact pnpm version declared in package.json.
RUN --mount=type=cache,target=/root/.npm \
    PNPM_VERSION="$(node -p "require('./package.json').packageManager.split('@')[1].split('+')[0]")" \
    && npm install -g "pnpm@${PNPM_VERSION}"

# Persist pnpm package store for repeated builds on the same builder.
RUN --mount=type=cache,id=waypoint-pnpm-${TARGETARCH},target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --prod=false

COPY web/ .

ENV NODE_ENV=production \
    VITE_VERSION=${VERSION} \
    VITE_BUILD_TIMESTAMP=${BUILD_TIMESTAMP} \
    VITE_COMMIT_HASH=${COMMIT_HASH}

RUN --mount=type=cache,id=waypoint-pnpm-${TARGETARCH},target=/root/.local/share/pnpm/store \
    pnpm run build


# =============================================================================
# Backend
# =============================================================================

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine AS backend-build

ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

WORKDIR /src

ENV GOFLAGS=-mod=readonly

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# Build dependencies only.
# hadolint ignore=DL3018
RUN apk add --no-cache \
        ca-certificates \
    && update-ca-certificates

# Keep Go dependency resolution independently cacheable.
COPY go.mod go.sum ./

RUN --mount=type=cache,id=waypoint-go-mod-${TARGETARCH},target=/go/pkg/mod \
    go mod download

# Application source.
COPY . .

# Embed frontend into backend source.
COPY --from=frontend-build /app/dist ./cmd/server/static

# Cross-compile on BUILDPLATFORM for TARGETOS/TARGETARCH.
RUN --mount=type=cache,id=waypoint-go-mod-${TARGETARCH},target=/go/pkg/mod \
    --mount=type=cache,id=waypoint-go-build-${TARGETARCH},target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS="${TARGETOS}" \
    GOARCH="${TARGETARCH}" \
    go build \
        -mod=readonly \
        -trimpath \
        -ldflags="-s -w \
          -X github.com/hibare/Waypoint/internal/constants.Version=${VERSION} \
          -X github.com/hibare/Waypoint/internal/constants.BuildTimestamp=${BUILD_TIMESTAMP} \
          -X github.com/hibare/Waypoint/internal/constants.CommitHash=${COMMIT_HASH}" \
        -o /bin/waypoint ./main.go


# =============================================================================
# Runtime tools
#
# This stage runs on TARGETPLATFORM so architecture-specific binaries are
# installed for the actual image architecture rather than BUILDPLATFORM.
# =============================================================================

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS runtime-tools

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# hadolint ignore=DL3018
RUN apk add --no-cache \
        ca-certificates \
        curl \
        cosign \
    && update-ca-certificates \
    && curl -sfL https://raw.githubusercontent.com/hibare/go-docker-healthcheck/main/install.sh | sh -s -- -d -v -b /usr/local/bin


# =============================================================================
# Runtime
# =============================================================================

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

WORKDIR /app

ENV WAYPOINT_SERVER_LISTEN_ADDR=0.0.0.0

COPY --from=runtime-tools /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-build /bin/waypoint /bin/waypoint
COPY --from=runtime-tools /usr/local/bin/healthcheck /bin/healthcheck

EXPOSE 5000

HEALTHCHECK \
    --interval=30s \
    --timeout=3s \
    CMD ["healthcheck", "--url", "http://localhost:5000/ping"]

CMD ["/bin/waypoint", "serve"]
