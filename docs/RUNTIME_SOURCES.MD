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
| **OpenJDK Vulnerability Group** advisories (openjdk.org/groups/vulnerability/advisories/) | one HTML page per quarterly release (34 pages, 2019-04 to 2026-10); a "risk matrix" table per page: CVE, component (`xml/jaxp`, `security-libs/java.security`, `core-libs`, …), CVSS 3.1 score and vector, and a mark per affected feature line (8, 11, 17, 21, 25); a sentence naming the affected update versions ("The affected versions are 25, 21.0.8, 17.0.16, …"); posted to the signed vuln-announce list | the OpenJDK source code of the maintained update lines | the fixed release of a line is the next update release of that line in upstream numbering | the source for builds that follow upstream update numbering (Eclipse Temurin). The component column is what a later code-level model (CVE:F8c) needs. Terms for automated retrieval still to be confirmed before F8b fetches it |
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
