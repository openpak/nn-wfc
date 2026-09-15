# Ports — nn-wfc (trimmed)

> Trimmed copy for this repository; the canonical document lives at
> `Openpak/ports.md` and governs. Last synchronised 2026-09-15.

One block per concern, nothing below 20000, nothing at or above 27000 (Photon-Nextendo and
Steam live there). Every service reads its listeners from `<SVC>_HTTP_ADDR`, `<SVC>_GRPC_ADDR`,
`<SVC>_METRICS_ADDR`; the values below are the defaults and the local-run convention. Each
service owns a block of ten: +0 HTTP, +1 gRPC, +2 metrics/pprof, +3..+9 spare.

## NAS

| Block | Service | HTTP | gRPC | metrics |
| --- | --- | --- | --- | --- |
| 20120 | `nn-wfc` NAS (Wii/DS auth, DLC, SAKE, gamestats; plain HTTP behind Traefik :80; `/api/profile` for the website's friend-code link, internal network only) | 20120 | — | — |

## GameSpy (fixed ports)

| Port(s) | Service |
| --- | --- |
| 28910 / 29900 / 29901 / 29920 tcp, 27900 / 27901 udp | `nn-wfc` — GameSpy server browser, GPCM, GPSP, gamestats, QR2, NATNEG (fixed by the games) |

Other services' rows live in the canonical ports.md.
