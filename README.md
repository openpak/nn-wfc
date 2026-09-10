# OpenPak nn-wfc — Nintendo WFC for Wii and DS

Fork of WiiLink's `wfc-server` (AGPL-3.0): NAS authentication, GameSpy GPCM/GPSP,
server browser, QR2, NATNEG, SAKE, gamestats and DLC for Wii and Nintendo DS games. Nothing
console-specific was changed; OpenPak adds the container and its configuration through
`WFC_*` environment variables (`example.env`, rendered into `config.xml` at start).

Ports: 28910, 29900, 29901, 29920 TCP and 27900, 27901 UDP on the host; NAS is plain HTTP on
`WFC_NAS_PORT` behind Traefik's port-80 router for `nas`, `naswii`, `dls1`, `sake`,
`gamestats` and `race` names. Consoles reach it through their DNS setting
([`nn-sssl-dns`](../nn-sssl-dns)); no CFW is needed on a Wii or DS. Image:
`ghcr.io/openpak/nn-wfc` on tag. Not yet console-verified.

---

# WiiLink WFC
WiiLink Wi-Fi Connection is an open source server replacement for the late Nintendo Wi-Fi Connection, supporting both Nintendo DS and Wii games. This repository contains the server-side source code.

## Setup
You will need:
- PostgreSQL

1. Create a PostgreSQL database. Note the database name, username, and password.
2. Use the `schema.sql` found in the root of this repo and import it into your PostgreSQL database.
3. Copy `config-example.xml` to `config.xml` and insert all the correct data.
4. Run `go build`. The resulting executable `wwfc` is the executable of the server.

## Linking a Wii or DS to an OpenPak account

WFC has no login: a console gets a server-issued user id on first contact and every game
profile hangs off it. Linking therefore happens on openpak.org, not on the console: the
player types a friend code from any game, the website asks this server
`GET /api/profile?fc=…&secret=…` (secret = `WFC_API_SECRET`), and links the returned
`user_id` to the account in the core under the `wfc` namespace. Nothing is stored here; the
core owns the link, and an unlinked console keeps playing as before.
