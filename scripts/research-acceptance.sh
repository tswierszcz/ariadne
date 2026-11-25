#!/usr/bin/env bash
# RESEARCH_HARNESS RH-A0: regenerates the producer acceptance package
# docs/research-interface/acceptance (ACCEPTANCE.md there) from the control
# library testdata/research with a bin/ariadne built from a clean checkout:
# bundles (-scope all and reachable), real responses to
# testdata/research/acceptance/requests.json, and mutated responses with the
# rule each one breaks. Every response must pass `research check` and
# replay, every mutation must break its rule, or nothing is replaced.
#
# Usage: scripts/research-acceptance.sh [--allow-dirty]
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
if [ "${1:-}" != "--allow-dirty" ] && [ -n "$(git status --porcelain -- cmd internal go.mod go.sum testdata/research)" ]; then
    echo "research-acceptance: commit the producer first (or --allow-dirty): the package records its revision" >&2
    exit 2
fi
go build -o bin/ariadne ./cmd/ariadne
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
go run ./cmd/research-acceptance -ariadne bin/ariadne -out "$tmp/acceptance"
rm -rf docs/research-interface/acceptance
mv "$tmp/acceptance" docs/research-interface/acceptance
(cd docs/research-interface/acceptance && sha256sum -c --quiet SHA256SUMS)
echo "docs/research-interface/acceptance regenerated"
