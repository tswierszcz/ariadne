# Supported capabilities and limitations

This document describes the current implementation of Ariadne. Future requirements live in [RECOMMENDATION](../RECOMMENDATION/README.MD); their presence does not mean a feature is available.

Last source review: 2026-10-15. This initial inventory is based on repository source inspection. Tests and live analyses were not run to produce this inventory. Links to test files identify existing verification material, not new passing results. Future implementation sessions must record the checks they actually perform.

## Reading the status

- **Implemented:** the specific behavior described is present; the label does not cover every variation of the mechanism.
- **Partial:** a defined subset exists, with material gaps listed.
- **Experimental:** available behavior undergoing qualification, with its conditions stated.
- **Not implemented:** the current pipeline does not provide that behavior.
- **Unassessed:** evidence is insufficient to state the current scope.

Implementation status, tested versions and evidence quality are separate. A completed design phase does not change product support.

## Input and dependency identity

| Capability | Status | Available scope and conditions | Limits and evidence |
|---|---|---|---|
| Classes directory and plain application JAR | Implemented | `-app` accepts a directory or plain JAR; dependencies can come from `-deps`. | Dependency directory is scanned for top-level JAR entries. [Input normalization](../internal/input/source.go), [dependency resolution](../internal/input/resolve.go). |
| Spring Boot fat JAR and WAR | Implemented | Recognizes the implemented BOOT-INF/WEB-INF layouts and extracts classes/libraries. | Extracted libraries replace the `-deps` source in the current CLI; no promise of arbitrary archive layouts. [Input](../internal/input/source.go), [CLI](../cmd/ariadne/main.go). |
| EAR | Partial | Merges nested WAR application classes/libraries and treats top-level JARs as dependencies. | EJB JAR classes are not application roots; class/library names can collide and module isolation is not modeled. [Input](../internal/input/source.go). |
| Maven component identity | Partial | Embedded `pom.properties`, then dependency-list filename fallback; unidentified JARs are reported. | One selected GAV per dependency JAR; multiple metadata entries do not establish complete embedded-component inventory. [Resolver](../internal/input/resolve.go). |
| Shaded/relocated components | Not implemented | Multiple POM entries may be detected as metadata candidates. | No complete component-to-class ownership or relocation model. [Resolver](../internal/input/resolve.go), [planned CVE work](../RECOMMENDATION/CVE.MD). |

## Bytecode, graph and framework knowledge

| Capability | Status | Available scope and conditions | Limits and evidence |
|---|---|---|---|
| Pure-Go class parsing and invoke extraction | Implemented | Reads the supported classfile structures and extracts implemented invoke instructions without running a JDK. | Java 8/21 fixture JARs exist; this is not qualification of every compiler, language feature or JVM version. [Parser](../internal/classfile/classfile.go), [calls](../internal/classfile/calls.go), [tests](../internal/classfile/classfile_test.go). |
| Direct calls, CHA and inheritance | Partial | Direct symbolic edges, subtype expansion, and inheritance-aware vulnerable-method matching. | CHA and vulnerable-method matching use names without exact overload selection; inherited bodies can remain disconnected; JDK-prefix expansion is skipped. [Builder](../internal/callgraph/build.go), [skip](../internal/callgraph/skip.go), [matching](../internal/reach/match.go). |
| Duplicate class selection | Partial | Keeps one definition and reports duplicates. | Current parallel read completion can affect the selected class; full loader isolation is absent. [Hierarchy build](../internal/hierarchy/build.go). |
| Implicit class initialization | Not implemented | Application `<clinit>` methods may be roots under `-roots all`, as other app methods are. | No general initialization edges from active-use operations to dependency initializers. This is distinct from treating an app initializer as a root. [Calls](../internal/classfile/calls.go), [roots](../internal/reach/reach.go). |
| Multi-release target selection | Not implemented | Reads base classes only. | Skips `META-INF/versions/`; no target-Java selector. [JAR reader](../internal/hierarchy/jarread.go). |
| Lambdas and method references | Partial | Recognizes LambdaMetafactory/altMetafactory implementation handles. | Current graph adds a target at creation and does not fully preserve invocation timing or virtual handle dispatch. Other bootstraps are counted as unresolved. [Calls](../internal/classfile/calls.go), [builder](../internal/callgraph/build.go). |
| General reflection, MethodHandle invocation and custom bootstrap models | Not implemented | API calls can appear as ordinary symbolic calls. | No general reconstruction of reflective targets, handle adaptations or arbitrary bootstrap results. [Calls](../internal/classfile/calls.go). |
| Profiles and selected roots | Implemented | Auto-detection/merging and explicit profiles: spring, servlet, jakarta, micronaut, quarkus, generic; `-roots all` or `profile`. | This is profile knowledge, not full runtime framework emulation. [Profile code](../internal/profile/profile.go), [profile tests](../internal/profile/profile_test.go). |
| Spring Data synthetic dispatch | Partial | The active profile can connect declared subtype methods of Repository to same-named SimpleJpaRepository methods. | A narrow bridge; not complete support for every repository, inherited operation, proxy or AOP chain. [Synthetic edges](../internal/callgraph/synthetic.go), [Spring profile](../internal/profile/profiles/spring.yaml). |
| Annotation names | Partial | Reads implemented runtime-visible class/method/parameter annotation names for profiles. | Values are discarded; constructor synthetic parameter alignment has a documented gap. [Annotations](../internal/classfile/annotations.go). |
| DI/AOP/configuration execution models | Not implemented | Existing annotations, profiles and the narrow Spring Data bridge supply some static hints. | No general bean binding, pointcut, active-environment, web.xml or configuration-driven execution model. [Profile](../internal/profile/profile.go). |
| ServiceLoader and JPMS service models | Not implemented | No service-provider execution model in the pipeline. | module-info is skipped by hierarchy loading; resource registration is not modeled. [JAR reader](../internal/hierarchy/jarread.go), [synthetic rules](../internal/callgraph/synthetic.go). |
| General callbacks and asynchronous execution models | Not implemented | Some targets may appear through ordinary bytecode calls or the current lambda approximation. | No dedicated task/listener/future/reactive registration-to-execution model. [Graph builder](../internal/callgraph/build.go). |
| RTA and points-to | Not implemented | Current builder uses CHA with configured skip behavior. | No allocation-driven RTA or receiver-specific points-to analysis. [Builder](../internal/callgraph/build.go). |
| Data flow and taint | Not implemented | Reachability searches method-call edges. | No stack/local/heap value-flow engine, source-to-sink data path or general sanitizer model. [Code parser](../internal/classfile/code.go), [BFS](../internal/reach/bfs.go). |

