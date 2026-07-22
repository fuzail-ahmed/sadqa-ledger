# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bookworm AS build

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl git make \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

RUN make templ tailwind-cli basecoat htmx \
	&& .tools/tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify \
	&& mkdir -p /out/data /out/uploads \
	&& GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o bin/sadqa-ledger ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=build /src/bin/sadqa-ledger /usr/local/bin/sadqa-ledger
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build --chown=65532:65532 /out/data/ /data/
COPY --from=build --chown=65532:65532 /out/uploads/ /app/uploads/

USER 65532:65532

ENV PORT=8080
ENV DATABASE_PATH=/data/sadqa-ledger.db

EXPOSE 8080

VOLUME ["/data", "/app/uploads"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1

ENTRYPOINT ["sadqa-ledger"]
