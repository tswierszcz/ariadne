#!/bin/bash
# Library mode measurement on real libraries (LIBRARY_MODE LB2/LB4).
#
# Fetches four libraries with vulnerable dependencies from Maven Central
# (mvn dependency:copy-dependencies) into work/libs/<name>/ — lib.jar is the
# library, deps/ its runtime dependencies — and runs bin/ariadne on each
# with -roots all and -roots api, under CHA and RTA, printing one summary
# line per run. Results of 2026-11-17: RECOMMENDATION/IMPLEMENTATION_STATUS.MD
# (LIBRARY_MODE:LB2 pomiar, lb2-5 for the starter, lb3-1 for the
# dependency declarations).
#
# Usage: scripts/library-measure.sh [OUTDIR] [extra ariadne args, e.g. -offline]
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
out="${1:-$root/work/libs/out}"
shift || true
cd "$root"
go build -o bin/ariadne ./cmd/ariadne

libs=(
  "configuration2:org.apache.commons:commons-configuration2:2.7"           # facade/builder; commons-text 1.8, commons-lang3 3.9
  "jackson-yaml:com.fasterxml.jackson.dataformat:jackson-dataformat-yaml:2.12.3" # factory/mapper subclass; snakeyaml 1.27, databind 2.12.3
  "jjwt-jackson:io.jsonwebtoken:jjwt-jackson:0.11.2"                       # SPI (META-INF/services); databind 2.9.10.4
  # Spring Boot starter (lb2-5): the library is the autoconfigure jar, its
  # dependencies what the starter brings (mybatis 3.5.4, mybatis-spring
  # 2.0.4, Spring Boot 2.2.5 jdbc).
  "mybatis-starter:org.mybatis.spring.boot:mybatis-spring-boot-starter:2.1.2:mybatis-spring-boot-autoconfigure"
)
for spec in "${libs[@]}"; do
  IFS=: read -r name g a v libA <<<"$spec"
  libA="${libA:-$a}"
  d="work/libs/$name"
  if [ ! -f "$d/lib.jar" ]; then
    mkdir -p "$d/deps"
    cat >"$d/pom.xml" <<EOF
<project xmlns="http://maven.apache.org/POM/4.0.0"><modelVersion>4.0.0</modelVersion>
<groupId>measure</groupId><artifactId>$name</artifactId><version>1</version>
<dependencies><dependency><groupId>$g</groupId><artifactId>$a</artifactId><version>$v</version></dependency></dependencies></project>
EOF
    (cd "$d" && mvn -q dependency:copy-dependencies -DoutputDirectory=all -DincludeScope=runtime)
    for f in "$d"/all/*.jar; do
      if [ "$(basename "$f")" = "$libA-$v.jar" ]; then cp "$f" "$d/lib.jar"; else cp "$f" "$d/deps/"; fi
    done
  fi
done

# Poms (LB3) and reference jars come from the local Maven repository the
# step above filled, before the cache and the network (Maven Central
# rate-limits automated clients: run once online, then with -offline).
export ARIADNE_MAVEN_LOCAL="${ARIADNE_MAVEN_LOCAL:-$HOME/.m2/repository}"
mkdir -p "$out"
for spec in "${libs[@]}"; do
  name="${spec%%:*}"
  for roots in all api; do
    for mode in cha rta; do
      base="$out/$name-$roots-$mode"
      bin/ariadne -cache-dir work/libs/cache -app "work/libs/$name/lib.jar" -deps "work/libs/$name/deps" -profile generic \
        -roots "$roots" -callgraph-mode "$mode" -json-format v2 -out "$base.md" -json "$base.json" "$@" >"$base.log" 2>&1
    done
  done
done

python3 -I - "$out" <<'PY'
import collections, json, sys
out = sys.argv[1]
for name in ["configuration2", "jackson-yaml", "jjwt-jackson", "mybatis-starter"]:
    for roots in ["all", "api"]:
        for mode in ["cha", "rta"]:
            j = json.load(open(f"{out}/{name}-{roots}-{mode}.json"))
            st = dict(collections.Counter(f["reachability"] for f in j["findings"]))
            exp = [len(f.get("exposed_api", [])) for f in j["findings"] if f["reachability"] in ("REACHABLE", "FRAMEWORK")]
            cd = sum(1 for f in j["findings"] for x in f.get("exposed_api", []) if x.get("client_dispatch"))
            cc = [(g["count"], g.get("reachable")) for g in j["coverage"].get("gaps", []) if g["mechanism"] == "client_code"]
            api = j["coverage"].get("api", {}).get("methods")
            ann = j["coverage"].get("api", {}).get("announced_classes")
            print(f"{name:15} {roots:3} {mode}: {st} api_methods={api} announced_classes={ann} exposed_sum={sum(exp)} exposed_max={max(exp) if exp else 0} "
                  f"client_code(sites,reachable)={cc} client_dispatch={cd}")
            deps = j["coverage"].get("api", {}).get("dependencies")
            if deps:
                by = dict(collections.Counter(d["supplied_by"] for d in deps["declared"]))
                absent = sum(1 for d in deps["declared"] if d.get("version_match") == "absent" and d["supplied_by"] == "library")
                vm = dict(collections.Counter(d.get("version_match") for d in deps["declared"] if d.get("instance")))
                fs = dict(collections.Counter((f["reachability"], f.get("supplied_by")) for f in j["findings"]))
                print(f"{'':15} lb3: pom={deps['pom'] or '-'} supplied_by={by} library_absent={absent} input_match={vm} notes={len(deps.get('notes', []))}")
                print(f"{'':15} lb3: findings (reachability, supplied_by)={fs}")
PY
