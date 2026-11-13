# Target runtime: identification and advisory sources (CVE:F8a)

Study for CVE:F8a, task 6: which data can tell whether the Java runtime an
application is deployed on has a known vulnerability. Checked 2026-11-13.
This document records what was looked at and the decision; it is not a
statement of support (see [SUPPORT.MD](SUPPORT.MD)).

## Identifying the runtime (implemented, `internal/runtimeinfo`)

- `-jdk-home DIR` reads `DIR/release`, the key=value file every JDK/JRE
  build since JDK 9 (and most JDK 8 builds) has at its root: `IMPLEMENTOR`,
  `IMPLEMENTOR_VERSION`, `JAVA_RUNTIME_VERSION`, `JAVA_VERSION`,
  `JAVA_VERSION_DATE`, `OS_NAME`, `OS_ARCH`, `LIBC`, `SOURCE`, `MODULES`.
  Nothing is executed; the scanner's own JDK is never used as a default.
- `-runtime-file FILE` is a declared description (YAML: `vendor`,
  `distribution`, `version`, `os`, `arch`, `libc`,
  `package: {name, version, distro}`); unknown fields are errors.
- Normalization names a vendor/distribution only for an `IMPLEMENTOR` value
  in a fixed table. `Oracle Corporation` (Oracle JDK and the jdk.java.net
  OpenJDK builds) and `Red Hat, Inc.` name a vendor but no distribution. A
  missing `IMPLEMENTOR`, a version without a build number, a missing OS or
  architecture and — for a Debian/Ubuntu package — the distribution release
  (absent from the release file) are listed as unresolved.
- Versions: JEP 322 (`feature.interim.update.patch+build-opt`) and the
  legacy `1.8.0_NNN-bBB` form.

## Advisory sources

| Source | Format | Coverage | Backports | Notes |
|---|---|---|---|---|
| **OpenJDK Vulnerability Group** advisories (openjdk.org/groups/vulnerability/advisories/) | one HTML page per quarterly release (34 pages, 2019-04 to 2026-10); a "risk matrix" table per page: CVE, component (`xml/jaxp`, `security-libs/java.security`, `core-libs`, …), CVSS 3.1 score and vector, and a mark per affected feature line (8, 11, 17, 21, 25); a sentence naming the affected update versions ("The affected versions are 25, 21.0.8, 17.0.16, …"); posted to the signed vuln-announce list | the OpenJDK source code of the maintained update lines | the fixed release of a line is the next update release of that line in upstream numbering | the source for builds that follow upstream update numbering (Eclipse Temurin). The component column is what a later code-level model (CVE:F8c) needs. **Not fetched automatically** — see "Terms of use" below |
| **Debian/Ubuntu via OSV** (`DEBIAN-CVE-…` records for `openjdk-N` source packages) | structured OSV records: package version ranges per distribution release (`Debian:12` …) | JDKs installed from the distribution's packages | in the data (distribution package versions) | goes through ariadne's existing OSV path, cache and snapshots; the distribution release is not in the release file, so it has to be declared |
| **Oracle Critical Patch Updates** | HTML risk matrices per quarter | Oracle JDK / Java SE | Oracle's own release numbering | product-specific; Oracle's findings are not transferred to OpenJDK distributions |

## Decision (2026-11-13)

The first provider (CVE:F8b) is the **upstream OpenJDK** advisories, for
builds that follow upstream update numbering and ship upstream's security
fixes in the same update release (Eclipse Temurin first). Other
distributions get no verdict until their own data (or an explicit statement
that they follow upstream) is added; a runtime from an unsupported vendor
is reported as not assessed, never as unaffected. Debian/Ubuntu packages
through OSV are the second candidate.

## Terms of use (checked 2026-11-13, CVE:F8b)

The OpenJDK web site's Terms of Use (openjdk.org/legal/tou/terms, the
access section) do not allow "any robot, spider, scraper or other
automated means" to access the site without Oracle's express written
permission. The exceptions are crawling to build a search index and
read-only automatic access "as may be necessary in the normal course of
development in connection with authorized use of the code" — a
vulnerability scanner fetching advisories on every run is neither. Two
other machine-readable routes were checked and do not carry the data:
Eclipse Adoptium's release-notes API (api.adoptium.net, Temurin) lists the
JBS issues of a release but no CVE ids; NVD's records for these CVEs name
Oracle JDK/JRE CPEs at single versions, not OpenJDK builds or ranges.
(During this study a handful of openjdk.org pages — the advisory index,
one advisory, the legal pages — were requested once with curl before the
terms were read; nothing from them is in the repository.)

**Decision:** a curated catalog. ariadne never connects to openjdk.org.

## Curating the catalog

1. Save each advisory page from a browser as `YYYY-MM-DD.html` (the
   advisory's date) into a local directory, e.g. `work/openjdk-advisories/`
   (gitignored).
2. `go run ./cmd/openjdk-catalog -version N DIR > /tmp/catalog.yaml` — reads
   only those files; the result is validated by the same loader ariadne
   uses (dates, URLs, CVE ids, versions; a CVE may mark only lines the
   affected-versions sentence names).
3. Check the draft against the pages — in particular `earlier_affected`
   (true only when the page says "… and earlier") and the line marks — and
   try it: `ariadne … -runtime-file temurin.yaml -openjdk-catalog /tmp/catalog.yaml`.
4. Copy it to `internal/runtimevuln/openjdk_advisories.yaml`, bump
   `version`, commit. Quarterly (January, April, July, October).

## How a runtime is assessed (`internal/runtimevuln`)

Only runtimes whose distribution the provider covers are assessed —
**Temurin** first (it uses upstream update numbering and ships upstream's
fixes in the same update). Any other distribution, or an unidentified one,
gets `not_performed` with the reason. For each catalog advisory that covers
the runtime's feature line, against the line's last affected release `L`:
a CVE the matrix does not mark for the line is `not_affected`; a runtime
later than `L` is `not_affected` (fixed); equal to `L` is `affected`;
earlier than `L` is `affected` only when the advisory says earlier releases
are affected too, otherwise `undetermined`. An advisory that does not cover
the line gives no rows and a note (no data, not unaffected). When the
newest catalog advisory is older than 100 days the assessment says the
catalog is behind. Reachability of the runtime's code is never assessed
(`not_assessed`, CVE:F8c); these findings are not part of the v1 verdicts.
