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

## OpenPak bans

A Wii or DS linked to an OpenPak account that is banned in the admin is refused at GPCM login
with WiiLink's ban message (22002) and kicked within a minute if already online. The server asks
the website's `GET /internal/bans?namespace=wfc&subject=<console user id>`
([`website/docs/ban-lookup.md`](../website/docs/ban-lookup.md)); set `WEBSITE_INTERNAL_URL`
(`http://website:20010` on the box) and `WEBSITE_INTERNAL_KEY` to turn it on. Unlinked
consoles are not OpenPak accounts and are never refused; a lookup outage lets logins through.

## The core bridge (`corebridge`, universal-social translator)

When `WFC_CORE_ADDRESS` (and `WFC_CORE_KEY`, the core's `X-API-Key`) are set, the backend
polls the account core's event stream every 30 seconds and keeps its position in a
`core_events_cursor` table in the same Postgres database. What it does with each social
event — `friend_requested`, `friend_accepted`, `friend_removed`, presence, invitations,
chat — is deliver it nowhere and log the drop: WFC has no push of any kind, so there is no
transport to deliver with. The bridge exists so the family is honestly part of the universal
model, and so the model can prove it needs no special case for a family that receives
nothing (prds/platform-wii-ds-prd.md WD-2). Run without `WFC_CORE_ADDRESS`, nothing
changes.

Regression checks for the console-facing surface (NAS auth form and reply encoding, DLS1 DLC
counts, host routing, SAKE identity checks, gamestats token and `get2.asp` framing) run with
`go test ./...`; they need no database, only the repository root as working directory.