## Vulnerability data, mapping and reports

| Capability | Status | Available scope and conditions | Limits and evidence |
|---|---|---|---|
| OSV Maven queries and alias merging | Implemented | Queries resolved package versions and fetches advisory details, then merges aliases. | Coverage depends on identity and source data; extra-page fetching is bounded. [OSV client](../internal/cves/cves.go). |
| Manual and fix-diff method mapping | Partial | Manual map takes priority; otherwise scans non-test Java source changes from recognized GitHub commits. | Name-only methods; no version-scoped exact mapping schema. Diff mapping uses post-fix changed lines and can miss deletions; touched method is not proof of vulnerability. [Resolver](../internal/vulnmap/resolve.go), [diff source](../internal/vulnmap/diffsource.go), [diff parser](../internal/javadiff/diff.go). |
| Optional LLM mapping | Partial | `-llm`, configured endpoint/model; suggestions checked for method existence. Advisory-only negative reachability is downgraded. | Existence does not verify vulnerability or complete mapping; source quality remains distinct. [LLM grounding](../internal/vulnmap/llmanswer.go), [reachability](../internal/reach/reach.go). |
| Disk cache and offline | Implemented | Cached HTTP bodies and explicit offline cache misses; existing optional LLM namespaces. | No TTL, freshness metadata, force-refresh or full snapshot replay. [Cache](../internal/httpcache/cache.go), [fetch](../internal/httpcache/fetch.go). |
| Reachability statuses | Implemented | REACHABLE, FRAMEWORK, NOT_REACHABLE, UNKNOWN and WebJar NOT_APPLICABLE; two graph passes. | Results reflect model assumptions and gaps. Current missing method matches may still become NOT_REACHABLE; see planned CVE changes. [Engine](../internal/reach/reach.go). |
| Markdown and JSON reports | Implemented | Writes current report and result list; includes paths, source labels and selected warnings. | Current output is not the proposed component-scoped JSON v2. [Rendering](../internal/report/render.go), [template](../internal/report/templates/report.md.tmpl). |
| Exploit-check and framework-check | Implemented | Optional LLM annotations about existing REACHABLE/FRAMEWORK paths. | Heuristic opinions, not taint analysis, observed execution or proof of exploitation. [Checks](../internal/exploit/exploit.go), [framework checker](../internal/exploit/framework.go). |
| Deserialization callback/gadget analysis | Not implemented | Manual mappings include selected deserialization entry points. | Reaching an entry API does not reconstruct a gadget chain, its data or filter conditions. [Manual map](../internal/vulnmap/vuln_methods.yaml). |
| Target JDK vulnerability assessment | Not implemented | Reports max classfile version seen. | Bytecode version is not runtime vendor/build or security-patch level; no target-JDK advisory adapter. [CLI](../cmd/ariadne/main.go). |
| Runtime tracing, dynamic loader identity and JNI models | Not implemented | Static analysis of supplied classes and dependencies. | No trace importer/agent, full dynamic-class inventory or native callback model. [Hierarchy](../internal/hierarchy/hierarchy.go). |
| Other ecosystems, including Go | Not implemented | Current CLI analyzes JVM/Maven inputs. | No govulncheck import or non-JVM frontend. [CLI](../cmd/ariadne/main.go). |

## Interpretation and maintenance

`NOT_REACHABLE` means no path was found under the implemented model and chosen roots. It is not a guarantee that the application is safe. `NOT_APPLICABLE` identifies the implemented WebJar-only case; browser vulnerabilities may still matter. A profile match, configured registration, static path, LLM opinion and runtime observation are different evidence types.

The source inventory above does not claim fresh test passes or new live runs. For future changes, attach the actual test/scenario evidence and update exact supported subsets, versions, flags and limitations. Update README and report limitations when their claims change; preserve unsupported cases in the description. Phase progress belongs in [IMPLEMENTATION_STATUS.MD](../RECOMMENDATION/IMPLEMENTATION_STATUS.MD).
