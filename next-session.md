# Next session — nn-wfc

Updated 2026-09-24.

WiiLink's `wfc-server` (AGPL) containerised for Wii and DS: NAS auth,
GPCM/GPSP, server browser, QR2, NATNEG, SAKE, gamestats, DLS1 DLC.
Deployed 2026-09-10.

Current status 2026-09-24: latest tag v0.2.0 (c266221, 2026-09-23), tree
clean. v0.2.0 shipped WD-2 + WD-3 (98d0acd) plus OpenPak bans (GPCM refuses
and kicks consoles linked to a banned account, 22002; needs
`WEBSITE_INTERNAL_URL` + `WEBSITE_INTERNAL_KEY`). Since: CI on `v*.*.*` tags
only, docs.

## Where things stand

- `v0.1.3` (2026-09-10) is what deployed on 2026-09-10: container, `WFC_*` env
  rendered into `config.xml` at start; NAS plain HTTP on `WFC_NAS_PORT`
  behind Traefik's port-80 router (nas, naswii, dls1, sake, gamestats,
  race names); GameSpy ports 28910/29900/29901/29920 TCP and
  27900/27901 UDP on the host; Postgres `wwfc`.
- `1752ea9` `/api/profile`: friend code → profile + user id, guarded by
  `WFC_API_SECRET`. openpak.org's link flow calls it and the core owns the
  link (`wfc` namespace); nothing link-related is stored here.
- `98d0acd` (2026-09-12, WD-2 + WD-3), shipped in v0.2.0:
  - WD-2 `corebridge`: polls the core's event stream every 30 s (cursor in
    `core_events_cursor`) and drops every social event kind through one
    logged branch — the no-transport translator the Wii/DS PRD wanted.
    Needs `WFC_CORE_ADDRESS` + `WFC_CORE_KEY` set at deploy.
  - WD-3: database-free regression suites (nas, gamestats, sake) via
    `go test ./...` — NAS auth form round trip, DLS1 counts, host routing,
    SAKE identity checks, gamestats token + `get2.asp` framing.
- v0.2.0 bans: the website's `/internal/bans?namespace=wfc` is asked at
  GPCM login and by a sweep; gate off unless both `WEBSITE_INTERNAL_*` set.
- Not console-verified. The DS needs only its DNS setting; the Wii client
  is ../nn-wfc-patcher-wii.
- 2026-09-15 docs pass committed: `CHANGELOG.md`, `docs/`, `prds/` stubs.

## Next steps

1. ~~Tag `v0.1.4` to ship WD-2 + WD-3~~ — shipped as v0.2.0 (c266221).
   Still confirm the corebridge env (`WFC_CORE_ADDRESS`, `WFC_CORE_KEY`) and
   the ban gate env (`WEBSITE_INTERNAL_URL`, `WEBSITE_INTERNAL_KEY`) are set
   on the box.
2. Console-verify: one Wii or DS game through NAS (login + one online
   session), and one friend-code link from openpak.org.

## Pointers

- README: link flow, corebridge section, regression scope
- ../prds/platform-wii-ds-prd.md — WD-2/WD-3 marked done 2026-09-12

## Scratch (research and throwaway work)

Decompiles, Ghidra projects, dumps, exefs/romfs extracts, packet captures,
strace and emulator logs, probe harnesses: put them in
`~/REPOS/Openpak/scratch/<topic>`. That folder is a local mount of the media pool,
outside every repository, so nothing in it is committed. Never use `/tmp` (a
shared 15 GB RAM disk) or elsewhere on `/home` for this. Keys and signing
material never go there. Rule: `docs/playbooks/conventions.md` in the workspace.
