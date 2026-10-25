# Wii/DS: one protocol service per repository

The initial grouped extraction is superseded by the full split. Each protocol
has its own Git repository, executable, image, CI checks and release tag:

- wfc-nas: native NAS authentication and the console HTTP routing front door.
- wfc-dls: DLS1 downloads and DLC counts.
- wfc-gpcm: GameSpy login, player sessions, friends and presence (TCP 29900).
- wfc-gpsp: player search/otherslist (TCP 29901).
- wfc-serverbrowser: server queries and messages (TCP 28910).
- wfc-qr2: registration, heartbeat, server directory and room state (UDP 27900).
- wfc-natneg: NAT negotiation sessions (UDP 27901).
- wfc-gamestats: statistics TCP 29920 and native HTTP statistics protocol.
- wfc-sake: storage/file HTTP protocol.
- wfc-race: Nintendo race/leaderboard SOAP protocol.
- wfc-api: operator profile/moderation/statistics API and core-event translator.
- wfc-common: shared Go wire types, codecs, API clients and legacy database helpers.

nn-wfc is a routing-only HTTP compatibility gateway on its original private
address/port. wfc-content is a routing-only compatibility gateway for old content
URLs. Neither launches protocol implementations or hosts public game sockets.
Every new repository has topics openpak/infra/wii/ds; the shared module adds library.
Services use normal vX.Y.Z tags and publish independent multiarch images.

## State ownership and authenticated contracts

GPCM alone owns authenticated player sessions and session-key/game verification.
GPSP calls it over /internal/v1/gpcm/verify; an outage or mismatched key/game fails
verification. Each TCP service retains its own in-container loopback frontend/
backend reload transport, scoped to that protocol. No net/rpc listener crosses
containers or is published externally.

QR2 alone owns the server directory, room state and registration socket. Browser,
NATNEG, GPCM and operator API use /internal/v1/qr2/* operations. Requests use the
shared WFC_INTERNAL_KEY, bounded HTTP timeouts and bounded JSON bodies. Public
native paths never expose these internal routes. GPCM mutations carry a session
generation; QR2 rejects stale generations, and logout cannot remove a replacement
login. QR2 callbacks are queued outside room locks, retried with bounded attempts,
and checked against the current generation by GPCM before a connection is kicked.
No synchronous GPCM->QR2->GPCM lock cycle is introduced.

QR2 reconciles GPCM session snapshots every five seconds. Same-generation refresh
preserves room pointers; newer generations replace old projections. Missing
sessions are removed after an authoritative snapshot, with protection for logins
created while that snapshot was in flight. A 30-second lease bounds stale login
projections during authority outages; cleanup occurs on the next reconciliation
pass. Logged-out generations are tombstoned against delayed snapshot resurrection.
Directory queries fail with 503 when no recent GPCM snapshot exists. Browser
closes a failed directory request instead of advertising a successful empty list.

A QR2 restart loses its own in-memory rooms/registrations. Active GPCM sessions
are rehydrated from snapshots and consoles re-register through ordinary heartbeat/
challenge. It does not close GPCM connections or interrupt already-direct peer
traffic. A GPCM restart closes its own connections and expired projections are
removed; it does not fabricate logged-in users from database records.

## HTTP, tokens and data

NAS forwards each native hostname family to its actual owner; DLS1 is no longer
implemented in NAS, SAKE/race no longer in wfc-content, and statistics/operator API
no longer in nn-wfc. The old console proxy and website URLs keep working through
thin gateways. Internal hops use fresh HTTP connections for legacy NAS framing
compatibility. Database schema, profile IDs, friend codes, shared token key,
operator API secrets and the existing payload/DLC assets are preserved.

Shared code is vendored with sibling replace directives, so each repository
builds and tests independently with GOWORK=off and -mod=vendor. Refresh vendor in
all consumers after contract changes. No sibling service implementation is a
production dependency. Library and service release tags identify independent
versions; the library has no daemon/image.

## Migration, verification and rollback

Before publishing, save original Quadlets, exact images, environment/routing and
a database dump. Start new owners on the internal network without public ports,
verify authenticated internal APIs, then stop the old socket owner and publish
each TCP/UDP port on its corresponding service. Retain the compatibility gateway
at the old address. Dependencies use startup ordering without stop propagation.
All units belong to the normal boot target.

The operator-generated shared token key is unchanged. The initial transfer ends
connections owned by the grouped runtime; subsequent unrelated service releases
leave GPCM connections intact. Rollback stops new socket owners, restores exact
old images/units and the original runtime ports; the unchanged schema needs no
reverse migration. Never put keys, live environment files or raw container inspect
backups in public repositories.

Validation includes generation/lease race tests and an isolated real-Postgres
stack: native NAS->GPCM login, GPSP verification, QR2 heartbeat registration, live
GPCM connection survival across all unrelated protocol restarts, QR2 restart
rehydration and heartbeat recovery, stale callback/logout rejection, authority
outage lease expiry and fail-closed directory queries. Isolated test identities
are not production accounts. Real-console gameplay and multiplayer are separate
checks and are not implied by transport/protocol probes.
