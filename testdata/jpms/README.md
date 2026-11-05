# JPMS compiled fixture (SERVICE_LOADING S3)

`classes/` holds checked-in javac output read by
`internal/classfile/module_test.go` and `internal/serviceload/modules_test.go`.
The Go tests pack these directories into jars in a temp dir; they do not need
a JDK.

| Directory | Becomes in tests | Purpose |
|---|---|---|
| `classes/mods/fixture.jpms.api` | explicit module | `exports fixture.jpms.api` (`Codec`) |
| `classes/mods/fixture.jpms.impl` | explicit module | `provides Codec with ModuleCodec, FactoryCodec`; `META-INF/services` lists `ModuleCodec`, `LegacyCodec` — one mismatch in each direction; `FactoryCodec` has `public static provider()` |
| `classes/mods/fixture.jpms.app` | explicit module | `uses Codec`, calls `ServiceLoader.load(Codec.class)` |
| `classes/auto` | automatic module (manifest name added by the test, or a file-name-derived name) | `AutoCodec` via `META-INF/services` |
| `classes/mr` | multi-release jar, `module-info.class` only under `META-INF/versions/9/` | descriptor selection by `-target-java` |

Regenerate from the repository root (this JDK has no `ct.sym`, so
`-source/-target` instead of `--release`):

```sh
J="java --module jdk.compiler/com.sun.tools.javac.Main"
cd testdata/jpms && rm -rf classes
$J -source 11 -target 11 -nowarn -d classes/mods --module-source-path src \
  --module fixture.jpms.api,fixture.jpms.impl,fixture.jpms.app
$J -source 8 -target 8 -nowarn -cp classes/mods/fixture.jpms.api -d classes/auto \
  src/auto/fixture/jpms/auto/AutoCodec.java
$J -source 8 -target 8 -nowarn -d classes/mr src/mr/base/fixture/jpms/mr/Util.java
$J -source 9 -target 9 -nowarn -d classes/mr/META-INF/versions/9 \
  --patch-module fixture.jpms.mr=classes/mr src/mr/versions9/module-info.java
cp -R src/fixture.jpms.impl/META-INF classes/mods/fixture.jpms.impl/
cp -R src/auto/META-INF classes/auto/
cp -R src/mr/META-INF classes/mr/
```
