#!/bin/sh
set -eu
# Requires Go 1.27.1, Node.js 24 and pnpm 11.25.0 on Linux.
GOWORK=off
export GOWORK
cd web
pnpm install --frozen-lockfile --ignore-scripts
pnpm exec vue-tsc --noEmit
pnpm exec vite build
cd ..
mkdir -p internal/web/dist
cp -R web/dist/. internal/web/dist/
mkdir -p personal-dist
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -trimpath -buildvcs=false -tags 'with_utls nomsgpack' -ldflags "-s -w -X github.com/yibaiba/hideck/internal/global.Version=0.1.2-personal-classic -X github.com/yibaiba/hideck/internal/global.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o personal-dist/vohive-plus_linux_amd64 ./cmd/hideck
echo 'Run the binary inside the codec-enabled container runtime, not directly on musl iStoreOS.'
