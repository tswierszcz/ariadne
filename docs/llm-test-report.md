# Full live test run with the LLM source

Run on 2026-10-12 against `main` at `33baf95` (LLM source + `NOT_APPLICABLE` status), model
`z-ai/glm-5.3` via OpenRouter. This page records what was run and what came out; it is a
record of one run, not a benchmark.

Full reports from the `-llm` runs (the `.md` is the human report, the `.json` the same
results as data, including each LLM result's `model`, `evidence` and `note`):

| App | Report | Results |
|---|---|---|
| spring-petclinic 2.4.5 | [llm-run/petclinic.md](llm-run/petclinic.md) | [llm-run/petclinic.json](llm-run/petclinic.json) |
| WebGoat 8.2.2 (control case) | [llm-run/webgoat.md](llm-run/webgoat.md) | [llm-run/webgoat.json](llm-run/webgoat.json) |

## What was tested

1. **`go test ./... -count=1`**: every package passes, with no JDK and no network (the LLM
   source is tested against a fake HTTP server).
2. **One live LLM test** (`TestLLM_realEndpoint`, `ARIADNE_LLM_E2E=1`): the real prompts
   sent to the real endpoint for a synthetic CVE; passed (valid reply, diff evidence).
3. **Two full end-to-end runs per application**, same inputs:
   - **baseline**, without `-llm` (sources 1 and 2 only);
   - **with `-llm`** (source 3 fills the CVEs sources 1 and 2 left unmapped).

   **Every LLM answer in this run is a fresh call**: the LLM cache from the previous run
   was moved aside first. The OSV and GitHub caches were kept, so the CVE list and the
   fix diffs are the same as in the previous run and the two can be compared.

   Commands (inputs are the Maven builds under `work/`):

   ```bash
   set -a; . ./.env; set +a
   bin/ariadne -app work/spring-petclinic/target/classes \
     -deps work/spring-petclinic/target/dependency -deps-list work/spring-petclinic/deps.txt \
     -llm -out docs/llm-run/petclinic.md -json docs/llm-run/petclinic.json
   bin/ariadne -app work/WebGoat/app-classes.jar \
     -deps work/WebGoat/app-dependency -deps-list work/WebGoat/deps.txt \
     -llm -out docs/llm-run/webgoat.md -json docs/llm-run/webgoat.json
   ```

## Results

| | petclinic baseline | petclinic `-llm` | WebGoat baseline | WebGoat `-llm` |
|---|---|---|---|---|
| CVEs | 148 | 148 | 160 | 160 |
| mapped to methods | 56 (manual 1, diff 55) | 95 (+39 llm) | 104 (manual 37, diff 67) | 110 (+6 llm) |
| REACHABLE | 13 | 13 | 51 | 52 |
| FRAMEWORK | 18 | 49 | 32 | 33 |
| NOT_REACHABLE | 25 | 32 | 21 | 23 |
| UNKNOWN | 86 | 48 | 56 | 52 |
| NOT_APPLICABLE | 6 | 6 | 0 | 0 |
| wall time | 6 s | 1 m 16 s* | 15 s | 47 s |

\* The first petclinic `-llm` attempt was killed partway through (exit 137, on a machine
with ~1 GB free memory) after caching 37 answers; the rerun reused those, so its time
covers the remaining 55 calls only.

### NOT_APPLICABLE

petclinic's 6 `org.webjars:bootstrap` CVEs are `NOT_APPLICABLE`: the jar is a WebJar
(JavaScript/CSS, no JVM bytecode). This does not depend on `-llm`. Empty Maven aggregator
jars are deliberately not treated this way: Spring4Shell (CVE-2022-22965) is filed against
`spring-boot-starter-web`, which has no classes, and it keeps its normal status.

### What the LLM source did

- **132 distinct prompts** (identical prompts shared by both apps are answered once), **0
  failed requests, 0 invalid replies**, about **$0.20** at OpenRouter's reported cost.
- By what the model was shown:

  | Evidence | Prompts | Methods named | Empty (UNKNOWN) |
  |---|---|---|---|
  | fix diff (variant A) | 39 | 39 | 0 |
  | commit with no non-test Java (variant B) | 5 | 0 | 5 |
  | advisory text only, no commit (variant C) | 88 | 10 | 78 |

- **Grounding in bytecode dropped 21 methods (petclinic) and 10 (WebGoat)** that the model
  named but that do not exist in the analyzed jars.
- **Option C:** 3 results (1 petclinic, 2 WebGoat) were `NOT_REACHABLE` on advisory-only
  evidence and are reported as `UNKNOWN` with a note.
- **Where the status changes against the baseline came from:**
  - petclinic: 38 CVEs left UNKNOWN: 30 to FRAMEWORK and 7 to NOT_REACHABLE with diff
    evidence, 1 to FRAMEWORK on advisory evidence. 30 of the 31 new FRAMEWORK results are
    Tomcat CVEs, which source 2 cannot map because of its `src/main/` filter (see the
    deferred follow-up in [PLAN.md](../PLAN.md)).
  - WebGoat: 4 CVEs left UNKNOWN: 1 FRAMEWORK and 2 NOT_REACHABLE with diff evidence, 1
    REACHABLE on advisory evidence (CVE-2022-22978).

## Repeatability: this run against the previous one

The same 132 prompts were sent in an earlier run the same day (`temperature: 0` both
times). Comparing the answers prompt by prompt:

| Evidence | Identical answer | Same methods partly (overlap) | Empty in one run only | No overlap |
|---|---|---|---|---|
| fix diff (A) | 24 | 15 | 0 | 0 |
| commit with no Java (B) | 5 | 0 | 0 | 0 |
| advisory only (C) | 78 | 1 | 7 | 2 |

- **With a diff the answers are stable**: every answer shares methods with the previous
  run; the differences are extra or missing entry points around the same vulnerable method.
- **Advisory-only answers are not stable**: in 9 of 88 the model named methods in one run
  and nothing in the other, or entirely different methods. This is the evidence behind
  never letting an advisory-only answer produce NOT_REACHABLE (option C).
- Effect on statuses: petclinic's status counts are identical to the previous run. WebGoat
  changed in 2 CVEs, both advisory-only:
  - CVE-2026-47838: REACHABLE → UNKNOWN. The model named
    `SubjectDnX509PrincipalExtractor#extractPrincipal` last time and `#getPrincipal` this
    time; the latter does not exist in the bytecode, so grounding dropped it and nothing
    was left.
  - CVE-2024-3653: FRAMEWORK → UNKNOWN. The model named a method last time and returned an
    empty list this time.

## What to read into it, and what not

- The control case still holds: 36 of WebGoat's REACHABLE results are manual-table XStream
  entries (the LLM never overrides the manual table).
- **None of the LLM's method lists were checked by hand.** Grounding shows a method
  exists, not that it is the vulnerable one.
- **Treat advisory-only REACHABLE results as leads, not findings.** The one in this run
  (WebGoat CVE-2022-22978, via `RegexRequestMatcher#matches`) should be checked against
  its call path in the report before relying on it; a similar one from the previous run did
  not reproduce.
- One model, two runs, no ground truth for the unmapped CVEs: this shows the source works
  end to end, how much it moves the statuses, and how repeatable it is; not how accurate
  it is.
