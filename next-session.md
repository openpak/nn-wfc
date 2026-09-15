# Next session — nn-wfc

Updated 2026-09-15.

WiiLink's `wfc-server` (AGPL) containerised for Wii and DS: NAS auth,
GPCM/GPSP, server browser, QR2, NATNEG, SAKE, gamestats, DLS1 DLC.
Deployed 2026-09-10; HEAD carries WD-2 + WD-3 one commit past the last
tag, so the newest work is not shipped.

## Where things stand

- `v0.1.3` (2026-09-10; 4 tags) is what deployed: container, `WFC_*` env
  rendered into `config.xml` at start; NAS plain HTTP on `WFC_NAS_PORT`
  behind Traefik's port-80 router (nas, naswii, dls1, sake, gamestats,
  race names); GameSpy ports 28910/29900/29901/29920 TCP and
  27900/27901 UDP on the host; Postgres `wwfc`.
- `1752ea9` `/api/profile`: friend code → profile + user id, guarded by
  `WFC_API_SECRET`. openpak.org's link flow calls it and the core owns the
  link (`wfc` namespace); nothing link-related is stored here.
- HEAD `98d0acd` (2026-09-12, WD-2 + WD-3) is UNTAGGED — a tag is how the
  image ships, so prod does not have it:
  - WD-2 `corebridge`: polls the core's event stream every 30 s (cursor in
    `core_events_cursor`) and drops every social event kind through one
    logged branch — the no-transport translator the Wii/DS PRD wanted.
    Needs `WFC_CORE_ADDRESS` + `WFC_CORE_KEY` set at deploy.
  - WD-3: database-free regression suites (nas, gamestats, sake) via
    `go test ./...` — NAS auth form round trip, DLS1 counts, host routing,
    SAKE identity checks, gamestats token + `get2.asp` framing.
- Not console-verified. The DS needs only its DNS setting; the Wii client
  is ../nn-wfc-patcher-wii.
- Untracked (2026-09-15 docs pass): `CHANGELOG.md`, `docs/`, `prds/` stubs.

## Next steps

1. Tag `v0.1.4` to ship WD-2 + WD-3, and confirm the corebridge env
   (`WFC_CORE_ADDRESS`, `WFC_CORE_KEY`) is set on the box.
2. Console-verify: one Wii or DS game through NAS (login + one online
   session), and one friend-code link from openpak.org.

## Pointers

- README: link flow, corebridge section, regression scope
- ../prds/platform-wii-ds-prd.md — WD-2/WD-3 marked done 2026-09-12
