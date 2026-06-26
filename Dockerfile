ARG GOLANG_VERSION=1.26.0
ARG VERSION=unknown
ARG BUILD_TIMESTAMP=unknown
ARG COMMIT_HASH=unknown

# ================== Build Frontend ================== #
FROM node:26-alpine AS frontend-build

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

WORKDIR /app

COPY web/package.json web/pnpm-lock.yaml* ./

# hadolint ignore=DL3016
RUN npm install -g pnpm \
	&& pnpm install --frozen-lockfile

COPY web/ ./

ENV NODE_ENV=production
ENV VITE_VERSION=$VERSION
ENV VITE_BUILD_TIMESTAMP=$BUILD_TIMESTAMP
ENV VITE_COMMIT_HASH=$COMMIT_HASH

RUN pnpm run build

# ================== Build App ================== #

FROM golang:${GOLANG_VERSION}-alpine AS build

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

ARG VERSION=unknown
ARG BUILD_TIMESTAMP=unknown
ARG COMMIT_HASH=unknown

# prefer ash on Alpine so -o pipefail behaves as expected
SHELL ["/bin/ash", "-o", "pipefail", "-c"]

# Install healthcheck cmd
# hadolint ignore=DL3018
RUN apk update \
    && apk add --no-cache ca-certificates curl cosign \
    && rm -rf /var/cache/apk/* \
    && curl -sfL https://raw.githubusercontent.com/hibare/go-docker-healthcheck/main/install.sh | sh -s -- -d -v -b /usr/local/bin

WORKDIR /src/

COPY . /src/

# Copy built frontend into backend source for embedding
COPY --from=frontend-build /app/dist ./cmd/server/static

# hadolint ignore=DL3018
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/hibare/Waypoint/internal/constants.Version=$VERSION -X github.com/hibare/Waypoint/internal/constants.BuildTimestamp=$BUILD_TIMESTAMP -X github.com/hibare/Waypoint/internal/constants.CommitHash=$COMMIT_HASH" -o /bin/waypoint ./main.go

# ================== Build Final Image ================== #
# hadolint ignore=DL3006
FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

# ensure consistent shell in final stage too
SHELL ["/bin/ash", "-o", "pipefail", "-c"]

ENV WAYPOINT_SERVER_LISTEN_ADDR=0.0.0.0

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /bin/waypoint /bin/waypoint

COPY --from=build /usr/local/bin/healthcheck /bin/healthcheck

HEALTHCHECK \
    --interval=30s \
    --timeout=3s \
    CMD ["healthcheck", "--url", "http://localhost:5000/ping"]

EXPOSE 5000

CMD ["/bin/waypoint", "serve"]
