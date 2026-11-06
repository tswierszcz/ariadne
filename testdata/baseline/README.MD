# CVE:F0 baseline

Results of the analyzer at the CVE:F0 commit on the prepared petclinic and WebGoat inputs
(`scripts/prepare-petclinic.sh`, `scripts/prepare-webgoat.sh`), run offline with
`-profile spring`. **A reference for measuring change, not ground truth.**

| File | Content |
|---|---|
| `<app>/results.json` | v1 results (petclinic `3/62/26/51/6`, WebGoat `42/44/18/56/0`) |
| `<app>/inputs.tsv` | SHA-256 of the app jar, `deps.txt` and every dependency jar |
| `<app>/responses.tsv` | every cached network response the run read (OSV, GitHub, …): cache key, SHA-256, size |
| `tool.txt` | analyzer commit, Go version, flags |
| `VERDICT_CHANGE_ESTIMATE.MD` | which v1 verdicts F1's contract rules would change |
| `F9_COMPARISON.MD` | CVE:F9b: every v1 verdict change from this baseline to the F0–F7 increments, with its cause (`scripts/cve-compare.sh`) |

The responses are **not** in the repository (petclinic 1083 + WebGoat 896 entries, 1405
distinct, 35 MB; 5.3 MB compressed). Regenerate and export them with:

```sh
scripts/cve-baseline.sh --export /path/outside/repo/baseline-responses.tar.gz
```

The export holds exactly the responses listed in `responses.tsv`. Restoring it into an
empty cache directory and running with `-offline -cache-dir <dir>` reproduces both
`results.json` files byte-for-byte; this was verified when the baseline was taken (every
restored response matched its recorded SHA-256). The script records responses with the
`ARIADNE_CACHE_TRACE` environment variable (`internal/httpcache`), which lists the keys of
cache hits and has no effect on the analysis. Since CVE:F3, `-snapshot-out` records the same
information (plus misses, errors, input hashes, configuration and analyzer versions) in a
self-contained, verifiable form; this baseline stays as the F0 reference.
