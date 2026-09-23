# Changelog — nn-wfc

Generated from git history on 2026-09-15. `git log` stays the source
of truth; this file is the readable summary.

Note: the early history below is the upstream project (WiiLink's Wii WFC server);
OpenPak work starts at the port/fork commit.

## Unreleased

- OpenPak bans (website/docs/ban-lookup.md): a Wii or DS linked to a banned OpenPak account is
  refused at GPCM login with WiiLink's own ban message (error 22002, the path WiiLink uses for
  its profile bans), and a logged-in session whose account becomes banned is kicked within a
  minute with the "You have been banned" message (22002). The website's
  `GET /internal/bans?namespace=wfc&subject=<console user id>` is asked; an unlinked console is
  not an OpenPak account and keeps playing, and a lookup outage lets logins through (the next
  sweep catches them). New environment variables `WEBSITE_INTERNAL_URL` and
  `WEBSITE_INTERNAL_KEY`; the gate is off unless both are set.

- WD-2 + WD-3: the no-transport translator and C10 regression checks [98d0acd]


## v0.1.2 — 2026-09-10



## v0.1.1 — 2026-09-10



## v0.1.0 — 2026-09-10



## v0.1.0 — 2026-09-10

- OpenPak: container, config from WFC_* environment, tag-driven ghcr release [c297250]
- Don't panic if closing a listener errors [f23a6bf]
- API: Reformat API error responses [b70c693]
- API: Migrated other endpoints to use new utility functions [51610cb]
- API: Add baninfo endpoint [4d92f7b]
- GPCM: Don't panic if close connection fails after logout [9005b47]
- NAS: Oops fix payload download requests haha [0b50416]
- NAS: Forward payload subpaths to payload-server [128935f]
- Database: Add support for SUBSTRING in Sake filter query [d543917]
- NAS: Fix out of bounds slice in TLS read [fdfa5ea]
- NAS: Case-insensitive form actions [70bec8f]
- NAS: Use errors.Is for os instead of == operator [c00420f]
- NAS: Don't panic on invalid stage 1 version [6205224]
- NAS: Log DS devname and gsbr code as string [3ae2941]
- Merge branch 'feat/nas-refactor' [72cbdb5]
- NAS: Rename dlc.go to dls.go [b3ef500]
- NAS: Fix a few issues with TLS 1.0 for Wii [db453ee]
- NAS: Use specialized TLS connection instead of piping data [dbed549]
- NAS: Optimize client hello parsing [1fcb4da]
- NAS: Fix issue with proxyConsoleTLS read [08bc107]
- NAS: Rework TLS proxy to use io.Copy [00ee95a]
- NAS: Directly feed data from TLS to HTTP server [24d2a40]
- NAS: Fix index check in proxyConsoleTLS [8c6c3a2]
- NAS: Fix /download endpoint [cf40cf2]
- NAS: Refactor auth requests and auth token [10765a1]
- NAS: Close pipe if error on filtering headers [3d713a2]
- NAS: Run HTTP header filters in a detached goroutine [6320486]
- Rework HTTP request handling entirely [c002bb8]
- Replace nhttp with a custom listener [9a1e0a2]
- NAS: Apply lint suggestions [49e9f82]
- API: Apply lint suggestions [1056114]
- Race: Apply lint suggestions [f940b59]
- Sake: Apply lint suggestions [a4a7fb4]
- ServerBrowser: Apply lint suggestions [277cee6]
- Natneg: Apply lint suggestions [7ca8f47]
- QR2: Apply lint suggestions [5258702]
- GameStats: Apply lint suggestions [0ea3d33]
- GPSP: Apply lint suggestions [5dee992]
- Filter: Apply lint suggestions [a700af4]
- GPCM: Apply lint suggestions [32e7c6a]

- … 440 earlier commits omitted (see `git log`)
