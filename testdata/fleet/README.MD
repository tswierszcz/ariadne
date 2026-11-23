# Synthetic fleet results for `cmd/ariadne-fleet`

Hand-written, minimal v2 results (only the fields the report reads) — not ariadne output:

- `lib-a`: a library (`callgraph.roots: api`), two `REACHABLE` findings (one supplied by
  the library, one by the user, the latter `SUPPLIED_DISPATCH_ONLY`), Spring counted as
  framework, an `-alt-versions` entry whose newer version no longer has CVE-B.
- `app-b`: an application sharing CVE-A with `lib-a`, OSV data from another day.
- `app-c`: a failed analysis (`meta.json` status `failed`).
- `lib-d`: a library assessed with another tool version, one version not assessed.
- `lib-e`: a library with no advisories and a zero OSV validation time (no OSV answer was dated).
