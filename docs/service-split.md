# Wii/DS service split

The first extraction separates independently releasable HTTP services from the
stateful gameplay runtime. Each is its own repository, binary and image:

- nn-wfc: GPCM/GPSP, QR2, NATNEG, server browser, gamestats, internal moderation/
  profile APIs and the core-event bridge. Keeps the frontend/backend reload mechanism.
- wfc-nas: NAS authentication, DLS1 downloads and the existing HTTP front door.
- wfc-content: SAKE storage and Nintendo race/leaderboard HTTP routes.
- wfc-common: shared protocol, database and authenticated transport module.

GPCM and QR2 have bidirectional live-state operations and error callbacks;
server browser/NATNEG share QR2 state; gamestats imports GPCM protocol handling.
These remain one runtime deliberately. Splitting them further requires a versioned
session/matchmaking contract and explicit recovery semantics, not package copying.
DLS1 shares the NAS form codec and stays with NAS. No Wii-versus-DS duplication.

## Contracts and compatibility

The existing database/schema, profile IDs, friend codes, moderation and event cursor
are preserved. Shared database helpers are compatibility code, not a new identity
owner. Keep the existing database and payload volume. No database migration is added.

Previously tokens used process-random keys. All processes now load the same
operator-created binary WFC_TOKEN_KEY_FILE (>=32 bytes) and derive separate NAS and
login-ticket keys/IVs/magic. Native wire formats and lifetimes are unchanged.
Initial migration invalidates old process-random tokens just like an old backend
restart did; consoles must reconnect once. Subsequent auth/content deployments do
not end runtime sockets or invalidate shared-key tokens. Do not rotate this key
as part of normal deployment.

NAS is the existing HTTP entry point. It forwards SAKE/race hosts to
WFC_CONTENT_URL and gamestats/API routes to WFC_RUNTIME_URL, preserving hostname,
path/query, method and body. It overwrites X-WFC-Internal-Key with WFC_INTERNAL_KEY;
private owners require it. Internal outages return 503, without local fallback.
Existing API-secret checks remain in addition to transport authentication.

Internal defaults: NAS 20120, runtime HTTP 20121, content HTTP 20122. Keep the six
existing public GameSpy TCP/UDP ports on nn-wfc. NAS owns the console HTTP surface. The runtime retains a compatibility HTTP
forward on :20120 to WFC_NAS_URL, so existing proxy and website URLs keep working.
It contains no auth/content implementation. Runtime's net/rpc frontend/backend transport remains loopback-only.

## Build and deployment

Each service commits vendor/ with a sibling replace for wfc-common. CI tests and
images build independently with GOWORK=off; refresh vendor in all consumers after
shared changes. Each service has its own tag-triggered multiarch release workflow.
Use example.env for non-secret configuration shape, supplying operator credentials.
Never commit rendered config.xml, live environment files or token keys.

Before migration save original units, exact image IDs and a database dump. Pin the
old runtime image before publishing latest so automatic updates cannot deploy a
half-configured topology. Start content and NAS with the shared key, then replace
the runtime and move existing HTTP routing to NAS. Preserve the payload volume.
Validate NAS encoding, token verification across processes, direct GameSpy listeners,
SAKE/race/gamestats paths, API transport auth and original host routing. Verify a
runtime challenge connection remains open across independent NAS/content restarts.
Rollback stops NAS/content, restores original runtime image/unit and HTTP routing;
old process-random tokens require another reconnect. Database schema is unchanged.

Console/gameplay verification is separate from automated protocol checks.

## Production validation (2026-10-25)

Deployed releases: nn-wfc 0.3.2, wfc-nas 0.1.2 and wfc-content 0.1.0.
All release images are published for amd64 and arm64. The runtime retains its
existing public ports and private address. NAS owns the existing payload volume;
no database migration or profile export/import was required.

Offline tests/vet and ARM64 builds passed from isolated source contexts. Release
CI uses clean checkouts with vendor mode. Native NAS login produces a ticket
verified by an independent process using the shared key. Live checks cover SAKE
SOAP error framing, gamestats token framing, API authorization and direct private
listener rejection without transport credentials. A GameSpy challenge/keepalive
socket remains connected through separate NAS and content restarts, with the same
runtime container ID. Public NAS and the four GameSpy TCP listeners were checked, along with QR2
availability and NATNEG pre-init over both public UDP ports.

Internal forwarding uses fresh HTTP connections to avoid handing a streamed POST
a stale connection after legacy NAS idle/close behavior or a service restart.
This affects only short internal HTTP hops; persistent game sockets remain in
the runtime. No console gameplay, matchmaking session or positive SAKE record
write is claimed by these checks.

Rollback retains the exact old image, original unit and database dump on the
operator's server. Restore the original combined runtime unit and stop the two
new services; no schema reversal is needed. Keep the shared key and environment
files private. Console trust and external proxy configuration are unchanged.
