# ServiceLoader compiled fixture

`classes/` contains checked-in Java 8 classfiles and the matching
`META-INF/services` declaration used by `TestCompiledServiceLoaderFixture`.
The ordinary Go test reads these files directly; it does not need a JDK.

To regenerate them with a JDK, run from the repository root (a modern JDK
that still supports Java 8 source/target):

```sh
mkdir -p testdata/serviceloader/classes
java --module jdk.compiler/com.sun.tools.javac.Main -source 8 -target 8 \
  -d testdata/serviceloader/classes \
  testdata/serviceloader/src/fixture/service/Decoder.java \
  testdata/serviceloader/src/fixture/provider/DecoderProvider.java \
  testdata/serviceloader/src/fixture/app/Entry.java
cp -R testdata/serviceloader/src/META-INF \
  testdata/serviceloader/classes/
```

The checked-in classfiles were compiled from these sources with the JDK
compiler. They use classfile major version 52 (Java 8).

The fixture covers helper-return provenance, a direct static helper that
returns its loader parameter, static and final instance loader fields, an
unused lazy load, an unknown helper/parameter caller, a mutable instance
field, and cyclic static-field summaries. Unsupported cases must remain
explicit gaps without provider edges.

For Stream API coverage it includes the linear
`stream().findFirst().get().get()` path, a returned stream with no terminal
operation (lazy, no provider construction), and a `map` callback that calls
`Provider.get()` (reported as unsupported, with no guessed provider target).

`DecoderProvider`'s constructor prints `ORACLE-PROVIDER fixture.provider.DecoderProvider`
for the real-JVM oracle in `testdata/oracle` (SERVICE_LOADING S4).
