FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY micron ./micron
COPY web ./web

RUN GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/micron.wasm ./cmd/wasm \
    && cp /usr/local/go/lib/wasm/wasm_exec.js web/wasm_exec.js

FROM nginxinc/nginx-unprivileged:alpine

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/web /usr/share/nginx/html

EXPOSE 8080
