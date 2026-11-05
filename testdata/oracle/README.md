# ServiceLoader real-JVM oracle (SERVICE_LOADING S4, ROADMAP B4)

`expected.tsv` records which providers the **real JVM** instantiated for each
fixture entry method in each launch scenario; `internal/serviceload`'s
`TestServiceLoaderOracle` runs ariadne's model on the same bytecode and
launch and compares:

- a provider the JVM created that the model does not connect fails the test,
  unless `model-gaps.tsv` declares it a known gap **and** the analysis reports
  itself Incomplete for that entry (no silent misses);
- a provider the model connects that the JVM did not create fails the test,
  unless `model-extras.tsv` declares it with a reason (a static may-analysis
  over-approximates);
- every line in either file must still apply (stale lines fail).

Columns of `expected.tsv`: scenario, `Class#method`, sorted provider classes
(`-` for none), optionally `ERROR <exception>` when the entry threw (e.g.
`ServiceConfigurationError` for a missing `uses`, an NPE for a null field).

Ordinary `go test` reads the checked-in files and needs no JDK. Regenerate
after changing a fixture (compile it first, see `testdata/serviceloader` and
`testdata/jpms`):

```sh
scripts/serviceloader-oracle.sh
```

The script compiles `src/oracle/OracleRunner.java` (outside every analyzed
input), runs the entries with `java` in the class-path, module-path
(`--add-modules ALL-MODULE-PATH`, the automatic module packed as a jar) and
mixed launches, and collects the `ORACLE-PROVIDER` lines the fixture
providers print from their constructors or `provider()` methods.
