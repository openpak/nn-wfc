# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
	--mount=type=bind,source=go.sum,target=go.sum \
	--mount=type=bind,source=go.mod,target=go.mod \
	go mod download -x
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ CGO_ENABLED=0 go build -trimpath -o /out/wwfc .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/wwfc /usr/local/bin/wwfc
COPY --chown=app:app config.template.xml entrypoint.sh game_list.tsv motd.txt filter ./
RUN chmod +x entrypoint.sh && mkdir -p /app/logs && chown app:app /app/logs
USER app
# GameSpy: 28910 29900 29901 29920 tcp, 27900 27901 udp. NAS HTTP: WFC_NAS_PORT (plain HTTP, behind Traefik).
ENTRYPOINT ["/app/entrypoint.sh"]
