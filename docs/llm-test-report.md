# Full test run with the LLM source (Phase 4, source 3)

Run on 2026-10-12, branch `phase-4-llm-report` (code as of `main` at `c28b57a`), model
`z-ai/glm-5.3` via OpenRouter. This page says what was run and what came out; it is
a record of one run, not a benchmark.

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
3. **Two full end-to-end runs per application**, same inputs, same warm OSV/GitHub cache:
   - **baseline**, without `-llm` (sources 1 and 2 only);
   - **with `-llm`** (source 3 fills the CVEs sources 1 and 2 left unmapped).

   Commands (inputs are the Maven builds under `work/`):

   ```bash
   bin/ariadne -app work/spring-petclinic/target/classes \
     -deps work/spring-petclinic/target/dependency -deps-list work/spring-petclinic/deps.txt \
     -llm -out docs/llm-run/petclinic.md -json docs/llm-run/petclinic.json
   bin/ariadne -app work/WebGoat/app-classes.jar \
     -deps work/WebGoat/app-dependency -deps-list work/WebGoat/deps.txt \
     -llm -out docs/llm-run/webgoat.md -json docs/llm-run/webgoat.json
   ```

   The baseline runs reproduced the numbers recorded earlier in PLAN.md exactly.

## Results

| | petclinic baseline | petclinic `-llm` | WebGoat baseline | WebGoat `-llm` |
|---|---|---|---|---|
| CVEs | 148 | 148 | 160 | 160 |
| mapped to methods | 56 (manual 1, diff 55) | 99 (+43 llm) | 104 (manual 37, diff 67) | 113 (+9 llm) |
| REACHABLE | 13 | 13 | 51 | 53 |
| FRAMEWORK | 18 | 49 | 32 | 34 |
| NOT_REACHABLE | 25 | 32 | 21 | 23 |
| UNKNOWN | 92 | 54 | 56 | 50 |
| wall time | 6.7 s | 2 m 18 s | n/a | 52 s |

### What the LLM source did

- **132 distinct prompts** (identical prompts shared by both apps are answered once), **0
  failed requests, 0 invalid replies**, about **$0.24** in total at OpenRouter's reported
  cost.
- By what the model was shown:

  | Evidence | Prompts | Methods named | Empty (UNKNOWN) |
  |---|---|---|---|
  | fix diff (variant A) | 39 | 39 | 0 |
  | commit with no non-test Java (variant B) | 5 | 0 | 5 |
  | advisory text only, no commit (variant C) | 88 | 13 | 75 |

  With a diff the model always named methods. Without one it returned an empty list for 75
  of 88, which is what rule 4 of the prompt asks for when the advisory does not name a
  method.
- **Grounding in bytecode dropped 17 methods (petclinic) and 6 (WebGoat)** that the model
  named but that do not exist in the analyzed jars, so none of those could become a false
  verdict.
- **Option C (advisory-only results are never reported as safe):** 8 results (5 petclinic,
  3 WebGoat) were `NOT_REACHABLE` on advisory-only evidence and were therefore reported as
  `UNKNOWN` with a note.
- **Where the status changes came from:** petclinic 38 CVEs left UNKNOWN (30 to FRAMEWORK,
  7 to NOT_REACHABLE with diff evidence, 1 to FRAMEWORK on advisory evidence); WebGoat 6
  (1 FRAMEWORK, 2 NOT_REACHABLE, 3 on advisory evidence: 1 FRAMEWORK, 2 REACHABLE).

## What to read into it, and what not

- The control case still holds: 36 of WebGoat's REACHABLE results are manual-table XStream
  entries (the LLM never overrides the manual table).
- **None of the LLM's method lists were checked by hand.** Grounding shows a method
  exists, not that it is the vulnerable one.
- 30 of the 31 new FRAMEWORK results on petclinic are Tomcat CVEs (`org.apache.catalina`,
  `coyote`, `tomcat`), read from diffs. Source
  2 cannot map them because its `src/main/` filter hides Tomcat's `java/org/apache/...`
  layout (see the deferred follow-up in [PLAN.md](../PLAN.md)); the model read them instead.
- **Two new WebGoat REACHABLE results rest on advisory text only** (CVE-2022-22978, via
  `RegexRequestMatcher#matches`; CVE-2026-47838, via
  `SubjectDnX509PrincipalExtractor#extractPrincipal`). A reachable verdict from advisory
  evidence is shown on purpose, but these two are exactly the ones worth a manual look at
  the call path in the report before relying on them.
- One model, one run, no ground truth for the unmapped CVEs: this shows the source works
  end to end and how much it moves the statuses, not how accurate it is.
