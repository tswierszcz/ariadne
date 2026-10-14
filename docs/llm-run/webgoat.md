# ariadne reachability report

- **App:** work/WebGoat/app-classes.jar
- **Generated:** 2026-10-14 13:32:30 CEST
- **ariadne version:** v0.0.0-20261014100226-85b4009bd1c9+dirty
- **Profile:** spring+servlet
- **Max bytecode version seen:** 59
- **Call graph:** 339786 nodes, 2592178 edges (737799 direct, 1853605 added by CHA)
- **Unresolved `invokedynamic` sites:** 13
- **Dependencies:** 126 total, 126 identified
- **LLM-assisted mapping:** z-ai/glm-5.3 (6 CVEs mapped, 0 skipped)
- **Exploit-check:** z-ai/glm-5.3 (40 REACHABLE CVEs checked, 0 skipped)
- **Framework-check:** z-ai/glm-5.3 (45 FRAMEWORK CVEs checked, 0 skipped)

## Summary

**160 CVEs analyzed.**

| Status | Count |
|---|---|
| REACHABLE | 40 |
| FRAMEWORK | 45 |
| NOT_REACHABLE | 23 |
| UNKNOWN | 52 |
| NOT_APPLICABLE | 0 |

| Severity | Count |
|---|---|
| CRITICAL | 16 |
| HIGH | 72 |
| LOW | 9 |
| MODERATE | 63 |


## CVEs

| CVE | Package | Version | Severity | Status | Vulnerable method(s) | Source |
|---|---|---|---|---|---|---|
| CVE-2013-7285 | com.thoughtworks.xstream:xstream | 1.4.5 | CRITICAL | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2016-1000027 | org.springframework:spring-web | 5.3.4 | CRITICAL | NOT_REACHABLE | `org.springframework.remoting.httpinvoker.HttpInvokerServiceExporter#handleRequest` | llm (diff) |
| CVE-2016-3674 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2017-7957 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2020-14340 | org.jboss.xnio:xnio-nio | 3.8.0.Final | MODERATE | UNKNOWN |  |  |
| CVE-2020-26217 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2020-26258 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2020-26259 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2020-36518 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#createContextual`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapArrayToArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#<init>`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeAnyScalar`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeNR`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeObjectAtName`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_mapObjectWithDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_putValueHandleDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_squashDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#addValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#childArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#childObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#deserialize`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#deserializeWithType`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#emptyList`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#emptyMap`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishBranchArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishBranchObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishRootArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishRootObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#instance`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#isObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#logicalType`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#putDeferredValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#putValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#resetAsArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#resetAsObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#rootArrayScope`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#rootObjectScope`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#supportsUpdate` | diff |
| CVE-2021-21341 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21342 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21343 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21344 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21345 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21346 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21347 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21348 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21349 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21350 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-21351 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-22060 | org.springframework:spring-core | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2021-22096 | org.springframework:spring-core | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2021-22118 | org.springframework:spring-web | 5.3.4 | HIGH | FRAMEWORK | `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#getBlockingOperationScheduler`, `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#read`, `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#setFileStorageDirectory`, `org.springframework.http.codec.multipart.FileStorage#<init>`, `org.springframework.http.codec.multipart.FileStorage#createNewDirectoryIfDeleted`, `org.springframework.http.codec.multipart.FileStorage#directory`, `org.springframework.http.codec.multipart.FileStorage#fromPath`, `org.springframework.http.codec.multipart.FileStorage#tempDirectory`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#<init>`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#accept`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#createTempDirectory`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#read`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#setFileStorageDirectory` | diff |
| CVE-2021-22119 | org.springframework.security:spring-security-core | 5.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2021-28170 | org.glassfish:jakarta.el | 3.0.3 | MODERATE | UNKNOWN |  |  |
| CVE-2021-29505 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-3597 | io.undertow:undertow-core | 2.2.4.Final | MODERATE | UNKNOWN |  |  |
| CVE-2021-3629 | io.undertow:undertow-core | 2.2.4.Final | HIGH | UNKNOWN |  |  |
| CVE-2021-3690 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.websockets.jsr.FrameHandler#onFullPongMessage` | diff |
| CVE-2021-3859 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.protocols.http2.HpackEncoder#encode`, `io.undertow.protocols.http2.Http2Channel#parseFrame`, `io.undertow.protocols.http2.Http2Channel#parseFrameNoContinuation` | diff |
| CVE-2021-39139 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39140 | com.thoughtworks.xstream:xstream | 1.4.5 | MODERATE | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39141 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39144 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39145 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39146 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39147 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39148 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39149 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39150 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39151 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39152 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39153 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-39154 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-42550 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | FRAMEWORK | `ch.qos.logback.classic.selector.ContextJNDISelector#findConfigFileURL`, `ch.qos.logback.classic.util.JNDIUtil#lookup`, `ch.qos.logback.core.util.OptionHelper#isEmpty` | diff |
| CVE-2021-43466 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2021-43859 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2021-46877 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.databind.node.NodeSerialization#_read`, `com.fasterxml.jackson.databind.node.NodeSerialization#readExternal` | diff |
| CVE-2022-1471 | org.yaml:snakeyaml | 1.27 | HIGH | FRAMEWORK | `org.yaml.snakeyaml.Yaml#load`, `org.yaml.snakeyaml.Yaml#loadAll`, `org.yaml.snakeyaml.Yaml#loadAs`, `org.yaml.snakeyaml.constructor.Constructor#<init>` | manual |
| CVE-2022-2053 | io.undertow:undertow-core | 2.2.4.Final | HIGH | UNKNOWN |  |  |
| CVE-2022-22950 | org.springframework:spring-expression | 5.3.4 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ast.ConstructorReference#checkNumElements`, `org.springframework.expression.spel.ast.ConstructorReference#createArray` | diff |
| CVE-2022-22965 | org.springframework.boot:spring-boot-starter-web | 2.4.3 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22965 | org.springframework:spring-beans | 5.3.4 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22965 | org.springframework:spring-webmvc | 5.3.4 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22968 | org.springframework:spring-context | 5.3.4 | HIGH | NOT_REACHABLE | `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields` | diff |
| CVE-2022-22970 | org.springframework:spring-beans | 5.3.4 | HIGH | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces`, `org.springframework.beans.CachedIntrospectionResults#isInvalidReadOnlyPropertyType`, `org.springframework.beans.CachedIntrospectionResults#isPlainAccessor` | diff |
| CVE-2022-22976 | org.springframework.security:spring-security-core | 5.4.5 | MODERATE | REACHABLE | `org.springframework.security.crypto.bcrypt.BCrypt#char64`, `org.springframework.security.crypto.bcrypt.BCrypt#checkpw`, `org.springframework.security.crypto.bcrypt.BCrypt#crypt_raw`, `org.springframework.security.crypto.bcrypt.BCrypt#decode_base64`, `org.springframework.security.crypto.bcrypt.BCrypt#ekskey`, `org.springframework.security.crypto.bcrypt.BCrypt#gensalt`, `org.springframework.security.crypto.bcrypt.BCrypt#hashpw`, `org.springframework.security.crypto.bcrypt.BCrypt#hashpwforcheck`, `org.springframework.security.crypto.bcrypt.BCrypt#init_key`, `org.springframework.security.crypto.bcrypt.BCrypt#key`, `org.springframework.security.crypto.bcrypt.BCrypt#streamtoword`, `org.springframework.security.crypto.bcrypt.BCrypt#streamtoword_bug`, `org.springframework.security.crypto.bcrypt.BCrypt#streamtowords`, `org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder#<init>`, `org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder#encode`, `org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder#getVersion` | diff |
| CVE-2022-22978 | org.springframework.security:spring-security-core | 5.4.5 | CRITICAL | FRAMEWORK | `org.springframework.security.web.util.matcher.RegexRequestMatcher#matches` | llm (advisory) |
| CVE-2022-22978 | org.springframework.security:spring-security-web | 5.4.5 | CRITICAL | FRAMEWORK | `org.springframework.security.web.util.matcher.RegexRequestMatcher#matches` | llm (advisory) |
| CVE-2022-25857 | org.yaml:snakeyaml | 1.27 | HIGH | FRAMEWORK | `org.yaml.snakeyaml.LoaderOptions#getNestingDepthLimit`, `org.yaml.snakeyaml.LoaderOptions#setNestingDepthLimit`, `org.yaml.snakeyaml.composer.Composer#<init>`, `org.yaml.snakeyaml.composer.Composer#composeNode`, `org.yaml.snakeyaml.composer.Composer#decreaseNestingDepth`, `org.yaml.snakeyaml.composer.Composer#increaseNestingDepth` | diff |
| CVE-2022-38749 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38750 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38751 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38752 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-40151 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2022-41853 | org.hsqldb:hsqldb | 2.5.1 | CRITICAL | UNKNOWN |  |  |
| CVE-2022-41854 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-41966 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2022-42003 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.BeanDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.BeanDeserializerBase#_delegateDeserializer`, `com.fasterxml.jackson.databind.deser.BeanDeserializerBase#deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.DateDeserializers#_parseDate`, `com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.MapEntryDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.NumberDeserializers#deserialize`, `com.fasterxml.jackson.databind.deser.std.StackTraceElementDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeFromEmpty`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeWrappedValue`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBoolean`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBooleanPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseByte`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBytePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDate`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDateFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDouble`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDoublePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloat`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloatPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseIntPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseInteger`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLong`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLongPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShort`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShortPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseString`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#handleNestedArrayForSingle`, `com.fasterxml.jackson.databind.deser.std.StringDeserializer#deserialize` | diff |
| CVE-2022-42004 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.BeanDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.DateDeserializers#_parseDate`, `com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.NumberDeserializers#deserialize`, `com.fasterxml.jackson.databind.deser.std.StackTraceElementDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeWrappedValue`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBoolean`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBooleanPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseByte`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBytePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDate`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDateFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDouble`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDoublePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloat`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloatPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseIntPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseInteger`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLong`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLongPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShort`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShortPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseString`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#handleNestedArrayForSingle`, `com.fasterxml.jackson.databind.deser.std.StringDeserializer#deserialize` | diff |
| CVE-2022-4492 | io.undertow:undertow-core | 2.2.4.Final | CRITICAL | UNKNOWN |  |  |
| CVE-2023-1108 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.protocols.ssl.SslConduit#wrapAndFlip` | diff |
| CVE-2023-1932 | org.hibernate.validator:hibernate-validator | 6.1.7.Final | MODERATE | UNKNOWN | `org.hibernate.validator.internal.constraintvalidators.hv.SafeHtmlValidator#isValid` | llm (advisory) |
| CVE-2023-1973 | io.undertow:undertow-core | 2.2.4.Final | MODERATE | FRAMEWORK | `io.undertow.security.impl.FormAuthenticationMechanism#handleRedirectBack`, `io.undertow.security.impl.FormAuthenticationMechanism#storeInitialLocation`, `io.undertow.servlet.handlers.security.ServletFormAuthenticationMechanism#handleRedirectBack`, `io.undertow.servlet.handlers.security.ServletFormAuthenticationMechanism#storeInitialLocation` | diff |
| CVE-2023-20860 | org.springframework:spring-webmvc | 5.3.4 | CRITICAL | FRAMEWORK | `org.springframework.web.servlet.handler.HandlerMappingIntrospector#<init>`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#afterPropertiesSet`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#doWithHandlerMapping`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#doWithHandlerMappingIgnoringException`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#getCorsConfiguration`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#getMatchableHandlerMapping`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#initFallback`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#initHandlerMappings`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#match`, `org.springframework.web.servlet.handler.PathPatternMatchableHandlerMapping#<init>` | diff |
| CVE-2023-20861 | org.springframework:spring-expression | 5.3.4 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ast.OperatorMatches#<init>`, `org.springframework.expression.spel.ast.OperatorMatches#getValueInternal`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#eatRelationalExpression` | diff |
| CVE-2023-20863 | org.springframework:spring-expression | 5.3.4 | HIGH | FRAMEWORK | `org.springframework.expression.spel.standard.InternalSpelExpressionParser#checkExpressionLength`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#doParseExpression` | diff |
| CVE-2023-20873 | org.springframework.boot:spring-boot-actuator-autoconfigure | 2.4.3 | CRITICAL | NOT_REACHABLE | `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.CloudFoundryWebFluxEndpointHandlerMapping#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.CloudFoundryWebFluxEndpointHandlerMapping#getAllEndpoints`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#cloudFoundryWebFluxEndpointHandlerMapping`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#getPaths`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#postProcess`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#postProcessAfterInitialization`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#webFilterChainPostProcessor`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#cloudFoundryWebEndpointServletHandlerMapping`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#customize`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#ignoreCloudFoundryPathsWebSecurityCustomizer`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryWebEndpointServletHandlerMapping#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryWebEndpointServletHandlerMapping#getAllEndpoints` | diff |
| CVE-2023-20883 | org.springframework.boot:spring-boot-autoconfigure | 2.4.3 | HIGH | FRAMEWORK | `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#createWelcomePageHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#getIndexHtmlResource`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#welcomePageHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#welcomePageNotAcceptableHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#getViewName`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#isTemplated`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#resolve`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#getHandlerInternal`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#isHtmlTextAccepted`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#getHandlerInternal`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#handleRequest` | diff |
| CVE-2023-34055 | org.springframework.boot:spring-boot-actuator | 2.4.3 | MODERATE | FRAMEWORK | `org.springframework.boot.actuate.metrics.web.reactive.server.WebFluxTags#method`, `org.springframework.boot.actuate.metrics.web.servlet.WebMvcTags#method` | diff |
| CVE-2023-4639 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.util.Cookies#parseCookie` | diff |
| CVE-2023-5685 | org.jboss.xnio:xnio-api | 3.8.0.Final | HIGH | FRAMEWORK | `org.xnio.AbstractIoFuture#<init>`, `org.xnio.AbstractIoFuture#cancel`, `org.xnio.AbstractIoFuture#cancelRequested`, `org.xnio.AbstractIoFuture#doCancel`, `org.xnio.AbstractIoFuture#doNotifyCancelled`, `org.xnio.AbstractIoFuture#doNotifyDone`, `org.xnio.AbstractIoFuture#doNotifyFailed`, `org.xnio.AbstractIoFuture#isCancelRequested`, `org.xnio.AbstractIoFuture#notifyCancelled`, `org.xnio.AbstractIoFuture#notifyDone`, `org.xnio.AbstractIoFuture#notifyFailed` | diff |
| CVE-2023-6378 | ch.qos.logback:logback-classic | 1.2.3 | HIGH | FRAMEWORK | `ch.qos.logback.classic.spi.LoggingEventVO#readObject`, `ch.qos.logback.core.net.HardenedObjectInputStream#<init>`, `ch.qos.logback.core.net.HardenedObjectInputStream#initObjectFilter`, `ch.qos.logback.core.util.EnvUtil#getJDKVersion`, `ch.qos.logback.core.util.EnvUtil#isJDK9OrHigher`, `ch.qos.logback.core.util.EnvUtil#isJDK_N_OrHigher` | diff |
| CVE-2023-6378 | ch.qos.logback:logback-core | 1.2.3 | HIGH | FRAMEWORK | `ch.qos.logback.classic.spi.LoggingEventVO#readObject`, `ch.qos.logback.core.net.HardenedObjectInputStream#<init>`, `ch.qos.logback.core.net.HardenedObjectInputStream#initObjectFilter`, `ch.qos.logback.core.util.EnvUtil#getJDKVersion`, `ch.qos.logback.core.util.EnvUtil#isJDK9OrHigher`, `ch.qos.logback.core.util.EnvUtil#isJDK_N_OrHigher` | diff |
| CVE-2024-12798 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | FRAMEWORK | `ch.qos.logback.classic.boolex.JaninoEventEvaluator#evaluate`, `ch.qos.logback.classic.boolex.JaninoEventEvaluator#getDecoratedExpression`, `ch.qos.logback.classic.boolex.JaninoEventEvaluator#start`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#evaluate`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#setExpression`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#start` | llm (diff) |
| CVE-2024-12801 | ch.qos.logback:logback-core | 1.2.3 | LOW | NOT_REACHABLE | `ch.qos.logback.core.joran.event.SaxEventRecorder#buildSaxParser`, `ch.qos.logback.core.joran.event.SaxEventRecorder#error`, `ch.qos.logback.core.joran.event.SaxEventRecorder#fatalError`, `ch.qos.logback.core.joran.event.SaxEventRecorder#recordEvents`, `ch.qos.logback.core.joran.event.SaxEventRecorder#resolveEntity`, `ch.qos.logback.core.joran.event.SaxEventRecorder#warning` | diff |
| CVE-2024-1459 | io.undertow:undertow-core | 2.2.4.Final | MODERATE | FRAMEWORK | `io.undertow.server.protocol.http.HttpRequestParser#handlePath`, `io.undertow.server.protocol.http.ParseState#reset` | diff |
| CVE-2024-1635 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.conduits.WriteTimeoutStreamSinkConduit#<init>`, `io.undertow.conduits.WriteTimeoutStreamSinkConduit#handleResumeTimeout`, `io.undertow.conduits.WriteTimeoutStreamSinkConduit#suspendWrites`, `io.undertow.conduits.WriteTimeoutStreamSinkConduit#terminateWrites`, `io.undertow.conduits.WriteTimeoutStreamSinkConduit#truncateWrites`, `io.undertow.conduits.WriteTimeoutStreamSinkConduit#writeFinal`, `io.undertow.server.protocol.ajp.AjpReadListener#handleEvent`, `io.undertow.server.protocol.http.HttpRequestParser#handlePath`, `io.undertow.server.protocol.http.ParseState#reset` | diff |
| CVE-2024-22243 | org.springframework:spring-web | 5.3.4 | HIGH | UNKNOWN |  |  |
| CVE-2024-22257 | org.springframework.security:spring-security-core | 5.4.5 | HIGH | FRAMEWORK | `org.springframework.security.access.vote.AuthenticatedVoter#isFullyAuthenticated` | diff |
| CVE-2024-22259 | org.springframework:spring-web | 5.3.4 | HIGH | FRAMEWORK | `org.springframework.web.util.UriComponentsBuilder#checkSchemeAndHost`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UriComponentsBuilder#fromOriginHeader`, `org.springframework.web.util.UriComponentsBuilder#fromUriString` | diff |
| CVE-2024-22262 | org.springframework:spring-web | 5.3.4 | HIGH | UNKNOWN | `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UriComponentsBuilder#fromUriString` | llm (advisory) |
| CVE-2024-3653 | io.undertow:undertow-core | 2.2.4.Final | MODERATE | UNKNOWN |  |  |
| CVE-2024-38808 | org.springframework:spring-expression | 5.3.4 | MODERATE | FRAMEWORK | `org.springframework.expression.EvaluationContext#isAssignmentEnabled`, `org.springframework.expression.spel.ast.Assign#getValueInternal`, `org.springframework.expression.spel.ast.ConstructorReference#createArray`, `org.springframework.expression.spel.ast.OpDec#getValueInternal`, `org.springframework.expression.spel.ast.OpInc#getValueInternal`, `org.springframework.expression.spel.support.SimpleEvaluationContext#<init>`, `org.springframework.expression.spel.support.SimpleEvaluationContext#build`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forPropertyAccessors`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forReadOnlyDataBinding`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forReadWriteDataBinding`, `org.springframework.expression.spel.support.SimpleEvaluationContext#isAssignmentEnabled` | diff |
| CVE-2024-38809 | org.springframework:spring-web | 5.3.4 | MODERATE | NOT_REACHABLE | `org.springframework.http.ETag#<init>`, `org.springframework.http.ETag#formattedTag`, `org.springframework.http.ETag#isWildcard`, `org.springframework.http.ETag#parse`, `org.springframework.http.ETag#tag`, `org.springframework.http.ETag#toString`, `org.springframework.http.ETag#weak`, `org.springframework.http.HttpHeaders#getETagValuesAsList`, `org.springframework.web.context.request.ServletWebRequest#matchRequestedETags`, `org.springframework.web.context.request.ServletWebRequest#validateIfNoneMatch` | diff |
| CVE-2024-38816 | org.springframework:spring-webmvc | 5.3.4 | HIGH | NOT_REACHABLE | `org.springframework.web.reactive.function.server.PathResourceLookupFunction#apply`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#cleanDuplicateSlashes`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#cleanLeadingSlash`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidEncodedInputPath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidEncodedResourcePath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isResourceUnderLocation`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#processPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#apply`, `org.springframework.web.servlet.function.PathResourceLookupFunction#cleanDuplicateSlashes`, `org.springframework.web.servlet.function.PathResourceLookupFunction#cleanLeadingSlash`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidEncodedInputPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidEncodedResourcePath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isResourceUnderLocation`, `org.springframework.web.servlet.function.PathResourceLookupFunction#processPath` | diff |
| CVE-2024-38819 | org.springframework:spring-webmvc | 5.3.4 | HIGH | NOT_REACHABLE | `org.springframework.web.reactive.function.server.PathResourceLookupFunction#decode`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidPath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#normalizePath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#processPath`, `org.springframework.web.reactive.resource.ResourceWebHandler#decode`, `org.springframework.web.reactive.resource.ResourceWebHandler#isInvalidPath`, `org.springframework.web.reactive.resource.ResourceWebHandler#normalizePath`, `org.springframework.web.reactive.resource.ResourceWebHandler#processPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#decode`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#normalizePath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#processPath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#decode`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#isInvalidPath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#normalizePath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#processPath` | diff |
| CVE-2024-38820 | org.springframework:spring-context | 5.3.4 | MODERATE | REACHABLE | `org.springframework.docs.integration.observability.config.conventions.CustomServerRequestObservationConvention#getContextualName`, `org.springframework.http.HttpHeaders#<init>`, `org.springframework.http.HttpHeaders#setAcceptCharset`, `org.springframework.http.MediaTypeFactory#getMediaTypes`, `org.springframework.http.MediaTypeFactory#parseMimeTypes`, `org.springframework.http.client.JdkClientHttpRequest#buildRequest`, `org.springframework.http.client.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.client.observation.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.http.client.reactive.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.server.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.http.server.reactive.ServletServerHttpRequest#createDefaultHttpHeaders`, `org.springframework.http.server.reactive.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.mock.web.MockHttpServletRequest#setContentType`, `org.springframework.mock.web.MockHttpServletRequest#updateContentTypeHeader`, `org.springframework.mock.web.MockHttpServletResponse#setContentType`, `org.springframework.mock.web.MockHttpServletResponse#setExplicitCharacterEncoding`, `org.springframework.mock.web.MockHttpServletResponse#updateContentTypePropertyAndHeader`, `org.springframework.util.MimeType#<init>`, `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields`, `org.springframework.web.accept.ContentNegotiationManagerFactoryBean#addMediaType`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#<init>`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#lookupMediaType`, `org.springframework.web.accept.PathExtensionContentNegotiationStrategy#getMediaTypeKey`, `org.springframework.web.client.RestClientResponseException#copyHeaders`, `org.springframework.web.filter.ForwardedHeaderFilter#headerNames`, `org.springframework.web.filter.HiddenHttpMethodFilter#doFilterInternal`, `org.springframework.web.filter.reactive.HiddenHttpMethodFilter#mapExchange`, `org.springframework.web.multipart.support.MultipartResolutionDelegate#isMultipartContent`, `org.springframework.web.multipart.support.StandardMultipartHttpServletRequest#handleParseFailure`, `org.springframework.web.reactive.accept.ParameterContentTypeResolver#formatKey`, `org.springframework.web.reactive.function.client.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.web.reactive.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.reactive.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.reactive.resource.ResourceWebHandler#getMediaType`, `org.springframework.web.reactive.resource.ResourceWebHandler#setMediaTypes`, `org.springframework.web.reactive.result.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.method.annotation.AbstractMessageConverterMethodProcessor#safeExtension`, `org.springframework.web.servlet.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.servlet.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#getMediaType`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#setMediaTypes`, `org.springframework.web.servlet.view.xslt.XsltView#configureResponse`, `org.springframework.web.socket.WebSocketExtension#<init>`, `org.springframework.web.socket.adapter.standard.StandardToWebSocketExtensionAdapter#initParameters`, `org.springframework.web.socket.client.standard.StandardWebSocketClient#getPort`, `org.springframework.web.util.DisconnectedClientHelper#isClientDisconnectedException`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UrlPathHelper#removeJsessionid` | diff |
| CVE-2024-38820 | org.springframework:spring-web | 5.3.4 | MODERATE | REACHABLE | `org.springframework.docs.integration.observability.config.conventions.CustomServerRequestObservationConvention#getContextualName`, `org.springframework.http.HttpHeaders#<init>`, `org.springframework.http.HttpHeaders#setAcceptCharset`, `org.springframework.http.MediaTypeFactory#getMediaTypes`, `org.springframework.http.MediaTypeFactory#parseMimeTypes`, `org.springframework.http.client.JdkClientHttpRequest#buildRequest`, `org.springframework.http.client.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.client.observation.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.http.client.reactive.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.server.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.http.server.reactive.ServletServerHttpRequest#createDefaultHttpHeaders`, `org.springframework.http.server.reactive.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.mock.web.MockHttpServletRequest#setContentType`, `org.springframework.mock.web.MockHttpServletRequest#updateContentTypeHeader`, `org.springframework.mock.web.MockHttpServletResponse#setContentType`, `org.springframework.mock.web.MockHttpServletResponse#setExplicitCharacterEncoding`, `org.springframework.mock.web.MockHttpServletResponse#updateContentTypePropertyAndHeader`, `org.springframework.util.MimeType#<init>`, `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields`, `org.springframework.web.accept.ContentNegotiationManagerFactoryBean#addMediaType`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#<init>`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#lookupMediaType`, `org.springframework.web.accept.PathExtensionContentNegotiationStrategy#getMediaTypeKey`, `org.springframework.web.client.RestClientResponseException#copyHeaders`, `org.springframework.web.filter.ForwardedHeaderFilter#headerNames`, `org.springframework.web.filter.HiddenHttpMethodFilter#doFilterInternal`, `org.springframework.web.filter.reactive.HiddenHttpMethodFilter#mapExchange`, `org.springframework.web.multipart.support.MultipartResolutionDelegate#isMultipartContent`, `org.springframework.web.multipart.support.StandardMultipartHttpServletRequest#handleParseFailure`, `org.springframework.web.reactive.accept.ParameterContentTypeResolver#formatKey`, `org.springframework.web.reactive.function.client.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.web.reactive.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.reactive.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.reactive.resource.ResourceWebHandler#getMediaType`, `org.springframework.web.reactive.resource.ResourceWebHandler#setMediaTypes`, `org.springframework.web.reactive.result.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.method.annotation.AbstractMessageConverterMethodProcessor#safeExtension`, `org.springframework.web.servlet.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.servlet.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#getMediaType`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#setMediaTypes`, `org.springframework.web.servlet.view.xslt.XsltView#configureResponse`, `org.springframework.web.socket.WebSocketExtension#<init>`, `org.springframework.web.socket.adapter.standard.StandardToWebSocketExtensionAdapter#initParameters`, `org.springframework.web.socket.client.standard.StandardWebSocketClient#getPort`, `org.springframework.web.util.DisconnectedClientHelper#isClientDisconnectedException`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UrlPathHelper#removeJsessionid` | diff |
| CVE-2024-38821 | org.springframework.security:spring-security-web | 5.4.5 | CRITICAL | REACHABLE | `org.springframework.security.authentication.MapUserDetailsRepository#<init>`, `org.springframework.security.authentication.MapUserDetailsRepository#findByUsername`, `org.springframework.security.authentication.MapUserDetailsRepository#getKey`, `org.springframework.security.authentication.ReactiveAuthenticationManagerAdapter#<init>`, `org.springframework.security.authentication.ReactiveAuthenticationManagerAdapter#authenticate`, `org.springframework.security.authentication.UserDetailsRepositoryAuthenticationManager#<init>`, `org.springframework.security.authentication.UserDetailsRepositoryAuthenticationManager#authenticate`, `org.springframework.security.authorization.AuthenticatedAuthorizationManager#authenticated`, `org.springframework.security.authorization.AuthenticatedAuthorizationManager#check`, `org.springframework.security.authorization.AuthorityAuthorizationManager#<init>`, `org.springframework.security.authorization.AuthorityAuthorizationManager#check`, `org.springframework.security.authorization.AuthorityAuthorizationManager#hasAuthority`, `org.springframework.security.authorization.AuthorityAuthorizationManager#hasRole`, `org.springframework.security.authorization.AuthorizationDecision#<init>`, `org.springframework.security.authorization.AuthorizationDecision#isGranted`, `org.springframework.security.authorization.ReactiveAuthorizationManager#verify`, `org.springframework.security.config.web.server.AbstractServerWebExchangeMatcherRegistry#antMatchers`, `org.springframework.security.config.web.server.AbstractServerWebExchangeMatcherRegistry#anyExchange`, `org.springframework.security.config.web.server.AbstractServerWebExchangeMatcherRegistry#matcher`, `org.springframework.security.config.web.server.AbstractServerWebExchangeMatcherRegistry#matchers`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#access`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#anyExchange`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#authenticated`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#build`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#denyAll`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#hasAuthority`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#hasRole`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#permitAll`, `org.springframework.security.config.web.server.AuthorizeExchangeBuilder#registerMatcher`, `org.springframework.security.config.web.server.HeaderBuilder#<init>`, `org.springframework.security.config.web.server.HeaderBuilder#build`, `org.springframework.security.config.web.server.HeaderBuilder#cache`, `org.springframework.security.config.web.server.HeaderBuilder#contentTypeOptions`, `org.springframework.security.config.web.server.HeaderBuilder#disable`, `org.springframework.security.config.web.server.HeaderBuilder#frameOptions`, `org.springframework.security.config.web.server.HeaderBuilder#hsts`, `org.springframework.security.config.web.server.HeaderBuilder#includeSubdomains`, `org.springframework.security.config.web.server.HeaderBuilder#maxAge`, `org.springframework.security.config.web.server.HeaderBuilder#mode`, `org.springframework.security.config.web.server.HeaderBuilder#xssProtection`, `org.springframework.security.config.web.server.HttpBasicBuilder#authenticationManager`, `org.springframework.security.config.web.server.HttpBasicBuilder#build`, `org.springframework.security.config.web.server.HttpBasicBuilder#securityContextRepository`, `org.springframework.security.config.web.server.HttpSecurity#authenticationManager`, `org.springframework.security.config.web.server.HttpSecurity#authorizeExchange`, `org.springframework.security.config.web.server.HttpSecurity#build`, `org.springframework.security.config.web.server.HttpSecurity#headers`, `org.springframework.security.config.web.server.HttpSecurity#http`, `org.springframework.security.config.web.server.HttpSecurity#httpBasic`, `org.springframework.security.config.web.server.HttpSecurity#securityContextRepository`, `org.springframework.security.config.web.server.HttpSecurity#securityContextRepositoryWebFilter`, `org.springframework.security.web.reactive.result.method.annotation.AuthenticationPrincipalArgumentResolver#<init>`, `org.springframework.security.web.reactive.result.method.annotation.AuthenticationPrincipalArgumentResolver#findMethodAnnotation`, `org.springframework.security.web.reactive.result.method.annotation.AuthenticationPrincipalArgumentResolver#resolveArgument`, `org.springframework.security.web.reactive.result.method.annotation.AuthenticationPrincipalArgumentResolver#resolvePrincipal`, `org.springframework.security.web.reactive.result.method.annotation.AuthenticationPrincipalArgumentResolver#supportsParameter`, `org.springframework.security.web.server.HttpBasicAuthenticationConverter#apply`, `org.springframework.security.web.server.HttpBasicAuthenticationConverter#base64Decode`, `org.springframework.security.web.server.WebFilterChainFilter#<init>`, `org.springframework.security.web.server.WebFilterChainFilter#filter`, `org.springframework.security.web.server.WebFilterChainProxy#filter`, `org.springframework.security.web.server.WebFilterChainProxy#filterFirewalledExchange`, `org.springframework.security.web.server.WebFilterChainProxy#setExchangeRejectedHandler`, `org.springframework.security.web.server.WebFilterChainProxy#setFirewall`, `org.springframework.security.web.server.authentication.AuthenticationWebFilter#<init>`, `org.springframework.security.web.server.authentication.AuthenticationWebFilter#filter`, `org.springframework.security.web.server.authentication.AuthenticationWebFilter#setAuthenticationConverter`, `org.springframework.security.web.server.authentication.AuthenticationWebFilter#setAuthenticationSuccessHandler`, `org.springframework.security.web.server.authentication.AuthenticationWebFilter#setEntryPoint`, `org.springframework.security.web.server.authentication.DefaultAuthenticationSuccessHandler#setDelegate`, `org.springframework.security.web.server.authentication.DefaultAuthenticationSuccessHandler#setSecurityContextRepository`, `org.springframework.security.web.server.authentication.DefaultAuthenticationSuccessHandler#success`, `org.springframework.security.web.server.authentication.WebFilterChainAuthenticationSuccessHandler#success`, `org.springframework.security.web.server.authentication.www.HttpBasicAuthenticationEntryPoint#commence`, `org.springframework.security.web.server.authorization.AuthorizationContext#<init>`, `org.springframework.security.web.server.authorization.AuthorizationContext#getExchange`, `org.springframework.security.web.server.authorization.AuthorizationContext#getVariables`, `org.springframework.security.web.server.authorization.AuthorizationWebFilter#<init>`, `org.springframework.security.web.server.authorization.AuthorizationWebFilter#filter`, `org.springframework.security.web.server.authorization.DelegatingReactiveAuthorizationManager#<init>`, `org.springframework.security.web.server.authorization.DelegatingReactiveAuthorizationManager#add`, `org.springframework.security.web.server.authorization.DelegatingReactiveAuthorizationManager#build`, `org.springframework.security.web.server.authorization.DelegatingReactiveAuthorizationManager#builder`, `org.springframework.security.web.server.authorization.DelegatingReactiveAuthorizationManager#check`, `org.springframework.security.web.server.authorization.ExceptionTranslationWebFilter#filter`, `org.springframework.security.web.server.authorization.HttpStatusAccessDeniedHandler#<init>`, `org.springframework.security.web.server.authorization.HttpStatusAccessDeniedHandler#handle`, `org.springframework.security.web.server.context.SecurityContextRepositoryServerWebExchange#<init>`, `org.springframework.security.web.server.context.SecurityContextRepositoryServerWebExchange#getPrincipal`, `org.springframework.security.web.server.context.SecurityContextRepositoryWebFilter#<init>`, `org.springframework.security.web.server.context.SecurityContextRepositoryWebFilter#filter`, `org.springframework.security.web.server.context.ServerWebExchangeAttributeSecurityContextRepository#load`, `org.springframework.security.web.server.context.ServerWebExchangeAttributeSecurityContextRepository#save`, `org.springframework.security.web.server.context.WebSessionSecurityContextRepository#load`, `org.springframework.security.web.server.context.WebSessionSecurityContextRepository#save`, `org.springframework.security.web.server.firewall.HttpStatusExchangeRejectedHandler#<init>`, `org.springframework.security.web.server.firewall.HttpStatusExchangeRejectedHandler#handle`, `org.springframework.security.web.server.firewall.ServerExchangeRejectedException#<init>`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#<init>`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#containsOnlyPrintableAsciiCharacters`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#createDefaultAllowedHttpMethods`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#decodedUrlContains`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#encodedUrlContains`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#get`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getDecodedUrlBlocklist`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getEncodedUrlBlocklist`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getFirewalledExchange`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getFirst`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getHeaders`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getQueryParams`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#getRequest`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#hasCrlf`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#isNormalized`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#keySet`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#rejectForbiddenHttpMethod`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#rejectNonPrintableAsciiCharactersInFieldName`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#rejectedBlocklistedUrls`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#rejectedUntrustedHosts`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowBackSlash`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowNull`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowSemicolon`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedCarriageReturn`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedDoubleSlash`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedLineFeed`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedLineSeparator`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedParagraphSeparator`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedPercent`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedPeriod`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowUrlEncodedSlash`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedHeaderNames`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedHeaderValues`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedHostnames`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedHttpMethods`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedParameterNames`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setAllowedParameterValues`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#setUnsafeAllowAnyHttpMethod`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#urlBlocklistsAddAll`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#urlBlocklistsRemoveAll`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#validateAllowedHeaderName`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#validateAllowedHeaderValue`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#validateAllowedParameterName`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#validateAllowedParameterValue`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#validateCrlf`, `org.springframework.security.web.server.firewall.StrictServerWebExchangeFirewall#valueContains`, `org.springframework.security.web.server.header.CacheControlHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.CompositeHttpHeadersWriter#<init>`, `org.springframework.security.web.server.header.CompositeHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.ContentTypeOptionsHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.HttpHeaderWriterWebFilter#<init>`, `org.springframework.security.web.server.header.HttpHeaderWriterWebFilter#filter`, `org.springframework.security.web.server.header.StaticHttpHeadersWriter#<init>`, `org.springframework.security.web.server.header.StaticHttpHeadersWriter#build`, `org.springframework.security.web.server.header.StaticHttpHeadersWriter#builder`, `org.springframework.security.web.server.header.StaticHttpHeadersWriter#header`, `org.springframework.security.web.server.header.StaticHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#<init>`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#isSecure`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#setIncludeSubDomains`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#setMaxAge`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#updateDelegate`, `org.springframework.security.web.server.header.StrictTransportSecurityHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.XContentTypeOptionsHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.XFrameOptionsHttpHeadersWriter#createDelegate`, `org.springframework.security.web.server.header.XFrameOptionsHttpHeadersWriter#setMode`, `org.springframework.security.web.server.header.XFrameOptionsHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#<init>`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#createHeaderValue`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#setBlock`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#setEnabled`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#updateDelegate`, `org.springframework.security.web.server.header.XXssProtectionHttpHeadersWriter#writeHttpHeaders`, `org.springframework.security.web.server.util.matcher.AndServerWebExchangeMatcher#<init>`, `org.springframework.security.web.server.util.matcher.AndServerWebExchangeMatcher#matches`, `org.springframework.security.web.server.util.matcher.AndServerWebExchangeMatcher#toString`, `org.springframework.security.web.server.util.matcher.OrServerWebExchangeMatcher#<init>`, `org.springframework.security.web.server.util.matcher.OrServerWebExchangeMatcher#matches`, `org.springframework.security.web.server.util.matcher.OrServerWebExchangeMatcher#toString`, `org.springframework.security.web.server.util.matcher.PathMatcherServerWebExchangeMatcher#<init>`, `org.springframework.security.web.server.util.matcher.PathMatcherServerWebExchangeMatcher#matches`, `org.springframework.security.web.server.util.matcher.PathMatcherServerWebExchangeMatcher#setPathMatcher`, `org.springframework.security.web.server.util.matcher.PathMatcherServerWebExchangeMatcher#toString`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatcher#<init>`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatcher#getVariables`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatcher#isMatch`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatcher#match`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatcher#notMatch`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers#<init>`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers#antMatchers`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers#anyExchange`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers#matchers`, `org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers#matches`, `sample.Application#authenticationManager`, `sample.Application#authenticationPrincipalArgumentResolver`, `sample.Application#configureArgumentResolvers`, `sample.Application#currentUserMatchesPath`, `sample.Application#main`, `sample.Application#nettyContext`, `sample.Application#springSecurityFilterChain`, `sample.MapUserRepository#<init>`, `sample.MapUserRepository#findAll`, `sample.MapUserRepository#findByUsername`, `sample.MapUserRepository#save`, `sample.User#<init>`, `sample.User#getFirstname`, `sample.User#getId`, `sample.User#getLastname`, `sample.User#getPassword`, `sample.User#getUsername`, `sample.User#setFirstname`, `sample.User#setId`, `sample.User#setLastname`, `sample.User#setPassword`, `sample.User#setUsername`, `sample.UserController#<init>`, `sample.UserController#Session`, `sample.UserController#admin`, `sample.UserController#me`, `sample.UserController#principal`, `sample.UserController#users`, `sample.UserRepositoryUserDetailsRepository#<init>`, `sample.UserRepositoryUserDetailsRepository#findByUsername`, `sample.UserRepositoryUserDetailsRepository#getAuthorities`, `sample.UserRepositoryUserDetailsRepository#isAccountNonExpired`, `sample.UserRepositoryUserDetailsRepository#isAccountNonLocked`, `sample.UserRepositoryUserDetailsRepository#isAdmin`, `sample.UserRepositoryUserDetailsRepository#isCredentialsNonExpired`, `sample.UserRepositoryUserDetailsRepository#isEnabled` | diff |
| CVE-2024-38827 | org.springframework.security:spring-security-core | 5.4.5 | MODERATE | FRAMEWORK | `org.springframework.aot.agent.InvocationsRecorderClassVisitor#rewriteMethodName`, `org.springframework.beans.support.PropertyComparator#compare`, `org.springframework.core.convert.support.StringToBooleanConverter#convert`, `org.springframework.core.env.SystemEnvironmentPropertySource#resolvePropertyName`, `org.springframework.expression.spel.ast.TypeReference#getValueInternal`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#eatPossiblyQualifiedId`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#eatToken`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#toString`, `org.springframework.expression.spel.standard.Tokenizer#lexIdentifier`, `org.springframework.format.datetime.standard.MonthFormatter#parse`, `org.springframework.jdbc.core.metadata.CallMetaDataContext#lowerCase`, `org.springframework.jdbc.core.metadata.CallMetaDataContext#matchInParameterValuesWithCallParameters`, `org.springframework.jdbc.core.metadata.CallMetaDataContext#reconcileParameters`, `org.springframework.jdbc.core.metadata.Db2CallMetaDataProvider#metaDataSchemaNameToUse`, `org.springframework.jdbc.core.metadata.DerbyCallMetaDataProvider#metaDataSchemaNameToUse`, `org.springframework.jdbc.core.metadata.GenericCallMetaDataProvider#identifierNameToUse`, `org.springframework.jdbc.core.metadata.GenericTableMetaDataProvider#findTableMetaData`, `org.springframework.jdbc.core.metadata.GenericTableMetaDataProvider#identifierNameToUse`, `org.springframework.jdbc.core.metadata.GenericTableMetaDataProvider#locateTableAndProcessMetaData`, `org.springframework.jdbc.core.metadata.TableMetaDataContext#createInsertString`, `org.springframework.jdbc.core.metadata.TableMetaDataContext#createInsertTypes`, `org.springframework.jdbc.core.metadata.TableMetaDataContext#matchInParameterValuesWithInsertColumns`, `org.springframework.jdbc.core.metadata.TableMetaDataContext#reconcileColumnsToUse`, `org.springframework.jdbc.core.namedparam.SqlParameterSourceUtils#extractCaseInsensitiveParameterNames`, `org.springframework.jdbc.core.simple.AbstractJdbcInsert#executeInsertAndReturnKeyHolderInternal`, `org.springframework.jms.config.JmsListenerContainerParser#parseSpecificContainerProperties`, `org.springframework.scheduling.quartz.LocalDataSourceJobStore#initialize`, `org.springframework.scheduling.support.CronField#replaceOrdinals`, `org.springframework.util.ResourceUtils#isJarFileURL` | diff |
| CVE-2024-38828 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2024-3884 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.Handlers#hostHeaderHandler`, `io.undertow.attribute.QueryStringAttribute#readAttribute`, `io.undertow.attribute.RequestLineAttribute#readAttribute`, `io.undertow.conduits.FixedLengthStreamSourceConduit#exitRead`, `io.undertow.security.handlers.SinglePortConfidentialityHandler#getRedirectURI`, `io.undertow.security.impl.DigestAuthenticationMechanism#handleDigestHeader`, `io.undertow.server.Connectors#setExchangeRequestPath`, `io.undertow.server.HttpServerExchange#cleanQueryString`, `io.undertow.server.HttpServerExchange#getDecodedQueryString`, `io.undertow.server.HttpServerExchange#getNonDecodedQueryString`, `io.undertow.server.HttpServerExchange#getQueryString`, `io.undertow.server.HttpServerExchange#setDecodedQueryString`, `io.undertow.server.HttpServerExchange#setNonDecodedQueryString`, `io.undertow.server.HttpServerExchange#setQueryString`, `io.undertow.server.handlers.HostHeaderHandler#<init>`, `io.undertow.server.handlers.HostHeaderHandler#handleRequest`, `io.undertow.server.handlers.HostHeaderHandler#terminate`, `io.undertow.server.handlers.HostHeaderHandler#wrap`, `io.undertow.server.handlers.HttpTraceHandler#handleRequest`, `io.undertow.server.handlers.JDBCLogHandler#logMessage`, `io.undertow.server.handlers.LearningPushHandler#handleRequest`, `io.undertow.server.handlers.RequestDumpingHandler#handleRequest`, `io.undertow.server.handlers.StuckThreadDetectionHandler#handleRequest`, `io.undertow.server.handlers.URLDecodingHandler#decodeQueryString`, `io.undertow.server.handlers.accesslog.ExtendedAccessLogParser#readAttribute`, `io.undertow.server.handlers.form.MultiPartParserDefinition#create`, `io.undertow.server.handlers.proxy.ProxyHandler#run`, `io.undertow.server.handlers.resource.DirectoryUtils#sendRequestedBlobs`, `io.undertow.server.protocol.ajp.AjpRequestParser#parse`, `io.undertow.server.protocol.http.HttpReadListener#<init>`, `io.undertow.server.protocol.http.HttpReadListener#handleEventWithNoRunningRequest`, `io.undertow.server.protocol.http.HttpRequestParser#handleQueryParameters`, `io.undertow.server.protocol.http2.Http2ReceiveListener#handleInitialRequest`, `io.undertow.servlet.attribute.ServletRelativePathAttribute#readAttribute`, `io.undertow.servlet.attribute.ServletRequestLineAttribute#readAttribute`, `io.undertow.servlet.spec.HttpServletRequestImpl#getQueryString`, `io.undertow.servlet.util.DispatchUtils#assignRequestPath`, `io.undertow.util.WeakCopyOnWriteMap#<init>`, `io.undertow.util.WeakCopyOnWriteMap#clear`, `io.undertow.util.WeakCopyOnWriteMap#containsKey`, `io.undertow.util.WeakCopyOnWriteMap#containsValue`, `io.undertow.util.WeakCopyOnWriteMap#entrySet`, `io.undertow.util.WeakCopyOnWriteMap#get`, `io.undertow.util.WeakCopyOnWriteMap#isEmpty`, `io.undertow.util.WeakCopyOnWriteMap#keySet`, `io.undertow.util.WeakCopyOnWriteMap#put`, `io.undertow.util.WeakCopyOnWriteMap#putAll`, `io.undertow.util.WeakCopyOnWriteMap#putIfAbsent`, `io.undertow.util.WeakCopyOnWriteMap#putInternal`, `io.undertow.util.WeakCopyOnWriteMap#remove`, `io.undertow.util.WeakCopyOnWriteMap#removeInternal`, `io.undertow.util.WeakCopyOnWriteMap#replace`, `io.undertow.util.WeakCopyOnWriteMap#size`, `io.undertow.util.WeakCopyOnWriteMap#values`, `io.undertow.websockets.core.WebSocketChannel#handleBrokenSourceChannel`, `io.undertow.websockets.core.WebSocketChannel#parseFrame`, `io.undertow.websockets.jsr.FrameHandler#run`, `io.undertow.websockets.jsr.UndertowContainerProvider#get`, `io.undertow.websockets.jsr.UndertowContainerProvider#getDefaultContainer` | diff |
| CVE-2024-4027 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.Handlers#hostHeaderHandler`, `io.undertow.attribute.QueryStringAttribute#readAttribute`, `io.undertow.attribute.RequestLineAttribute#readAttribute`, `io.undertow.conduits.FixedLengthStreamSourceConduit#exitRead`, `io.undertow.security.handlers.SinglePortConfidentialityHandler#getRedirectURI`, `io.undertow.security.impl.DigestAuthenticationMechanism#handleDigestHeader`, `io.undertow.server.Connectors#setExchangeRequestPath`, `io.undertow.server.HttpServerExchange#cleanQueryString`, `io.undertow.server.HttpServerExchange#getDecodedQueryString`, `io.undertow.server.HttpServerExchange#getNonDecodedQueryString`, `io.undertow.server.HttpServerExchange#getQueryString`, `io.undertow.server.HttpServerExchange#setDecodedQueryString`, `io.undertow.server.HttpServerExchange#setNonDecodedQueryString`, `io.undertow.server.HttpServerExchange#setQueryString`, `io.undertow.server.RequestStatistics#getQueryString`, `io.undertow.server.handlers.HostHeaderHandler#<init>`, `io.undertow.server.handlers.HostHeaderHandler#handleRequest`, `io.undertow.server.handlers.HostHeaderHandler#terminate`, `io.undertow.server.handlers.HostHeaderHandler#wrap`, `io.undertow.server.handlers.HttpTraceHandler#handleRequest`, `io.undertow.server.handlers.JDBCLogHandler#logMessage`, `io.undertow.server.handlers.LearningPushHandler#handleRequest`, `io.undertow.server.handlers.RequestDumpingHandler#handleRequest`, `io.undertow.server.handlers.StuckThreadDetectionHandler#handleRequest`, `io.undertow.server.handlers.URLDecodingHandler#decodeQueryString`, `io.undertow.server.handlers.accesslog.ExtendedAccessLogParser#readAttribute`, `io.undertow.server.handlers.form.MultiPartParserDefinition#create`, `io.undertow.server.handlers.proxy.ProxyHandler#run`, `io.undertow.server.handlers.resource.DirectoryUtils#sendRequestedBlobs`, `io.undertow.server.protocol.ajp.AjpRequestParser#parse`, `io.undertow.server.protocol.http.HttpReadListener#<init>`, `io.undertow.server.protocol.http.HttpReadListener#handleEventWithNoRunningRequest`, `io.undertow.server.protocol.http.HttpRequestParser#handleQueryParameters`, `io.undertow.server.protocol.http2.Http2ReceiveListener#handleInitialRequest`, `io.undertow.servlet.attribute.ServletRelativePathAttribute#readAttribute`, `io.undertow.servlet.attribute.ServletRequestLineAttribute#readAttribute`, `io.undertow.servlet.spec.HttpServletRequestImpl#getQueryString`, `io.undertow.servlet.util.DispatchUtils#assignRequestPath`, `io.undertow.util.WeakCopyOnWriteMap#<init>`, `io.undertow.util.WeakCopyOnWriteMap#clear`, `io.undertow.util.WeakCopyOnWriteMap#containsKey`, `io.undertow.util.WeakCopyOnWriteMap#containsValue`, `io.undertow.util.WeakCopyOnWriteMap#entrySet`, `io.undertow.util.WeakCopyOnWriteMap#get`, `io.undertow.util.WeakCopyOnWriteMap#isEmpty`, `io.undertow.util.WeakCopyOnWriteMap#keySet`, `io.undertow.util.WeakCopyOnWriteMap#put`, `io.undertow.util.WeakCopyOnWriteMap#putAll`, `io.undertow.util.WeakCopyOnWriteMap#putIfAbsent`, `io.undertow.util.WeakCopyOnWriteMap#putInternal`, `io.undertow.util.WeakCopyOnWriteMap#remove`, `io.undertow.util.WeakCopyOnWriteMap#removeInternal`, `io.undertow.util.WeakCopyOnWriteMap#replace`, `io.undertow.util.WeakCopyOnWriteMap#size`, `io.undertow.util.WeakCopyOnWriteMap#values`, `io.undertow.websockets.core.WebSocketChannel#handleBrokenSourceChannel`, `io.undertow.websockets.core.WebSocketChannel#parseFrame`, `io.undertow.websockets.jsr.FrameHandler#run`, `io.undertow.websockets.jsr.UndertowContainerProvider#get`, `io.undertow.websockets.jsr.UndertowContainerProvider#getDefaultContainer` | diff |
| CVE-2024-47072 | com.thoughtworks.xstream:xstream | 1.4.5 | HIGH | REACHABLE | `com.thoughtworks.xstream.XStream#fromXML`, `com.thoughtworks.xstream.XStream#unmarshal` | manual |
| CVE-2024-5971 | io.undertow:undertow-core | 2.2.4.Final | HIGH | UNKNOWN |  |  |
| CVE-2024-6162 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.server.protocol.ajp.AjpRequestParser#decode` | diff |
| CVE-2024-7885 | io.undertow:undertow-core | 2.2.4.Final | HIGH | FRAMEWORK | `io.undertow.server.protocol.proxy.ProxyProtocolReadListener#parseProxyProtocolV1` | diff |
| CVE-2025-11226 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | NOT_REACHABLE | `ch.qos.logback.core.model.processor.conditional.IfModelHandler#handle`, `ch.qos.logback.core.model.processor.conditional.IfModelHandler#hasNew` | diff |
| CVE-2025-12543 | io.undertow:undertow-core | 2.2.4.Final | CRITICAL | UNKNOWN |  |  |
| CVE-2025-22233 | org.springframework:spring-context | 5.3.4 | LOW | NOT_REACHABLE | `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields` | diff |
| CVE-2025-22235 | org.springframework.boot:spring-boot | 2.4.3 | HIGH | UNKNOWN |  |  |
| CVE-2025-35036 | org.hibernate.validator:hibernate-validator | 6.1.7.Final | MODERATE | NOT_REACHABLE | `org.hibernate.validator.constraintvalidation.HibernateConstraintViolationBuilder#enableExpressionLanguage`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#constraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#customViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#<init>`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#getExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#isCustomViolation`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#isExpressionLanguageEnabled`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#toString`, `org.hibernate.validator.internal.engine.PredefinedScopeValidatorFactoryImpl#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineBeanMetaDataClassNormalizer`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineGetterPropertySelectionStrategy`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineLocaleResolver`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determinePropertyNodeNameProvider`, `org.hibernate.validator.internal.engine.ValidatorFactoryImpl#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#build`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#setConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#setCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addBeanNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addConstraintViolation`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addContainerElementNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addPropertyNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#buildConstraintViolationWithTemplate`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#enableExpressionLanguage`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#getDefaultConstraintViolationCreationContext`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#getExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#isCustomViolation`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#isExpressionLanguageEnabled`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#toString`, `org.hibernate.validator.internal.engine.constraintvalidation.CrossParameterConstraintValidatorContextImpl#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.CrossParameterConstraintValidatorContextImpl#buildConstraintViolationWithTemplate`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#addVariablesToElContext`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#bindContextValues`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#getElContext`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#interpolate`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanMethodsELContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesELResolver#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesELResolver#invoke`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#addELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#getELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.DisabledFeatureELException#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getCommonPropertyType`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getFeatureDescriptors`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getType`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getValue`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#invoke`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#isReadOnly`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#setValue`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#addELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#getELResolver`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#addConstraintFailure`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#createConstraintValidatorContextFor`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#interpolate`, `org.hibernate.validator.internal.engine.validationcontext.ParameterExecutableValidationContext#createConstraintValidatorContextFor`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#<init>`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.messageinterpolation.AbstractMessageInterpolator#interpolateMessage`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#<init>`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#interpretDefaultForConstraints`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#interpretDefaultForCustomViolations`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#of`, `org.hibernate.validator.messageinterpolation.ParameterMessageInterpolator#interpolate`, `org.hibernate.validator.messageinterpolation.ResourceBundleMessageInterpolator#interpolate` | diff |
| CVE-2025-41242 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2025-41249 | org.springframework:spring-core | 5.3.4 | HIGH | FRAMEWORK | `org.springframework.core.annotation.AnnotatedMethod#isOverrideFor`, `org.springframework.core.annotation.AnnotationsScanner#hasSameGenericTypeParameters` | diff |
| CVE-2025-48924 | org.apache.commons:commons-lang3 | 3.4 | MODERATE | NOT_REACHABLE | `org.apache.commons.lang3.ClassUtils#getClass`, `org.apache.commons.lang3.ClassUtils#toCanonicalName` | diff |
| CVE-2025-49128 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.core.JsonLocation#_appendSourceDesc`, `com.fasterxml.jackson.core.JsonLocation#_truncate`, `com.fasterxml.jackson.core.json.ReaderBasedJsonParser#<init>`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_reportInvalidToken` | diff |
| CVE-2025-52999 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | UNKNOWN |  |  |
| CVE-2025-9784 | io.undertow:undertow-core | 2.2.4.Final | HIGH | UNKNOWN |  |  |
| CVE-2026-10532 | ch.qos.logback:logback-core | 1.2.3 | LOW | UNKNOWN |  |  |
| CVE-2026-1225 | ch.qos.logback:logback-core | 1.2.3 | LOW | NOT_REACHABLE | `ch.qos.logback.classic.joran.ModelClassToModelHandlerLinker#link`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#<init>`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#addAppenderDeclaration`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#getAppenderNameSet`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#getSupportedModelClass`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#handle`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#isAppenderDeclared`, `ch.qos.logback.core.model.processor.AppenderRefModelHandler#attachReferencedAppenders`, `ch.qos.logback.core.model.processor.DefaultProcessor#allDependenciesStarted`, `ch.qos.logback.core.model.processor.DefaultProcessor#dependencyIsADirectSubmodel`, `ch.qos.logback.core.model.processor.DependencyDefinition#<init>`, `ch.qos.logback.core.model.processor.DependencyDefinition#getDependency`, `ch.qos.logback.core.model.processor.DependencyDefinition#toString`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#getDependencyNamesForModel`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#hasDependers`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#isNamedDependemcyStarted` | diff |
| CVE-2026-19032 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.databind.ext.NioPathDeserializer#<init>`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#_allowedSchemesDesc`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#_isSchemeAllowed`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#<init>`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_allowedSchemesDesc`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_isSchemeAllowed`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#findDeserializer` | diff |
| CVE-2026-22732 | org.springframework.security:spring-security-web | 5.4.5 | CRITICAL | UNKNOWN |  |  |
| CVE-2026-22733 | org.springframework.boot:spring-boot-starter-actuator | 2.4.3 | HIGH | UNKNOWN |  |  |
| CVE-2026-22735 | org.springframework:spring-webmvc | 5.3.4 | LOW | UNKNOWN |  |  |
| CVE-2026-22737 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2026-22741 | org.springframework:spring-webmvc | 5.3.4 | LOW | UNKNOWN |  |  |
| CVE-2026-22745 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2026-3260 | io.undertow:undertow-core | 2.2.4.Final | MODERATE | UNKNOWN |  |  |
| CVE-2026-40477 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40477 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40478 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40478 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40973 | org.springframework.boot:spring-boot | 2.4.3 | HIGH | UNKNOWN |  |  |
| CVE-2026-40984 | io.micrometer:micrometer-core | 1.6.4 | HIGH | NOT_REACHABLE | `io.micrometer.core.instrument.binder.http.HttpJakartaServletRequestTags#method`, `io.micrometer.core.instrument.binder.http.HttpMethods#<init>`, `io.micrometer.core.instrument.binder.http.HttpMethods#isStandard`, `io.micrometer.core.instrument.binder.http.HttpRequestTags#method`, `io.micrometer.core.instrument.binder.jersey.server.DefaultJerseyObservationConvention#getContextualName`, `io.micrometer.core.instrument.binder.jersey.server.HttpMethods#<init>`, `io.micrometer.core.instrument.binder.jersey.server.HttpMethods#isStandard`, `io.micrometer.core.instrument.binder.jersey.server.JerseyKeyValues#method`, `io.micrometer.core.instrument.binder.jersey.server.JerseyTags#method`, `io.micrometer.jetty12.server.DefaultJettyCoreRequestTagsProvider#method`, `io.micrometer.jetty12.server.HttpMethods#<init>`, `io.micrometer.jetty12.server.HttpMethods#isStandard` | diff |
| CVE-2026-41706 | org.springframework.security:spring-security-web | 5.4.5 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41711 | org.springframework.data:spring-data-commons | 2.4.5 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41716 | org.springframework.data:spring-data-commons | 2.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2026-41721 | org.springframework.data:spring-data-commons | 2.4.5 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41841 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41842 | org.springframework:spring-webmvc | 5.3.4 | HIGH | UNKNOWN |  |  |
| CVE-2026-41843 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41844 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | NOT_REACHABLE | `org.springframework.web.reactive.result.view.ViewResolutionResultHandler#getDefaultViewName`, `org.springframework.web.reactive.result.view.ViewResolutionResultHandler#handleResult`, `org.springframework.web.servlet.view.DefaultRequestToViewNameTranslator#getViewName`, `org.springframework.web.servlet.view.DefaultRequestToViewNameTranslator#transformPath` | diff |
| CVE-2026-41845 | org.springframework:spring-webmvc | 5.3.4 | HIGH | NOT_REACHABLE | `org.springframework.web.util.JavaScriptUtils#javaScriptEscape` | diff |
| CVE-2026-41846 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | NOT_REACHABLE | `org.springframework.web.servlet.tags.form.AbstractHtmlElementTag#resolveCssClass`, `org.springframework.web.servlet.tags.form.AbstractHtmlElementTag#writeOptionalAttributes`, `org.springframework.web.servlet.tags.form.FormTag#resolveCssClass` | diff |
| CVE-2026-41848 | org.springframework:spring-core | 5.3.4 | LOW | REACHABLE | `org.springframework.util.AntPathMatcher#<init>`, `org.springframework.util.AntPathMatcher#charAt`, `org.springframework.util.AntPathMatcher#isEmpty`, `org.springframework.util.AntPathMatcher#length`, `org.springframework.util.AntPathMatcher#matchStrings`, `org.springframework.util.AntPathMatcher#subSequence`, `org.springframework.util.AntPathMatcher#toString` | diff |
| CVE-2026-41849 | org.springframework:spring-expression | 5.3.4 | HIGH | UNKNOWN |  |  |
| CVE-2026-41850 | org.springframework:spring-expression | 5.3.4 | HIGH | UNKNOWN |  |  |
| CVE-2026-41851 | org.springframework:spring-expression | 5.3.4 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ExpressionState#assignVariable`, `org.springframework.expression.spel.ExpressionState#lookupVariable`, `org.springframework.expression.spel.ExpressionState#operate`, `org.springframework.expression.spel.ExpressionState#trackOperation`, `org.springframework.expression.spel.SpelParserConfiguration#<init>`, `org.springframework.expression.spel.SpelParserConfiguration#getMaximumOperations`, `org.springframework.expression.spel.SpelParserConfiguration#retrieveMaxOperations`, `org.springframework.expression.spel.ast.Assign#getValueInternal`, `org.springframework.expression.spel.ast.BeanReference#getValueInternal`, `org.springframework.expression.spel.ast.ConstructorReference#getValueInternal`, `org.springframework.expression.spel.ast.Elvis#getValueInternal`, `org.springframework.expression.spel.ast.FunctionReference#getValueInternal`, `org.springframework.expression.spel.ast.Indexer#<init>`, `org.springframework.expression.spel.ast.Indexer#getArrayElement`, `org.springframework.expression.spel.ast.Indexer#getValue`, `org.springframework.expression.spel.ast.Indexer#getValueRef`, `org.springframework.expression.spel.ast.Indexer#growCollectionIfNecessary`, `org.springframework.expression.spel.ast.Indexer#setArrayElement`, `org.springframework.expression.spel.ast.Indexer#setValue`, `org.springframework.expression.spel.ast.InlineList#<init>`, `org.springframework.expression.spel.ast.InlineList#createList`, `org.springframework.expression.spel.ast.InlineList#determineIfConstant`, `org.springframework.expression.spel.ast.InlineList#getConstantValue`, `org.springframework.expression.spel.ast.InlineList#getValueInternal`, `org.springframework.expression.spel.ast.InlineList#isCompilable`, `org.springframework.expression.spel.ast.InlineList#isConstant`, `org.springframework.expression.spel.ast.InlineMap#<init>`, `org.springframework.expression.spel.ast.InlineMap#createMap`, `org.springframework.expression.spel.ast.InlineMap#determineIfConstant`, `org.springframework.expression.spel.ast.InlineMap#getConstantValue`, `org.springframework.expression.spel.ast.InlineMap#getValueInternal`, `org.springframework.expression.spel.ast.InlineMap#isConstant`, `org.springframework.expression.spel.ast.MethodReference#<init>`, `org.springframework.expression.spel.ast.MethodReference#getValue`, `org.springframework.expression.spel.ast.MethodReference#getValueInternal`, `org.springframework.expression.spel.ast.OpAnd#getValueInternal`, `org.springframework.expression.spel.ast.OpDec#getValueInternal`, `org.springframework.expression.spel.ast.OpDivide#getValueInternal`, `org.springframework.expression.spel.ast.OpEQ#getValueInternal`, `org.springframework.expression.spel.ast.OpGE#getValueInternal`, `org.springframework.expression.spel.ast.OpGT#getValueInternal`, `org.springframework.expression.spel.ast.OpInc#getValueInternal`, `org.springframework.expression.spel.ast.OpLE#getValueInternal`, `org.springframework.expression.spel.ast.OpLT#getValueInternal`, `org.springframework.expression.spel.ast.OpMinus#<init>`, `org.springframework.expression.spel.ast.OpMinus#getValueInternal`, `org.springframework.expression.spel.ast.OpModulus#getValueInternal`, `org.springframework.expression.spel.ast.OpMultiply#getValueInternal`, `org.springframework.expression.spel.ast.OpNE#getValueInternal`, `org.springframework.expression.spel.ast.OpOr#getValueInternal`, `org.springframework.expression.spel.ast.OpPlus#concatenate`, `org.springframework.expression.spel.ast.OpPlus#getValueInternal`, `org.springframework.expression.spel.ast.OperatorBetween#getValueInternal`, `org.springframework.expression.spel.ast.OperatorInstanceof#getValueInternal`, `org.springframework.expression.spel.ast.OperatorMatches#getValueInternal`, `org.springframework.expression.spel.ast.OperatorNot#getValueInternal`, `org.springframework.expression.spel.ast.OperatorPower#getValueInternal`, `org.springframework.expression.spel.ast.Projection#getValueRef`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#<init>`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValue`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValueInternal`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValueRef`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#readProperty`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#setValue`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#setValueInternal`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#writeProperty`, `org.springframework.expression.spel.ast.Selection#getValueRef`, `org.springframework.expression.spel.ast.Ternary#getValueInternal`, `org.springframework.expression.spel.ast.TypeReference#getValueInternal` | diff |
| CVE-2026-41852 | org.springframework:spring-expression | 5.3.4 | LOW | UNKNOWN |  |  |
| CVE-2026-41853 | org.springframework:spring-webmvc | 5.3.4 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41901 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-41901 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-47838 | org.springframework.security:spring-security-web | 5.4.5 | MODERATE | UNKNOWN |  |  |
| CVE-2026-47884 | org.springframework:spring-webmvc | 5.3.4 | CRITICAL | FRAMEWORK | `org.springframework.web.servlet.view.xslt.XsltView#getStylesheetSource` | diff |
| CVE-2026-49844 | org.apache.logging.log4j:log4j-api | 2.13.3 | MODERATE | FRAMEWORK | `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatDouble`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatDoubleArray`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatFloat`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatFloatArray`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatNumber` | diff |
| CVE-2026-50193 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.node.InternalNodeMapper#_serializeNonRecursive`, `com.fasterxml.jackson.databind.node.InternalNodeMapper#popOrNull`, `com.fasterxml.jackson.databind.node.InternalNodeMapper#push` | diff |
| CVE-2026-54512 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.databind.DatabindContext#_resolveAndValidateGeneric`, `com.fasterxml.jackson.databind.DatabindContext#_validateTypeParameter`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType` | diff |
| CVE-2026-54513 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#<init>`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#allowIfSubTypeIsArray`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#build`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#<init>`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#allowIfSubTypeIsArray`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#build`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType` | diff |
| CVE-2026-54514 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#_inetSocketAddress` | diff |
| CVE-2026-54515 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `tools.jackson.databind.deser.bean.BeanDeserializerBase#createContextual` | diff |
| CVE-2026-68494 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_finishNumberIntegralPart`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_startNegativeNumber`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_startPositiveNumber` | diff |
| CVE-2026-77310 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#_deserialize`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_convertDottedQuadToHex`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_ipStringToBytes`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_parseHextet`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_parseOctet`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_textToNumericFormatV4`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_textToNumericFormatV6`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#isInetAddress` | diff |
| CVE-2026-83557 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.jsontype.DefaultBaseTypeLimitingValidator#validateSubType`, `com.fasterxml.jackson.databind.jsontype.DefaultBaseTypeLimitingValidator$UnsafeBaseTypes#<init>` | llm (diff) |
| CVE-2026-89425 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.core.json.ReaderBasedJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_checkMatchEnd`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_handleUnexpectedValue`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_checkMatchEnd`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_handleUnexpectedValue`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_finishErrorToken` | diff |
| CVE-2026-91776 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.jsontype.impl.TypeDeserializerBase#_findDeserializer` | diff |
| CVE-2026-91777 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#add`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#handleUnresolvedReference`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#resolve`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#resolveForwardReference`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#handleUnresolvedReference`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#put`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#resolve`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#resolveForwardReference` | diff |
| CVE-2026-9828 | ch.qos.logback:logback-core | 1.2.3 | LOW | UNKNOWN |  |  |



### Notes on individual results

- **CVE-2023-1932** (UNKNOWN): LLM (z-ai/glm-5.3) suggested org.hibernate.validator.internal.constraintvalidators.hv.SafeHtmlValidator#isValid from the advisory text only (no fix diff); not reachable, but unverified

- **CVE-2024-22262** (UNKNOWN): LLM (z-ai/glm-5.3) suggested org.springframework.web.util.UriComponentsBuilder#fromHttpUrl, org.springframework.web.util.UriComponentsBuilder#fromUriString from the advisory text only (no fix diff); not reachable, but unverified


## REACHABLE — full call paths


### CVE-2013-7285

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson is a deliberately vulnerable lesson whose completed(String) handler receives attacker-supplied lesson input and passes it directly to XStream#fromXML, which is the known CVE-2013-7285 command-injection sink.


### CVE-2016-3674

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes an attacker-supplied lesson input (String) and passes it directly to XStream#fromXML, which is the classic XXE sink; the lesson is specifically designed to demonstrate vulnerable component usage with user-controlled XML.


### CVE-2017-7957

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson is a deliberately vulnerable exercise whose completed(String) method takes attacker-supplied HTTP request input and passes it directly to XStream#fromXML, exactly the pattern the CVE describes.


### CVE-2020-26217

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied HTTP parameter string and passes it directly to XStream#fromXML, which is the documented vulnerable unmarshal entry point for CVE-2020-26217.


### CVE-2020-26258

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied lesson answer string and passes it directly to XStream#fromXML, which is the vulnerable unmarshal entry point for CVE-2020-26258.


### CVE-2020-26259

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson passes attacker-supplied lesson input (an HTTP request parameter) directly into XStream#fromXML, which unmarshals it via the CVE-affected unmarshal path.


### CVE-2021-21341

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes the user-submitted lesson input string and passes it directly to XStream.fromXML, so attacker-controlled HTTP input reaches the vulnerable unmarshal call.


### CVE-2021-21342

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes the user-submitted lesson answer (an HTTP request parameter) and passes it directly to XStream#fromXML, so attacker-controlled XML reaches the vulnerable unmarshal call.


### CVE-2021-21343

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes a user-supplied lesson answer string and passes it directly to XStream#fromXML, which unmarshalls attacker-controlled XML — exactly the input path for CVE-2021-21343.


### CVE-2021-21344

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson.completed takes an attacker-supplied lesson input string and passes it directly to XStream.fromXML, which is exactly the unmarshal path for CVE-2021-21344.


### CVE-2021-21345

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes attacker-supplied lesson input (the 'completed' endpoint receives user-submitted XML from the HTTP request) and passes it directly to XStream#fromXML, which is exactly the vulnerable unmarshal path for CVE-2021-21345.


### CVE-2021-21346

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson passes the attacker-supplied lesson parameter (String) directly into XStream#fromXML, which is exactly the vulnerable unmarshal entry point for CVE-2021-21346.


### CVE-2021-21347

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson is a deliberately vulnerable exercise whose completed(String) method passes attacker-supplied request input directly to XStream#fromXML, which is the documented vulnerable unmarshal path for CVE-2021-21347.


### CVE-2021-21348

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson.completed(String) takes attacker-supplied lesson input (HTTP form parameter) and passes it directly to XStream.fromXML, which is exactly the intended vulnerable-component demo for this XStream CVE.


### CVE-2021-21349

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes the attacker-supplied lesson submission string and passes it directly to XStream#fromXML, which is exactly the unmarshal path for CVE-2021-21349.


### CVE-2021-21350

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — This is WebGoat's VulnerableComponentsLesson, whose completed(String) method parses the attacker-supplied HTTP parameter (the XStream payload field) directly via XStream#fromXML, so the CVE-2021-21350 unmarshal sink receives user-controlled input with no sanitization in between.


### CVE-2021-21351

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson passes attacker-supplied lesson input (an HTTP request parameter) directly to XStream.fromXML, which is exactly the vulnerable unmarshal path for CVE-2021-21351.


### CVE-2021-29505

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied HTTP parameter (the lesson's attack input) and passes it directly to XStream#fromXML, which is exactly the vulnerable unmarshal entry point for CVE-2021-29505.


### CVE-2021-39139

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes an attacker-supplied lesson parameter (String) and passes it directly to XStream.fromXML, which is the documented vulnerable unmarshal entry point for CVE-2021-39139.


### CVE-2021-39140

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes attacker-supplied lesson input (an HTTP request parameter) and passes it directly to XStream.fromXML, which is exactly the intended vulnerable demonstration of this CVE.


### CVE-2021-39141

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied lesson input string and passes it directly to XStream#fromXML, which is exactly the unmarshal path targeted by CVE-2021-39141.


### CVE-2021-39144

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson is a deliberately vulnerable teaching exercise where the lesson's completed(String) handler passes attacker-supplied HTTP input directly to XStream#fromXML, which is exactly the CVE-2021-39144 unmarshal entry point.


### CVE-2021-39145

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes a String parameter that is attacker-supplied HTTP input from the lesson form and passes it directly to XStream#fromXML, which is exactly the unmarshal path targeted by CVE-2021-39145.


### CVE-2021-39146

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied String (the lesson's HTTP form input) and passes it directly to XStream#fromXML, which internally calls the vulnerable unmarshal; this is the intended vulnerable-by-design demonstration of CVE-2021-39146.


### CVE-2021-39147

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson.completed(String) takes attacker-supplied lesson input and passes it directly to XStream.fromXML, which is exactly the unmarshal path for CVE-2021-39147.


### CVE-2021-39148

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied HTTP parameter (the lesson's answer field) and passes it directly to XStream#fromXML, which is exactly the vulnerable unmarshal entry point for CVE-2021-39148.


### CVE-2021-39149

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes an attacker-supplied lesson input string and passes it directly to XStream.fromXML, which unmarshals attacker-controlled XML — exactly the input path CVE-2021-39149 requires.


### CVE-2021-39150

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes the attacker-supplied lesson input String and feeds it directly to XStream.fromXML, so unmarshalling of untrusted attacker-controlled XML reaches the vulnerable XStream unmarshal path.


### CVE-2021-39151

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson is a deliberately vulnerable exercise whose completed(String) method takes attacker-supplied HTTP request input and passes it directly to XStream#fromXML, which internally calls the vulnerable unmarshal. This is the exact known exploitable path for this CVE in WebGoat.


### CVE-2021-39152

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed(String) is an exercise endpoint that feeds attacker-supplied lesson input directly into XStream.fromXML(String), which is exactly the XStream SSRF CVE-2021-39152 pattern of unmarshalling attacker-controlled XML.


### CVE-2021-39153

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes attacker-supplied lesson input (an HTTP request parameter) and passes it directly to XStream#fromXML, which is exactly the untrusted deserialization entry point for CVE-2021-39153.


### CVE-2021-39154

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied HTTP parameter string and passes it directly to XStream#fromXML, which is exactly the unmarshal path CVE-2021-39154 targets.


### CVE-2021-43859

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson takes user-supplied request input (the String parameter from the HTTP form) and passes it directly to XStream.fromXML, so attacker-controlled XML reaches the vulnerable unmarshal path.


### CVE-2022-22976

```
org.owasp.webgoat.WebSecurityConfig#configureGlobal(Lorg/springframework/security/config/annotation/authentication/builders/AuthenticationManagerBuilder;)V
 → org.springframework.security.config.annotation.authentication.builders.AuthenticationManagerBuilder#userDetailsService(Lorg/springframework/security/core/userdetails/UserDetailsService;)Lorg/springframework/security/config/annotation/authentication/configurers/userdetails/DaoAuthenticationConfigurer;
 → org.springframework.security.config.annotation.authentication.configurers.userdetails.DaoAuthenticationConfigurer#<init>(Lorg/springframework/security/core/userdetails/UserDetailsService;)V
 → org.springframework.security.config.annotation.authentication.configurers.userdetails.AbstractDaoAuthenticationConfigurer#<init>(Lorg/springframework/security/core/userdetails/UserDetailsService;)V
 → org.springframework.security.authentication.dao.DaoAuthenticationProvider#<init>()V
 → org.springframework.security.crypto.factory.PasswordEncoderFactories#createDelegatingPasswordEncoder()Lorg/springframework/security/crypto/password/PasswordEncoder;
 → org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder#<init>()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#configureGlobal(Lorg/springframework/security/config/annotation/authentication/builders/AuthenticationManagerBuilder;)V` → `org.springframework.security.config.annotation.authentication.builders.AuthenticationManagerBuilder#userDetailsService(Lorg/springframework/security/core/userdetails/UserDetailsService;)Lorg/springframework/security/config/annotation/authentication/configurers/userdetails/DaoAuthenticationConfigurer;`


**Exploitability (z-ai/glm-5.3):** unlikely — The path is a Spring Security configuration-time setup (configureGlobal -> DaoAuthenticationProvider -> createDelegatingPasswordEncoder -> BCryptPasswordEncoder#<init>) invoked once at startup, not on any attacker-controlled request path. Additionally, the default BCryptPasswordEncoder constructor uses the default work factor (not 31), which the advisory explicitly states is unaffected by this CVE.


### CVE-2022-40151

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson#completed takes an attacker-supplied lesson input string and passes it directly to XStream#fromXML, which is exactly the unmarshal path targeted by CVE-2022-40151, so attacker-controlled XML can reach the vulnerable call.


### CVE-2022-41966

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — WebGoat's VulnerableComponentsLesson.completed takes a user-submitted String (the lesson's attack input from the HTTP request) and passes it directly to XStream.fromXML, so attacker-controlled XML reaches the vulnerable unmarshal call.


### CVE-2024-38820

```
org.owasp.webgoat.users.UserService#createLessonsForUser(Lorg/owasp/webgoat/users/WebGoatUser;)V
 → org.springframework.jdbc.core.JdbcTemplate#execute(Ljava/lang/String;)V
 → org.springframework.jdbc.core.JdbcTemplate#execute(Lorg/springframework/jdbc/core/StatementCallback;Z)Ljava/lang/Object;
 → org.springframework.jdbc.core.JdbcTemplate$1QueryStatementCallback#doInStatement(Ljava/sql/Statement;)Ljava/lang/Object;
 → org.springframework.jdbc.core.JdbcOperationsExtensionsKt$query$1#extractData(Ljava/sql/ResultSet;)Ljava/lang/Object;
 → org.springframework.security.config.web.server.ServerContentSecurityPolicyDsl$get$1#invoke(Lorg/springframework/security/config/web/server/ServerHttpSecurity$HeaderSpec$ContentSecurityPolicySpec;)V
 → org.springframework.security.config.web.server.ServerHttpSecurity$HeaderSpec$ContentSecurityPolicySpec#policyDirectives(Ljava/lang/String;)Lorg/springframework/security/config/web/server/ServerHttpSecurity$HeaderSpec;
 → org.springframework.security.web.server.header.ContentSecurityPolicyServerHttpHeadersWriter#setPolicyDirectives(Ljava/lang/String;)V
 → org.springframework.security.web.server.header.ContentSecurityPolicyServerHttpHeadersWriter#createDelegate()Lorg/springframework/security/web/server/header/ServerHttpHeadersWriter;
 → org.springframework.security.web.server.header.StaticServerHttpHeadersWriter#builder()Lorg/springframework/security/web/server/header/StaticServerHttpHeadersWriter$Builder;
 → org.springframework.security.web.server.header.StaticServerHttpHeadersWriter$Builder#<init>()V
 → org.springframework.http.HttpHeaders#<init>()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.users.UserService#createLessonsForUser(Lorg/owasp/webgoat/users/WebGoatUser;)V` → `org.springframework.jdbc.core.JdbcTemplate#execute(Ljava/lang/String;)V`


**Exploitability (z-ai/glm-5.3):** unlikely — The path starts from an internal user-creation routine doing fixed SQL inserts via JdbcTemplate, and the steps through ServerHttpSecurity CSP configuration to HttpHeaders#<init>() carry only internal/config data with no attacker-controlled HTTP input reaching the constructor. The listed 'vulnerable method' here is just a trivial HttpHeaders constructor invoked during fixed security-header setup, so the CVE's DataBinder disallowedFields locale-matching issue is not meaningfully exercised.


### CVE-2024-38821

```
org.owasp.webgoat.users.UserService#createLessonsForUser(Lorg/owasp/webgoat/users/WebGoatUser;)V
 → org.springframework.jdbc.core.JdbcTemplate#execute(Ljava/lang/String;)V
 → org.springframework.jdbc.core.JdbcTemplate#execute(Lorg/springframework/jdbc/core/StatementCallback;Z)Ljava/lang/Object;
 → org.springframework.jdbc.core.JdbcTemplate$1QueryStatementCallback#doInStatement(Ljava/sql/Statement;)Ljava/lang/Object;
 → org.springframework.jdbc.core.JdbcOperationsExtensionsKt$query$1#extractData(Ljava/sql/ResultSet;)Ljava/lang/Object;
 → org.springframework.security.config.web.server.AuthorizeExchangeDsl$get$1#invoke(Lorg/springframework/security/config/web/server/ServerHttpSecurity$AuthorizeExchangeSpec;)V
 → org.springframework.security.config.web.server.ServerHttpSecurity$AuthorizeExchangeSpec#matchers([Lorg/springframework/security/web/server/util/matcher/ServerWebExchangeMatcher;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.users.UserService#createLessonsForUser(Lorg/owasp/webgoat/users/WebGoatUser;)V` → `org.springframework.jdbc.core.JdbcTemplate#execute(Ljava/lang/String;)V`


**Exploitability (z-ai/glm-5.3):** unlikely — The path runs WebGoat's UserService#createLessonsForUser through JdbcTemplate SQL execution and then inexplicably jumps into Spring Security's Kotlin DSL configuration methods (AuthorizeExchangeDsl$get$1#invoke, ServerHttpSecurity$AuthorizeExchangeSpec#matchers), which are build-time configuration code, not request-processing code. The JDBC-to-security-config transition is not a plausible data flow, and these DSL/matcher methods carry fixed configuration rather than attacker-controlled input.


### CVE-2024-47072

```
org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;
 → com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.vulnerable_components.VulnerableComponentsLesson#completed(Ljava/lang/String;)Lorg/owasp/webgoat/assignments/AttackResult;` → `com.thoughtworks.xstream.XStream#fromXML(Ljava/lang/String;)Ljava/lang/Object;`


**Exploitability (z-ai/glm-5.3):** likely — The WebGoat VulnerableComponentsLesson#completed takes an attacker-supplied lesson answer string and passes it directly to XStream#fromXML, so external input reaches the unmarshal call.


### CVE-2026-41848

```
org.owasp.webgoat.WebSecurityConfig#configure(Lorg/springframework/security/config/annotation/web/builders/HttpSecurity;)V
 → org.springframework.security.config.annotation.web.builders.HttpSecurity#formLogin()Lorg/springframework/security/config/annotation/web/configurers/FormLoginConfigurer;
 → org.springframework.security.config.annotation.web.configurers.FormLoginConfigurer#<init>()V
 → org.springframework.security.config.annotation.web.configurers.AbstractAuthenticationFilterConfigurer#<init>(Lorg/springframework/security/web/authentication/AbstractAuthenticationProcessingFilter;Ljava/lang/String;)V
 → org.springframework.security.config.annotation.web.configurers.AbstractAuthenticationFilterConfigurer#loginProcessingUrl(Ljava/lang/String;)Lorg/springframework/security/config/annotation/web/configurers/AbstractAuthenticationFilterConfigurer;
 → org.springframework.security.config.annotation.web.configurers.FormLoginConfigurer#createLoginProcessingUrlMatcher(Ljava/lang/String;)Lorg/springframework/security/web/util/matcher/RequestMatcher;
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher#<init>(Ljava/lang/String;Ljava/lang/String;)V
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher#<init>(Ljava/lang/String;Ljava/lang/String;Z)V
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher#<init>(Ljava/lang/String;Ljava/lang/String;ZLorg/springframework/web/util/UrlPathHelper;)V
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher$SpringAntMatcher#<init>(Ljava/lang/String;ZLorg/springframework/security/web/util/matcher/AntPathRequestMatcher$1;)V
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher$SpringAntMatcher#<init>(Ljava/lang/String;Z)V
 → org.springframework.security.web.util.matcher.AntPathRequestMatcher$SpringAntMatcher#createMatcher(Z)Lorg/springframework/util/AntPathMatcher;
 → org.springframework.util.AntPathMatcher#<init>()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#configure(Lorg/springframework/security/config/annotation/web/builders/HttpSecurity;)V` → `org.springframework.security.config.annotation.web.builders.HttpSecurity#formLogin()Lorg/springframework/security/config/annotation/web/configurers/FormLoginConfigurer;`


**Exploitability (z-ai/glm-5.3):** unlikely — The pattern supplied to AntPathMatcher comes from loginProcessingUrl with a hardcoded application constant (WebSecurityConfig's configure method), not from attacker input. The attacker-controlled path is matched at request time, but the vulnerable pattern here is fixed configuration data.




## FRAMEWORK — needs manual review

These CVEs' vulnerable methods are only reachable by tracing *through*
framework-package code (Spring MVC, Tomcat, Jackson's `@RequestBody`
handling, ...), which ariadne's static call graph cannot follow past —
that code is driven by reflection/dispatch tables the framework builds at
runtime. A human should check whether the app's own configuration
actually exposes this path.


### CVE-2020-36518

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Data's ResourceReaderRepositoryPopulator/Jackson2ResourceReader genuinely reads JSON repository seed files via Jackson's ObjectReader, and untyped JSON content is deserialized by UntypedObjectDeserializer, so the final hops are real. The event-multicaster chain (DataSourceInitializerInvoker publishing ContextRefreshedEvent reaching the repository populator) is standard Spring wiring, not a CHA artifact.


### CVE-2021-22118

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#applyBeanPostProcessorsAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcessAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcess(Lorg/springframework/security/web/server/WebFilterChainProxy;)Lorg/springframework/security/web/server/WebFilterChainProxy;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#lambda$postProcess$0(Lorg/springframework/web/server/ServerWebExchange;Lorg/springframework/web/server/WebFilterChain;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#filter(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#lambda$filter$0(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.adapter.HttpWebHandlerAdapter#handle(Lorg/springframework/http/server/reactive/ServerHttpRequest;Lorg/springframework/http/server/reactive/ServerHttpResponse;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.adapter.HttpWebHandlerAdapter#createExchange(Lorg/springframework/http/server/reactive/ServerHttpRequest;Lorg/springframework/http/server/reactive/ServerHttpResponse;)Lorg/springframework/web/server/ServerWebExchange;
 → org.springframework.web.server.adapter.DefaultServerWebExchange#<init>(Lorg/springframework/http/server/reactive/ServerHttpRequest;Lorg/springframework/http/server/reactive/ServerHttpResponse;Lorg/springframework/web/server/session/WebSessionManager;Lorg/springframework/http/codec/ServerCodecConfigurer;Lorg/springframework/web/server/i18n/LocaleContextResolver;Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.server.adapter.DefaultServerWebExchange#initFormData(Lorg/springframework/http/server/reactive/ServerHttpRequest;Lorg/springframework/http/codec/ServerCodecConfigurer;Ljava/lang/String;)Lreactor/core/publisher/Mono;
 → org.springframework.http.codec.multipart.MultipartHttpMessageReader#readMono(Lorg/springframework/core/ResolvableType;Lorg/springframework/http/ReactiveHttpInputMessage;Ljava/util/Map;)Lreactor/core/publisher/Mono;
 → org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#read(Lorg/springframework/core/ResolvableType;Lorg/springframework/http/ReactiveHttpInputMessage;Ljava/util/Map;)Lreactor/core/publisher/Flux;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The latter half (HttpWebHandlerAdapter.handle → DefaultServerWebExchange.<init> → initFormData → MultipartHttpMessageReader.readMono → DefaultPartHttpMessageReader.read) is the genuine WebFlux request-processing chain that parses multipart form data, and the front half (WebSecurityConfigurerAdapter.authenticationManager → AuthenticationConfiguration → lazyBean → postProcess initialization) is standard Spring Security configuration wiring; these are real framework behaviors, even if the linking of servlet-era config to the reactive stack in WebGoat is unusual.


### CVE-2021-3690

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.websockets.core.BufferedBinaryMessage$1#handleEvent(Lio/undertow/websockets/core/StreamSourceFrameChannel;)V
 → io.undertow.websockets.core.AbstractReceiveListener$1#complete(Lio/undertow/websockets/core/WebSocketChannel;Lio/undertow/websockets/core/BufferedBinaryMessage;)V
 → io.undertow.websockets.jsr.FrameHandler#onFullPongMessage(Lio/undertow/websockets/core/WebSocketChannel;Lio/undertow/websockets/core/BufferedBinaryMessage;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Security/Spring Boot setup portion (nodes 1-17) plausibly reaches Undertow deployment, but the tail is a CHA mishmash: handleFirstRequest would not dispatch to ServerSentEventHandler, and a generic flush()/invokeChannelListener edge connecting SSE sink channels to the WebSocket BufferedBinaryMessage/FrameHandler pong callback is exactly the kind of spurious listener-interface hop over-approximation produces.


### CVE-2021-3859

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.server.protocol.framed.AbstractFramedChannel$3#handleEvent(Lio/undertow/server/protocol/framed/AbstractFramedChannel;)V
 → io.undertow.server.protocol.framed.AbstractFramedChannel#receive()Lio/undertow/server/protocol/framed/AbstractFramedStreamSourceChannel;
 → io.undertow.protocols.http2.Http2Channel#parseFrame(Ljava/nio/ByteBuffer;)Lio/undertow/server/protocol/framed/FrameHeaderData;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from DeploymentManagerImpl#deploy() to ServletInitialHandler$2#call (request-time handling) is a CHA artifact — deployment does not invoke request handlers; likewise handleFirstRequest → ServerSentEventHandler#handleRequest connects unrelated handlers via over-broad dispatch. The Spring setup portion (1–17) is plausible, but the bridge into the vulnerable parseFrame relies on spurious edges.


### CVE-2021-42550

```
org.owasp.webgoat.AsciiDoctorTemplateResolver#<clinit>()V
 → org.slf4j.LoggerFactory#getLogger(Ljava/lang/Class;)Lorg/slf4j/Logger;
 → org.slf4j.LoggerFactory#getLogger(Ljava/lang/String;)Lorg/slf4j/Logger;
 → org.slf4j.LoggerFactory#getILoggerFactory()Lorg/slf4j/ILoggerFactory;
 → org.slf4j.impl.StaticLoggerBinder#getLoggerFactory()Lorg/slf4j/ILoggerFactory;
 → ch.qos.logback.classic.selector.ContextJNDISelector#getLoggerContext()Lch/qos/logback/classic/LoggerContext;
 → ch.qos.logback.classic.selector.ContextJNDISelector#findConfigFileURL(Ljavax/naming/Context;Lch/qos/logback/classic/LoggerContext;)Ljava/net/URL;
```

**Entry call (app → dependency):** `org.owasp.webgoat.AsciiDoctorTemplateResolver#<clinit>()V` → `org.slf4j.LoggerFactory#getLogger(Ljava/lang/Class;)Lorg/slf4j/Logger;`


**Path plausibility (z-ai/glm-5.3):** plausible — LoggerFactory#getLogger -> StaticLoggerBinder#getLoggerFactory is the standard real SLF4J-to-logback binding path, and ContextJNDISelector is logback's actual JNDI-based selector whose getLoggerContext genuinely calls findConfigFileURL to locate config via JNDI (the CVE's attack vector). The only conditional is that the app is configured to use the JNDI selector, which is exactly the configuration-dependent scenario the advisory describes.


### CVE-2022-1471

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.boot.env.EnvironmentPostProcessorApplicationListener#onApplicationEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.boot.env.EnvironmentPostProcessorApplicationListener#onApplicationEnvironmentPreparedEvent(Lorg/springframework/boot/context/event/ApplicationEnvironmentPreparedEvent;)V
 → org.springframework.boot.cloud.CloudFoundryVcapEnvironmentPostProcessor#postProcessEnvironment(Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplication;)V
 → org.springframework.boot.cloud.CloudFoundryVcapEnvironmentPostProcessor#getPropertiesFromApplication(Lorg/springframework/core/env/Environment;Lorg/springframework/boot/json/JsonParser;)Ljava/util/Properties;
 → org.springframework.boot.json.YamlJsonParser#parseMap(Ljava/lang/String;)Ljava/util/Map;
 → org.springframework.boot.json.YamlJsonParser#lambda$parseMap$0(Ljava/lang/String;)Ljava/util/Map;
 → org.yaml.snakeyaml.Yaml#loadAs(Ljava/lang/String;Ljava/lang/Class;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain relies on CHA over-broad dispatch through generic framework interfaces: InitializingBean#afterPropertiesSet (hop 8→9) and ApplicationListener#onApplicationEvent (hop 14→15) connect unrelated classes, and EnvironmentPostProcessorApplicationListener only handles ApplicationEnvironmentPreparedEvent, not the DataSourceInitializer events published at hop 11. While the final segment (CloudFoundryVcapEnvironmentPostProcessor → YamlJsonParser → Yaml#loadAs) is real Spring Boot code, the path as a whole is stitched together by spurious interface-dispatch edges and does not correspond to an actual runtime call chain from the security configuration entry point.


### CVE-2022-22950

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.transaction.interceptor.NameMatchTransactionAttributeSource#afterPropertiesSet()V
 → org.springframework.transaction.interceptor.DefaultTransactionAttribute#resolveAttributeStrings(Lorg/springframework/util/StringValueResolver;)V
 → org.springframework.beans.factory.config.EmbeddedValueResolver#resolveStringValue(Ljava/lang/String;)Ljava/lang/String;
 → org.springframework.context.expression.StandardBeanExpressionResolver#evaluate(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanExpressionContext;)Ljava/lang/Object;
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.ConstructorReference#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
 → org.springframework.expression.spel.ast.ConstructorReference#createArray(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — Hop 8→9 is a classic CHA over-approximation: invokeInitMethods calls InitializingBean#afterPropertiesSet on whatever bean is being initialized (here an AuthenticationManager from lazyBean), and CHA links it to every implementer, including NameMatchTransactionAttributeSource, which has no relationship to this bean. The rest of the chain (afterPropertiesSet → SpEL evaluation of a transaction attribute string) is internally plausible, but the specific bean being initialized in this path would never be a transaction-attribute source.


### CVE-2022-22965

```
org.owasp.webgoat.users.UserValidator#validate(Ljava/lang/Object;Lorg/springframework/validation/Errors;)V
 → org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V
 → org.springframework.validation.AbstractPropertyBindingResult#getActualFieldValue(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#getPropertyValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.beans.BeanWrapperImpl#getLocalPropertyHandler(Ljava/lang/String;)Lorg/springframework/beans/BeanWrapperImpl$BeanPropertyHandler;
 → org.springframework.beans.BeanWrapperImpl#getCachedIntrospectionResults()Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#forClass(Ljava/lang/Class;)Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#<init>(Ljava/lang/Class;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.users.UserValidator#validate(Ljava/lang/Object;Lorg/springframework/validation/Errors;)V` → `org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain rejectValue → getActualFieldValue → AbstractNestablePropertyAccessor → BeanWrapperImpl.getCachedIntrospectionResults → CachedIntrospectionResults.forClass/<init> is the genuine Spring property-binding/introspection flow used when resolving field values during validation, matching exactly how Spring4Shell reaches CachedIntrospectionResults.


### CVE-2022-22970

```
org.owasp.webgoat.users.UserValidator#validate(Ljava/lang/Object;Lorg/springframework/validation/Errors;)V
 → org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V
 → org.springframework.validation.AbstractPropertyBindingResult#getActualFieldValue(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#getPropertyValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.beans.BeanWrapperImpl#getLocalPropertyHandler(Ljava/lang/String;)Lorg/springframework/beans/BeanWrapperImpl$BeanPropertyHandler;
 → org.springframework.beans.BeanWrapperImpl#getCachedIntrospectionResults()Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#forClass(Ljava/lang/Class;)Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#<init>(Ljava/lang/Class;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.users.UserValidator#validate(Ljava/lang/Object;Lorg/springframework/validation/Errors;)V` → `org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — This is a genuine Spring data-binding flow: rejectValue on an AbstractPropertyBindingResult resolves the actual field value via BeanWrapperImpl, which loads CachedIntrospectionResults.forClass to introspect the target bean — exactly the code the CVE describes. The hops follow real, concrete Spring implementation calls, not generic interface dispatch.


### CVE-2022-22978

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.security.web.FilterChainProxy#afterPropertiesSet()V
 → org.springframework.security.config.http.DefaultFilterChainValidator#validate(Lorg/springframework/security/web/FilterChainProxy;)V
 → org.springframework.security.config.http.DefaultFilterChainValidator#checkLoginPageIsntProtected(Lorg/springframework/security/web/FilterChainProxy;Ljava/util/List;)V
 → org.springframework.security.web.access.intercept.DefaultFilterInvocationSecurityMetadataSource#getAttributes(Ljava/lang/Object;)Ljava/util/Collection;
 → org.springframework.security.web.util.matcher.RegexRequestMatcher#matches(Ljavax/servlet/http/HttpServletRequest;)Z
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain from WebSecurityConfigurerAdapter#authenticationManager through AuthenticationConfiguration's lazy bean post-processing to FilterChainProxy#afterPropertiesSet is a genuine Spring Security bootstrap flow, and DefaultFilterChainValidator really does inspect the filter chains' security metadata source, which invokes each RequestMatcher's matches() — WebGoat is known to use RegexRequestMatcher in its security config.


### CVE-2022-25857

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.boot.env.EnvironmentPostProcessorApplicationListener#onApplicationEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.boot.env.EnvironmentPostProcessorApplicationListener#onApplicationEnvironmentPreparedEvent(Lorg/springframework/boot/context/event/ApplicationEnvironmentPreparedEvent;)V
 → org.springframework.boot.cloud.CloudFoundryVcapEnvironmentPostProcessor#postProcessEnvironment(Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplication;)V
 → org.springframework.boot.cloud.CloudFoundryVcapEnvironmentPostProcessor#getPropertiesFromApplication(Lorg/springframework/core/env/Environment;Lorg/springframework/boot/json/JsonParser;)Ljava/util/Properties;
 → org.springframework.boot.json.YamlJsonParser#parseMap(Ljava/lang/String;)Ljava/util/Map;
 → org.springframework.boot.json.YamlJsonParser#lambda$parseMap$0(Ljava/lang/String;)Ljava/util/Map;
 → org.yaml.snakeyaml.Yaml#loadAs(Ljava/lang/String;Ljava/lang/Class;)Ljava/lang/Object;
 → org.yaml.snakeyaml.Yaml#loadFromReader(Lorg/yaml/snakeyaml/reader/StreamReader;Ljava/lang/Class;)Ljava/lang/Object;
 → org.yaml.snakeyaml.composer.Composer#<init>(Lorg/yaml/snakeyaml/parser/Parser;Lorg/yaml/snakeyaml/resolver/Resolver;Lorg/yaml/snakeyaml/LoaderOptions;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from SimpleApplicationEventMulticaster to EnvironmentPostProcessorApplicationListener is a CHA artifact: that listener only handles ApplicationEnvironmentPreparedEvent (fired very early during environment preparation), not the DataSourceInitializerInvoker events published from bean initialization, so the two halves of the chain are unrelated. Additionally, CloudFoundryVcapEnvironmentPostProcessor only activates on Cloud Foundry, so even the YamlJsonParser->snakeyaml tail is conditional on an environment WebGoat would not have.


### CVE-2022-42003

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The tail of the path is a genuine, documented Spring Data flow: ResourceReaderRepositoryPopulator uses Jackson2ResourceReader to deserialize repository bootstrap JSON via ObjectReader#readTree, which really dispatches into EnumDeserializer#deserialize. The earlier hops are Spring Security/bean-initialization wiring that plausibly triggers the ContextRefreshedEvent consumed by the repository populator.


### CVE-2022-42004

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The Spring Data repository populator (Jackson2ResourceReader reading a JSON resource via ObjectReader#readTree) is a real framework feature that genuinely deserializes JSON with Jackson, and the event-multicast hop into AbstractRepositoryPopulatorFactoryBean is standard Spring listener dispatch. The only soft spot is that EnumDeserializer#deserialize requires the populated JSON to contain an enum-typed property, but that is a normal occurrence rather than a CHA artifact.


### CVE-2023-1108

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.conduits.ConduitStreamSinkChannel#flush()Z
 → io.undertow.protocols.ssl.SslConduit#flush()Z
 → io.undertow.protocols.ssl.SslConduit#doWrap([Ljava/nio/ByteBuffer;II)J
 → io.undertow.protocols.ssl.SslConduit#wrapAndFlip([Ljava/nio/ByteBuffer;II)Ljavax/net/ssl/SSLEngineResult;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The early hops (Spring Boot onRefresh → createWebServer → UndertowServletWebServerFactory → DeploymentManagerImpl#deploy) are a real embedded-Undertow startup chain, but the jump from deploy() to ServletInitialHandler$2#call/handleFirstRequest (request-time dispatch) and then to ServerSentEventHandler#handleRequest is not a real call relationship — request handling never occurs during deployment, and SSE handler dispatch from generic request handling to SslConduit#flush is exactly the kind of spurious virtual-dispatch edge CHA produces.


### CVE-2023-1973

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.security.handlers.AuthenticationCallHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → io.undertow.security.impl.SecurityContextImpl#authenticate()Z
 → io.undertow.security.impl.SecurityContextImpl#authTransition()Z
 → io.undertow.security.impl.SecurityContextImpl#sendChallenges()Lio/undertow/security/impl/SecurityContextImpl$AuthenticationState;
 → io.undertow.security.impl.SecurityContextImpl$ChallengeSender#access$300(Lio/undertow/security/impl/SecurityContextImpl$ChallengeSender;)Lio/undertow/security/impl/SecurityContextImpl$AuthenticationState;
 → io.undertow.security.impl.SecurityContextImpl$ChallengeSender#transition()Lio/undertow/security/impl/SecurityContextImpl$AuthenticationState;
 → io.undertow.security.impl.FormAuthenticationMechanism#sendChallenge(Lio/undertow/server/HttpServerExchange;Lio/undertow/security/api/SecurityContext;)Lio/undertow/security/api/AuthenticationMechanism$ChallengeResult;
 → io.undertow.security.impl.FormAuthenticationMechanism#storeInitialLocation(Lio/undertow/server/HttpServerExchange;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The Undertow portion (SecurityContextImpl.authenticate → sendChallenges → ChallengeSender.transition → FormAuthenticationMechanism.sendChallenge → storeInitialLocation) is the genuine runtime security-challenge flow that fires on every unauthenticated request, and WebGoat really does use form login with embedded Undertow. The Spring prefix (lazyBean → postProcess → initializeBean → DelegatingFilterProxy init → context refresh → createWebServer) matches real Spring Boot bean-lifecycle and embedded-server wiring, so the chain corresponds to actual framework behavior rather than a CHA hop.


### CVE-2023-20860

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.servlet.handler.HandlerMappingIntrospector#afterPropertiesSet()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows real Spring Security configuration flow (WebSecurityConfigurerAdapter#authenticationManager -> AuthenticationConfiguration -> lazyBean -> postProcess -> initializeBean -> invokeInitMethods), and HandlerMappingIntrospector implements InitializingBean, so invokeInitMethods calling its afterPropertiesSet is genuine Spring bean-lifecycle dispatch, not a CHA artifact.


### CVE-2023-20861

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.transaction.interceptor.NameMatchTransactionAttributeSource#afterPropertiesSet()V
 → org.springframework.transaction.interceptor.DefaultTransactionAttribute#resolveAttributeStrings(Lorg/springframework/util/StringValueResolver;)V
 → org.springframework.beans.factory.config.EmbeddedValueResolver#resolveStringValue(Ljava/lang/String;)Ljava/lang/String;
 → org.springframework.context.expression.StandardBeanExpressionResolver#evaluate(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanExpressionContext;)Ljava/lang/Object;
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.OperatorMatches#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/spel/support/BooleanTypedValue;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows real Spring wiring: bean initialization invokes afterPropertiesSet, and NameMatchTransactionAttributeSource's attribute strings (e.g. 'PROPAGATION_REQUIRED, isolation:...') are resolved via EmbeddedValueResolver/StandardBeanExpressionResolver, which evaluates them as SpEL expressions. If such an attribute contains a 'matches' operator, OperatorMatches#getValueInternal is genuinely reached, so this is a real framework path rather than a CHA artifact.


### CVE-2023-20863

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.transaction.interceptor.NameMatchTransactionAttributeSource#afterPropertiesSet()V
 → org.springframework.transaction.interceptor.DefaultTransactionAttribute#resolveAttributeStrings(Lorg/springframework/util/StringValueResolver;)V
 → org.springframework.beans.factory.config.EmbeddedValueResolver#resolveStringValue(Ljava/lang/String;)Ljava/lang/String;
 → org.springframework.context.expression.StandardBeanExpressionResolver#evaluate(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanExpressionContext;)Ljava/lang/Object;
 → org.springframework.expression.common.TemplateAwareExpressionParser#parseExpression(Ljava/lang/String;Lorg/springframework/expression/ParserContext;)Lorg/springframework/expression/Expression;
 → org.springframework.expression.spel.standard.InternalSpelExpressionParser#doParseExpression(Ljava/lang/String;Lorg/springframework/expression/ParserContext;)Lorg/springframework/expression/spel/standard/SpelExpression;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — Bean initialization (invokeInitMethods) dispatching to InitializingBean#afterPropertiesSet is a real Spring mechanism, and NameMatchTransactionAttributeSource#resolveAttributeStrings genuinely resolves placeholder strings through EmbeddedValueResolver/StandardBeanExpressionResolver, which parses SpEL via TemplateAwareExpressionParser — a real chain Spring executes when a transaction attribute contains an expression.


### CVE-2023-20883

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.security.web.FilterChainProxy#afterPropertiesSet()V
 → org.springframework.security.config.http.DefaultFilterChainValidator#validate(Lorg/springframework/security/web/FilterChainProxy;)V
 → org.springframework.security.config.http.DefaultFilterChainValidator#checkLoginPageIsntProtected(Lorg/springframework/security/web/FilterChainProxy;Ljava/util/List;)V
 → org.springframework.security.web.access.intercept.DefaultFilterInvocationSecurityMetadataSource#getAttributes(Ljava/lang/Object;)Ljava/util/Collection;
 → org.springframework.security.web.servlet.util.matcher.MvcRequestMatcher#matches(Ljavax/servlet/http/HttpServletRequest;)Z
 → org.springframework.security.web.servlet.util.matcher.MvcRequestMatcher#getMapping(Ljavax/servlet/http/HttpServletRequest;)Lorg/springframework/web/servlet/handler/MatchableHandlerMapping;
 → org.springframework.web.servlet.handler.HandlerMappingIntrospector#getMatchableHandlerMapping(Ljavax/servlet/http/HttpServletRequest;)Lorg/springframework/web/servlet/handler/MatchableHandlerMapping;
 → org.springframework.web.servlet.handler.AbstractHandlerMapping#getHandler(Ljavax/servlet/http/HttpServletRequest;)Lorg/springframework/web/servlet/HandlerExecutionChain;
 → org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#getHandlerInternal(Ljavax/servlet/http/HttpServletRequest;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The final hop relies on HandlerMappingIntrospector#getMatchableHandlerMapping resolving to WelcomePageHandlerMapping, but at runtime the introspector resolves the specific handler-mapping bean recorded for the request (normally RequestMappingHandlerMapping), so AbstractHandlerMapping#getHandler would not dispatch into WelcomePageHandlerMapping#getHandlerInternal here. The preceding hops (DefaultFilterChainValidator → getAttributes → MvcRequestMatcher) are also generic virtual/interface dispatches that CHA over-approximates across unrelated security-metadata-source implementations.


### CVE-2023-34055

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#applyBeanPostProcessorsAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcessAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcess(Lorg/springframework/security/web/server/WebFilterChainProxy;)Lorg/springframework/security/web/server/WebFilterChainProxy;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#lambda$postProcess$0(Lorg/springframework/web/server/ServerWebExchange;Lorg/springframework/web/server/WebFilterChain;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#filter(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#lambda$filter$0(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#invokeFilter(Lorg/springframework/web/server/WebFilter;Lorg/springframework/web/server/handler/DefaultWebFilterChain;Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.boot.actuate.metrics.web.reactive.server.MetricsWebFilter#filter(Lorg/springframework/web/server/ServerWebExchange;Lreactor/core/publisher/Mono;)Lorg/reactivestreams/Publisher;
 → org.springframework.boot.actuate.metrics.web.reactive.server.MetricsWebFilter#lambda$filter$1(Lorg/springframework/web/server/ServerWebExchange;JLjava/lang/Void;)V
 → org.springframework.boot.actuate.metrics.web.reactive.server.MetricsWebFilter#onSuccess(Lorg/springframework/web/server/ServerWebExchange;J)V
 → org.springframework.boot.actuate.metrics.web.reactive.server.MetricsWebFilter#record(Lorg/springframework/web/server/ServerWebExchange;JLjava/lang/Throwable;)V
 → org.springframework.boot.actuate.metrics.web.reactive.server.DefaultWebFluxTagsProvider#httpRequestTags(Lorg/springframework/web/server/ServerWebExchange;Ljava/lang/Throwable;)Ljava/lang/Iterable;
 → org.springframework.boot.actuate.metrics.web.reactive.server.WebFluxTags#method(Lorg/springframework/web/server/ServerWebExchange;)Lio/micrometer/core/instrument/Tag;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The tail is a genuine, coherent Spring Boot actuator chain: MetricsWebFilter#filter → onSuccess → record → DefaultWebFluxTagsProvider#httpRequestTags → WebFluxTags#method is exactly how the framework records HTTP metrics, and the WebFilterChainPostProcessor really does post-process WebFilterChainProxy during bean initialization. The head is generic bean-factory wiring (BeanPostProcessor during initializeBean), which is a real Spring mechanism, though whether the reactive (WebFlux) filter chain actually runs in this servlet-based WebGoat app would need runtime confirmation.


### CVE-2023-4639

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.JvmRouteHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → io.undertow.server.HttpServerExchange#requestCookies()Ljava/lang/Iterable;
 → io.undertow.util.Cookies#parseRequestCookies(IZLjava/util/List;Ljava/util/Set;)V
 → io.undertow.util.Cookies#parseRequestCookies(IZLjava/util/List;Ljava/util/Set;Z)V
 → io.undertow.util.Cookies#parseRequestCookies(IZLjava/util/List;Ljava/util/Set;ZZ)V
 → io.undertow.util.Cookies#parseCookie(Ljava/lang/String;Ljava/util/Set;IZZZ)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows real Spring Boot startup wiring (AuthenticationConfiguration lazy bean -> bean initialization -> DelegatingFilterProxy init -> context refresh -> UndertowServletWebServerFactory createWebServer -> DeploymentManagerImpl deploy), and the tail is genuine Undertow request handling: ServletInitialHandler -> HttpServerExchange#requestCookies -> Cookies#parseRequestCookies, which is exactly where cookie parsing occurs. These hops correspond to actual framework behavior, not CHA-invented edges.


### CVE-2023-5685

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.client.http2.Http2ClearClientProvider$1#handleEvent(Lorg/xnio/StreamConnection;)V
 → org.xnio.FinishedIoFuture#addNotifier(Lorg/xnio/IoFuture$Notifier;Ljava/lang/Object;)Lorg/xnio/IoFuture;
 → org.xnio.IoFuture$HandlingNotifier#notify(Lorg/xnio/IoFuture;Ljava/lang/Object;)V
 → io.undertow.protocols.ssl.UndertowXnioSsl$3#handleCancelled(Lorg/xnio/IoFuture;)V
 → org.xnio.AbstractIoFuture#cancel()Lorg/xnio/IoFuture;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from DeploymentManagerImpl#deploy() (startup-time deployment) to ServletInitialHandler$2#call (per-request handling) is not a real call edge — deployment never invokes the request handler directly. The tail (flush() -> ChannelListeners -> FinishedIoFuture#addNotifier -> HandlingNotifier#notify -> UndertowXnioSsl$3#handleCancelled -> AbstractIoFuture#cancel) is likewise a CHA-stitched chain of unrelated generic dispatch (IoFuture$Notifier.notify reaches every HandlingNotifier implementation), not a coherent runtime sequence.


### CVE-2023-6378

```
org.owasp.webgoat.AsciiDoctorTemplateResolver#computeTemplateResource(Lorg/thymeleaf/IEngineConfiguration;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/util/Map;)Lorg/thymeleaf/templateresource/ITemplateResource;
 → ch.qos.logback.classic.Logger#warn(Ljava/lang/String;)V
 → ch.qos.logback.classic.Logger#filterAndLog_0_Or3Plus(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#buildLoggingEventAndAppend(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#callAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)V
 → ch.qos.logback.classic.Logger#appendLoopOnAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)I
 → ch.qos.logback.core.spi.AppenderAttachableImpl#appendLoopOnAppenders(Ljava/lang/Object;)I
 → ch.qos.logback.core.UnsynchronizedAppenderBase#doAppend(Ljava/lang/Object;)V
 → ch.qos.logback.core.OutputStreamAppender#append(Ljava/lang/Object;)V
 → ch.qos.logback.core.rolling.RollingFileAppender#subAppend(Ljava/lang/Object;)V
 → ch.qos.logback.core.rolling.RollingFileAppender#rollover()V
 → ch.qos.logback.core.rolling.RollingFileAppender#attemptRollover()V
 → ch.qos.logback.core.rolling.FixedWindowRollingPolicy#rollover()V
 → ch.qos.logback.core.rolling.helper.RenameUtil#rename(Ljava/lang/String;Ljava/lang/String;)V
 → ch.qos.logback.core.rolling.helper.RenameUtil#areOnDifferentVolumes(Ljava/io/File;Ljava/io/File;)Ljava/lang/Boolean;
 → ch.qos.logback.core.util.EnvUtil#isJDK7OrHigher()Z
 → ch.qos.logback.core.util.EnvUtil#isJDK_N_OrHigher(I)Z
```

**Entry call (app → dependency):** `org.owasp.webgoat.AsciiDoctorTemplateResolver#computeTemplateResource(Lorg/thymeleaf/IEngineConfiguration;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/util/Map;)Lorg/thymeleaf/templateresource/ITemplateResource;` → `ch.qos.logback.classic.Logger#warn(Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain is a genuine logback flow: a Logger#warn from application code goes through filterAndLog/callAppenders to a RollingFileAppender, whose rollover path really calls RenameUtil#areOnDifferentVolumes and EnvUtil#isJDK7OrHigher/isJDK_N_OrHigher. Note the path only reaches the EnvUtil helpers, not the actual vulnerable receiver deserialization (LoggingEventVO#readObject), so exploitability of CVE-2023-6378 via this path is doubtful even though the call sequence itself is real.


### CVE-2024-12798

```
org.owasp.webgoat.AsciiDoctorTemplateResolver#computeTemplateResource(Lorg/thymeleaf/IEngineConfiguration;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/util/Map;)Lorg/thymeleaf/templateresource/ITemplateResource;
 → ch.qos.logback.classic.Logger#warn(Ljava/lang/String;)V
 → ch.qos.logback.classic.Logger#filterAndLog_0_Or3Plus(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#buildLoggingEventAndAppend(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#callAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)V
 → ch.qos.logback.classic.Logger#appendLoopOnAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)I
 → ch.qos.logback.core.spi.AppenderAttachableImpl#appendLoopOnAppenders(Ljava/lang/Object;)I
 → ch.qos.logback.core.AppenderBase#doAppend(Ljava/lang/Object;)V
 → ch.qos.logback.core.net.SMTPAppenderBase#append(Ljava/lang/Object;)V
 → ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#evaluate(Ljava/lang/Object;)Z
```

**Entry call (app → dependency):** `org.owasp.webgoat.AsciiDoctorTemplateResolver#computeTemplateResource(Lorg/thymeleaf/IEngineConfiguration;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/util/Map;)Lorg/thymeleaf/templateresource/ITemplateResource;` → `ch.qos.logback.classic.Logger#warn(Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — The path is the genuine logback event pipeline: Logger.warn → filterAndLog → callAppenders → AppenderAttachableImpl.appendLoopOnAppenders → AppenderBase.doAppend dispatching to the concrete appender's append, and SMTPAppenderBase really does invoke its configured EventEvaluator (typically a JaninoEventEvaluator) on each event to decide whether to buffer/send mail. The only configuration assumption is that an SMTPAppender with a Janino evaluator is attached, which is a standard logback setup, not a CHA artifact.


### CVE-2024-1459

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.server.protocol.http.HttpOpenListener#handleEvent(Lorg/xnio/StreamConnection;Lio/undertow/connector/PooledByteBuffer;)V
 → io.undertow.server.protocol.http.HttpReadListener#newRequest()V
 → io.undertow.server.protocol.http.ParseState#reset()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain jumps from DeploymentManagerImpl#deploy() into ServletInitialHandler$2#call (a runtime request handler, not part of deployment), then hops handleFirstRequest → ServerSentEventHandler#handleRequest → SplitStreamSinkChannel#flush → ChannelListeners#invokeChannelListener → HttpOpenListener#handleEvent, which reverses the actual data flow (an open listener would lead to request parsing, not be reached from a channel flush during SSE handling). These edges connect unrelated Undertow components purely via generic interface/virtual dispatch, so the path to ParseState#reset is a CHA artifact.


### CVE-2024-1635

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.conduits.ConduitStreamSinkChannel#resumeWrites()V
 → io.undertow.conduits.WriteTimeoutStreamSinkConduit#resumeWrites()V
 → io.undertow.conduits.WriteTimeoutStreamSinkConduit#handleResumeTimeout()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — While the Spring Security → embedded Undertow server creation portion (nodes 1-17) is plausible framework setup, the hop from ServletInitialHandler#handleFirstRequest to ServerSentEventHandler#handleRequest (node 21) is a classic CHA artifact — the SSE handler is a separate HttpHandler added by explicit handler chain configuration, not something the servlet initial handler dispatches to. The subsequent ConduitStreamSinkChannel#resumeWrites edge to WriteTimeoutStreamSinkConduit is likewise a generic conduit-interface dispatch that only occurs if that conduit was actually installed in the pipeline, which the path does not establish.


### CVE-2024-22257

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.security.web.FilterChainProxy#afterPropertiesSet()V
 → org.springframework.security.config.http.DefaultFilterChainValidator#validate(Lorg/springframework/security/web/FilterChainProxy;)V
 → org.springframework.security.config.http.DefaultFilterChainValidator#checkLoginPageIsntProtected(Lorg/springframework/security/web/FilterChainProxy;Ljava/util/List;)V
 → org.springframework.security.access.vote.AffirmativeBased#decide(Lorg/springframework/security/core/Authentication;Ljava/lang/Object;Ljava/util/Collection;)V
 → org.springframework.security.access.vote.AuthenticatedVoter#vote(Lorg/springframework/security/core/Authentication;Ljava/lang/Object;Ljava/util/Collection;)I
 → org.springframework.security.access.vote.AuthenticatedVoter#isFullyAuthenticated(Lorg/springframework/security/core/Authentication;)Z
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is a genuine Spring Security flow: FilterChainProxy#afterPropertiesSet really invokes DefaultFilterChainValidator#validate, whose checkLoginPageIsntProtected really calls the configured AccessDecisionManager (AffirmativeBased#decide) with a null/anonymous Authentication, which dispatches to AuthenticatedVoter#vote and isFullyAuthenticated. The only questionable hop is lazyBean -> AutowireBeanFactoryObjectPostProcessor#postProcess -> initializeBean, but that matches Spring's actual lazy-init bean post-processing wiring, so the chain corresponds to real framework behavior.


### CVE-2024-22259

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#applyBeanPostProcessorsAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcessAfterInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcess(Lorg/springframework/security/web/server/WebFilterChainProxy;)Lorg/springframework/security/web/server/WebFilterChainProxy;
 → org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#lambda$postProcess$0(Lorg/springframework/web/server/ServerWebExchange;Lorg/springframework/web/server/WebFilterChain;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#filter(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#lambda$filter$0(Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.server.handler.DefaultWebFilterChain#invokeFilter(Lorg/springframework/web/server/WebFilter;Lorg/springframework/web/server/handler/DefaultWebFilterChain;Lorg/springframework/web/server/ServerWebExchange;)Lreactor/core/publisher/Mono;
 → org.springframework.web.cors.reactive.CorsWebFilter#filter(Lorg/springframework/web/server/ServerWebExchange;Lorg/springframework/web/server/WebFilterChain;)Lreactor/core/publisher/Mono;
 → org.springframework.web.cors.reactive.DefaultCorsProcessor#process(Lorg/springframework/web/cors/CorsConfiguration;Lorg/springframework/web/server/ServerWebExchange;)Z
 → org.springframework.web.cors.reactive.CorsUtils#isCorsRequest(Lorg/springframework/http/server/reactive/ServerHttpRequest;)Z
 → org.springframework.web.cors.reactive.CorsUtils#isSameOrigin(Lorg/springframework/http/server/reactive/ServerHttpRequest;)Z
 → org.springframework.web.util.UriComponentsBuilder#fromOriginHeader(Ljava/lang/String;)Lorg/springframework/web/util/UriComponentsBuilder;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — WebGoat is a servlet (non-reactive) Spring app, yet the path routes through ReactiveCloudFoundryActuatorAutoConfiguration's WebFilterChainPostProcessor, which would only run in a reactive Cloud Foundry deployment; the hop from applyBeanPostProcessorsAfterInitialization to that specific post-processor is a classic CHA over-approximation over BeanPostProcessor#postProcessAfterInitialization. While CorsUtils#isSameOrigin genuinely calls UriComponentsBuilder#fromOriginHeader, the connecting chain is not a real runtime path for this application.


### CVE-2024-38808

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.transaction.interceptor.NameMatchTransactionAttributeSource#afterPropertiesSet()V
 → org.springframework.transaction.interceptor.DefaultTransactionAttribute#resolveAttributeStrings(Lorg/springframework/util/StringValueResolver;)V
 → org.springframework.beans.factory.config.EmbeddedValueResolver#resolveStringValue(Ljava/lang/String;)Ljava/lang/String;
 → org.springframework.context.expression.StandardBeanExpressionResolver#evaluate(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanExpressionContext;)Ljava/lang/Object;
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.Assign#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain is realistic Spring wiring: WebSecurityConfigurerAdapter#authenticationManager delegates to AuthenticationConfiguration, whose lazyBean/ObjectPostProcessor path runs initializeBean/invokeInitMethods on beans; a NameMatchTransactionAttributeSource (an InitializingBean used by WebGoat's transactional proxies) then resolves its attribute strings through EmbeddedValueResolver -> StandardBeanExpressionResolver -> SpelExpression#getValue, which is exactly the SpEL evaluation the CVE concerns. The final SpelNodeImpl#getValue -> Assign#getValueInternal hop is CHA-broad (any AST node), but reaching SpelExpression#getValue on user-influenced bean definitions is the substantive vulnerable sink, so the overall path corresponds to real framework behavior.


### CVE-2024-38827

```
org.owasp.webgoat.asciidoc.WebGoatTmpDirMacro#process(Lorg/asciidoctor/ast/ContentNode;Ljava/lang/String;Ljava/util/Map;)Ljava/lang/Object;
 → org.springframework.core.env.AbstractEnvironment#getProperty(Ljava/lang/String;)Ljava/lang/String;
 → org.springframework.core.env.PropertySourcesPropertyResolver#getProperty(Ljava/lang/String;Ljava/lang/Class;Z)Ljava/lang/Object;
 → org.springframework.core.env.SystemEnvironmentPropertySource#getProperty(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.core.env.SystemEnvironmentPropertySource#resolvePropertyName(Ljava/lang/String;)Ljava/lang/String;
```

**Entry call (app → dependency):** `org.owasp.webgoat.asciidoc.WebGoatTmpDirMacro#process(Lorg/asciidoctor/ast/ContentNode;Ljava/lang/String;Ljava/util/Map;)Ljava/lang/Object;` → `org.springframework.core.env.AbstractEnvironment#getProperty(Ljava/lang/String;)Ljava/lang/String;`


**Path plausibility (z-ai/glm-5.3):** plausible — WebGoatTmpDirMacro#process reading an environment property via AbstractEnvironment#getProperty is a natural thing for a macro expanding a temp-dir path to do, and the resolver chain (PropertySourcesPropertyResolver -> SystemEnvironmentPropertySource#resolvePropertyName) is exactly how Spring resolves system environment properties, so the vulnerable method is genuinely on this dispatch path.


### CVE-2024-3884

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.core.DeploymentManagerImpl#handleExtensions(Lio/undertow/servlet/api/DeploymentInfo;Lio/undertow/servlet/spec/ServletContextImpl;)V
 → io.undertow.websockets.jsr.Bootstrap#handleDeployment(Lio/undertow/servlet/api/DeploymentInfo;Ljavax/servlet/ServletContext;)V
 → io.undertow.websockets.jsr.UndertowContainerProvider#getDefaultContainer()Lio/undertow/websockets/jsr/ServerWebSocketContainer;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows the real Spring Boot embedded-Undertow startup flow: ServletWebServerApplicationContext#createWebServer -> UndertowServletWebServerFactory -> DeploymentManagerImpl#deploy -> Bootstrap#handleDeployment -> UndertowContainerProvider#getDefaultContainer, which is exactly how Undertow's JSR-356 websocket extension initializes during servlet deployment. The earlier Spring Security bean-initialization hops are also genuine framework wiring, so this looks like a real setup-time call chain rather than a CHA artifact.


### CVE-2024-4027

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.core.DeploymentManagerImpl#handleExtensions(Lio/undertow/servlet/api/DeploymentInfo;Lio/undertow/servlet/spec/ServletContextImpl;)V
 → io.undertow.websockets.jsr.Bootstrap#handleDeployment(Lio/undertow/servlet/api/DeploymentInfo;Ljavax/servlet/ServletContext;)V
 → io.undertow.websockets.jsr.UndertowContainerProvider#getDefaultContainer()Lio/undertow/websockets/jsr/ServerWebSocketContainer;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is genuine Spring Boot embedded-Undertow startup wiring: UndertowServletWebServerFactory creates a DeploymentManager, deploy() runs deployment extensions, and the JSR-356 websocket Bootstrap.handleDeployment really does call UndertowContainerProvider.getDefaultContainer() to set up the ServerWebSocketContainer. Every hop matches real framework initialization code rather than a CHA-style generic dispatch.


### CVE-2024-6162

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.server.protocol.ajp.AjpReadListener#handleEvent(Lorg/xnio/channels/StreamSourceChannel;)V
 → io.undertow.server.protocol.ajp.AjpRequestParser#parse(Ljava/nio/ByteBuffer;Lio/undertow/server/protocol/ajp/AjpRequestParseState;Lio/undertow/server/HttpServerExchange;)V
 → io.undertow.server.protocol.ajp.AjpRequestParser#decode(Ljava/lang/String;Z)Ljava/lang/String;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — While the Spring Boot startup chain into Undertow's DeploymentManagerImpl#deploy is plausible, the tail is not: ServletInitialHandler#handleFirstRequest would not dispatch to ServerSentEventHandler#handleRequest, and a SplitStreamSinkChannel#flush() listener invocation resolving to AjpReadListener#handleEvent is a classic CHA artifact connecting unrelated channel listeners via a generic ChannelListener interface.


### CVE-2024-7885

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.filter.GenericFilterBean#afterPropertiesSet()V
 → org.springframework.web.filter.DelegatingFilterProxy#initFilterBean()V
 → org.springframework.web.filter.DelegatingFilterProxy#findWebApplicationContext()Lorg/springframework/web/context/WebApplicationContext;
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.undertow.UndertowServletWebServerFactory#createManager([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lio/undertow/servlet/api/DeploymentManager;
 → io.undertow.servlet.core.DeploymentManagerImpl#deploy()V
 → io.undertow.servlet.handlers.ServletInitialHandler$2#call(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)Ljava/lang/Object;
 → io.undertow.servlet.handlers.ServletInitialHandler#access$100(Lio/undertow/servlet/handlers/ServletInitialHandler;Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.servlet.handlers.ServletInitialHandler#handleFirstRequest(Lio/undertow/server/HttpServerExchange;Lio/undertow/servlet/handlers/ServletRequestContext;)V
 → io.undertow.server.handlers.sse.ServerSentEventHandler#handleRequest(Lio/undertow/server/HttpServerExchange;)V
 → org.xnio.channels.SplitStreamSinkChannel#flush()Z
 → org.xnio.ChannelListeners#invokeChannelListener(Ljava/nio/channels/Channel;Lorg/xnio/ChannelListener;)Z
 → io.undertow.server.protocol.proxy.ProxyProtocolReadListener#handleEvent(Lorg/xnio/channels/StreamSourceChannel;)V
 → io.undertow.server.protocol.proxy.ProxyProtocolReadListener#parseProxyProtocolV1(Lio/undertow/connector/PooledByteBuffer;Ljava/util/concurrent/atomic/AtomicBoolean;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot setup portion (WebSecurityConfig -> ... -> UndertowServletWebServerFactory -> DeploymentManagerImpl#deploy) is a plausible embedded-Undertow startup chain, but the tail is a CHA artifact: ServerSentEventHandler.handleFirstRequest -> SplitStreamSinkChannel#flush -> ChannelListeners#invokeChannelListener hops through the generic ChannelListener interface, and ProxyProtocolReadListener is an accept-time proxy-protocol parser, not a listener that SSE flush would ever invoke.


### CVE-2025-41249

```
org.owasp.webgoat.users.UserRepository#findAll()Ljava/util/List;
 → org.springframework.data.jpa.repository.support.SimpleJpaRepository#findAll(Lorg/springframework/data/domain/Example;)Ljava/util/List;
 → org.springframework.data.jpa.repository.support.SimpleJpaRepository#getQuery(Lorg/springframework/data/jpa/domain/Specification;Ljava/lang/Class;Lorg/springframework/data/domain/Sort;)Ljavax/persistence/TypedQuery;
 → org.springframework.data.jpa.repository.query.QueryUtils#toOrders(Lorg/springframework/data/domain/Sort;Ljavax/persistence/criteria/From;Ljavax/persistence/criteria/CriteriaBuilder;)Ljava/util/List;
 → org.springframework.data.jpa.repository.query.QueryUtils#toJpaOrder(Lorg/springframework/data/domain/Sort$Order;Ljavax/persistence/criteria/From;Ljavax/persistence/criteria/CriteriaBuilder;)Ljavax/persistence/criteria/Order;
 → org.springframework.data.jpa.repository.query.QueryUtils#toExpressionRecursively(Ljavax/persistence/criteria/From;Lorg/springframework/data/mapping/PropertyPath;)Ljavax/persistence/criteria/Expression;
 → org.springframework.data.jpa.repository.query.QueryUtils#toExpressionRecursively(Ljavax/persistence/criteria/From;Lorg/springframework/data/mapping/PropertyPath;Z)Ljavax/persistence/criteria/Expression;
 → org.springframework.data.jpa.repository.query.QueryUtils#requiresOuterJoin(Ljavax/persistence/metamodel/Bindable;ZZZ)Z
 → org.springframework.data.jpa.repository.query.QueryUtils#getAnnotationProperty(Ljavax/persistence/metamodel/Attribute;Ljava/lang/String;Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationUtils#getAnnotation(Ljava/lang/reflect/AnnotatedElement;Ljava/lang/Class;)Ljava/lang/annotation/Annotation;
 → org.springframework.core.annotation.TypeMappedAnnotations#get(Ljava/lang/Class;Ljava/util/function/Predicate;Lorg/springframework/core/annotation/MergedAnnotationSelector;)Lorg/springframework/core/annotation/MergedAnnotation;
 → org.springframework.core.annotation.TypeMappedAnnotations#scan(Ljava/lang/Object;Lorg/springframework/core/annotation/AnnotationsProcessor;)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationsScanner#scan(Ljava/lang/Object;Ljava/lang/reflect/AnnotatedElement;Lorg/springframework/core/annotation/MergedAnnotations$SearchStrategy;Lorg/springframework/core/annotation/AnnotationsProcessor;)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationsScanner#process(Ljava/lang/Object;Ljava/lang/reflect/AnnotatedElement;Lorg/springframework/core/annotation/MergedAnnotations$SearchStrategy;Lorg/springframework/core/annotation/AnnotationsProcessor;)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationsScanner#processMethod(Ljava/lang/Object;Ljava/lang/reflect/Method;Lorg/springframework/core/annotation/MergedAnnotations$SearchStrategy;Lorg/springframework/core/annotation/AnnotationsProcessor;)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationsScanner#processMethodHierarchy(Ljava/lang/Object;[ILjava/lang/Class;Lorg/springframework/core/annotation/AnnotationsProcessor;Ljava/lang/reflect/Method;Z)Ljava/lang/Object;
 → org.springframework.core.annotation.AnnotationsScanner#isOverride(Ljava/lang/reflect/Method;Ljava/lang/reflect/Method;)Z
 → org.springframework.core.annotation.AnnotationsScanner#hasSameParameterTypes(Ljava/lang/reflect/Method;Ljava/lang/reflect/Method;)Z
 → org.springframework.core.annotation.AnnotationsScanner#hasSameGenericTypeParameters(Ljava/lang/reflect/Method;Ljava/lang/reflect/Method;[Ljava/lang/Class;)Z
```

**Entry call (app → dependency):** `org.owasp.webgoat.users.UserRepository#findAll()Ljava/util/List;` → `org.springframework.data.jpa.repository.support.SimpleJpaRepository#findAll(Lorg/springframework/data/domain/Example;)Ljava/util/List;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain from SimpleJpaRepository through QueryUtils.toOrders/toExpressionRecursively into AnnotationUtils/AnnotationsScanner is a genuine Spring Data JPA query-building flow that really does annotation scanning on entity properties, so the vulnerable method is plausibly reached. The first hop (findAll() vs findAll(Example)) is a signature-mismatched CHA edge, but the framework-internal portion of the path corresponds to real Spring behavior.


### CVE-2025-49128

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#createParser(Ljava/io/InputStream;)Lcom/fasterxml/jackson/core/JsonParser;
 → com.fasterxml.jackson.core.JsonFactory#createParser(Ljava/io/InputStream;)Lcom/fasterxml/jackson/core/JsonParser;
 → com.fasterxml.jackson.core.JsonFactory#_createParser(Ljava/io/InputStream;Lcom/fasterxml/jackson/core/io/IOContext;)Lcom/fasterxml/jackson/core/JsonParser;
 → com.fasterxml.jackson.core.json.ByteSourceJsonBootstrapper#constructParser(ILcom/fasterxml/jackson/core/ObjectCodec;Lcom/fasterxml/jackson/core/sym/ByteQuadsCanonicalizer;Lcom/fasterxml/jackson/core/sym/CharsToNameCanonicalizer;I)Lcom/fasterxml/jackson/core/JsonParser;
 → com.fasterxml.jackson.core.json.ReaderBasedJsonParser#<init>(Lcom/fasterxml/jackson/core/io/IOContext;ILjava/io/Reader;Lcom/fasterxml/jackson/core/ObjectCodec;Lcom/fasterxml/jackson/core/sym/CharsToNameCanonicalizer;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — Each hop corresponds to real framework behavior: Spring Security's authenticationManager() really delegates through AuthenticationConfiguration/lazyBean to bean initialization, and InitializingBean.afterPropertiesSet on DataSourceInitializerInvoker publishing a ContextRefreshedEvent that Spring Data's repository populator handles by reading JSON via Jackson2ResourceReader/ObjectReader is a genuine, documented chain. The final hops (JsonFactory.createParser -> ByteSourceJsonBootstrapper.constructParser -> ReaderBasedJsonParser.<init>) are the actual Jackson parser-construction path named in the advisory, so the chain plausibly reaches the vulnerable code.


### CVE-2026-19032

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ext.NioPathDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/nio/file/Path;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The final hop is implausible: ObjectReader#readTree/_bindAsTree binds to a generic JsonNode tree and never invokes a typed deserializer such as NioPathDeserializer, which is only selected for a declared java.nio.file.Path target type. The earlier Spring wiring (event multicaster → repository populator → Jackson2ResourceReader) is plausible, but the JSON read is readTree, so the vulnerable Path deserialization cannot be reached this way.


### CVE-2026-41851

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.data.repository.core.support.RepositoryFactoryBeanSupport#afterPropertiesSet()V
 → org.springframework.data.repository.core.support.RepositoryFactoryBeanSupport#lambda$afterPropertiesSet$5(Lorg/springframework/data/repository/core/support/RepositoryComposition$RepositoryFragments;)Lorg/springframework/data/repository/Repository;
 → org.springframework.data.repository.core.support.RepositoryFactorySupport#getRepository(Ljava/lang/Class;Lorg/springframework/data/repository/core/support/RepositoryComposition$RepositoryFragments;)Ljava/lang/Object;
 → org.springframework.data.repository.core.support.RepositoryFactorySupport#getProjectionFactory(Ljava/lang/ClassLoader;Lorg/springframework/beans/factory/BeanFactory;)Lorg/springframework/data/projection/ProjectionFactory;
 → org.springframework.data.projection.SpelAwareProxyProjectionFactory#<init>()V
 → org.springframework.expression.spel.standard.SpelExpressionParser#<init>()V
 → org.springframework.expression.spel.SpelParserConfiguration#<init>()V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Data's RepositoryFactorySupport really does build a SpelAwareProxyProjectionFactory during repository bean initialization, and that class genuinely instantiates a SpelExpressionParser (which constructs a SpelParserConfiguration) — this is standard Spring Data projection wiring, not a CHA artifact.


### CVE-2026-47884

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#applyBeanPostProcessorsBeforeInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.context.support.ApplicationContextAwareProcessor#postProcessBeforeInitialization(Ljava/lang/Object;Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.context.support.ApplicationContextAwareProcessor#invokeAwareInterfaces(Ljava/lang/Object;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.xslt.XsltView#initApplicationContext()V
 → org.springframework.web.servlet.view.xslt.XsltView#loadTemplates()Ljavax/xml/transform/Templates;
 → org.springframework.web.servlet.view.xslt.XsltView#getStylesheetSource()Ljavax/xml/transform/Source;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — Steps 5-11 are genuine Spring bean-lifecycle wiring, but the hop 11→12 exists only because XsltView extends ApplicationObjectSupport and thus implements setApplicationContext, which ApplicationContextAwareProcessor invokes on ANY such bean via interface dispatch; nothing indicates WebGoat ever defines or instantiates an XsltView bean, so the connection to the vulnerable view is a CHA over-approximation.


### CVE-2026-49844

```
org.owasp.webgoat.i18n.PluginMessages#refreshProperties(Ljava/lang/String;Lorg/springframework/context/support/ReloadableResourceBundleMessageSource$PropertiesHolder;)Lorg/springframework/context/support/ReloadableResourceBundleMessageSource$PropertiesHolder;
 → org.apache.commons.logging.LogAdapter$Log4jLog#error(Ljava/lang/Object;)V
 → org.apache.commons.logging.LogAdapter$Log4jLog#log(Lorg/apache/logging/log4j/Level;Ljava/lang/Object;Ljava/lang/Throwable;)V
 → org.apache.logging.log4j.spi.AbstractLogger#logIfEnabled(Ljava/lang/String;Lorg/apache/logging/log4j/Level;Lorg/apache/logging/log4j/Marker;Lorg/apache/logging/log4j/util/MessageSupplier;Ljava/lang/Throwable;)V
 → org.apache.logging.log4j.simple.SimpleLogger#logMessage(Ljava/lang/String;Lorg/apache/logging/log4j/Level;Lorg/apache/logging/log4j/Marker;Lorg/apache/logging/log4j/message/Message;Ljava/lang/Throwable;)V
 → org.apache.logging.log4j.message.MapMessage#getFormattedMessage([Ljava/lang/String;)Ljava/lang/String;
 → org.apache.logging.log4j.message.MapMessage#format(Lorg/apache/logging/log4j/message/MapMessage$MapFormat;Ljava/lang/StringBuilder;)Ljava/lang/StringBuilder;
 → org.apache.logging.log4j.message.MapMessage#asJson(Ljava/lang/StringBuilder;)V
 → org.apache.logging.log4j.message.MapMessageJsonFormatter#format(Ljava/lang/StringBuilder;Ljava/lang/Object;)V
 → org.apache.logging.log4j.message.MapMessageJsonFormatter#format(Ljava/lang/StringBuilder;Ljava/lang/Object;I)V
 → org.apache.logging.log4j.message.MapMessageJsonFormatter#formatDoubleArray(Ljava/lang/StringBuilder;[D)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.i18n.PluginMessages#refreshProperties(Ljava/lang/String;Lorg/springframework/context/support/ReloadableResourceBundleMessageSource$PropertiesHolder;)Lorg/springframework/context/support/ReloadableResourceBundleMessageSource$PropertiesHolder;` → `org.apache.commons.logging.LogAdapter$Log4jLog#error(Ljava/lang/Object;)V`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop SimpleLogger#logMessage → MapMessage#getFormattedMessage is a CHA artifact: an error(Object) call wraps the argument as a Simple/ObjectMessage, and SimpleLogger would never invoke MapMessage's JSON-format getFormattedMessage([String) — that path is specific to JsonTemplateLayout. Additionally, SimpleLogger is only used when no real Log4j backend is configured, making the whole chain implausible.


### CVE-2026-54513

```
org.owasp.webgoat.WebGoat#restTemplate()Lorg/springframework/web/client/RestTemplate;
 → org.springframework.web.client.RestTemplate#<init>()V
 → org.springframework.http.converter.xml.MappingJackson2XmlHttpMessageConverter#<init>()V
 → org.springframework.http.converter.json.Jackson2ObjectMapperBuilder#build()Lcom/fasterxml/jackson/databind/ObjectMapper;
 → org.springframework.http.converter.json.Jackson2ObjectMapperBuilder#configure(Lcom/fasterxml/jackson/databind/ObjectMapper;)V
 → com.fasterxml.jackson.databind.ObjectMapper#registerModule(Lcom/fasterxml/jackson/databind/Module;)Lcom/fasterxml/jackson/databind/ObjectMapper;
 → org.springframework.security.jackson2.CoreJackson2Module#setupModule(Lcom/fasterxml/jackson/databind/Module$SetupContext;)V
 → org.springframework.security.jackson2.SecurityJackson2Modules#enableDefaultTyping(Lcom/fasterxml/jackson/databind/ObjectMapper;)V
 → org.springframework.security.jackson2.SecurityJackson2Modules#createAllowlistedDefaultTyping()Lcom/fasterxml/jackson/databind/jsontype/TypeResolverBuilder;
 → org.springframework.security.jackson2.SecurityJackson2Modules$AllowlistTypeResolverBuilder#<init>(Lcom/fasterxml/jackson/databind/ObjectMapper$DefaultTyping;)V
 → com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator$Builder#build()Lcom/fasterxml/jackson/databind/jsontype/BasicPolymorphicTypeValidator;
 → com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#<init>(Ljava/util/Set;[Lcom/fasterxml/jackson/databind/jsontype/BasicPolymorphicTypeValidator$TypeMatcher;[Lcom/fasterxml/jackson/databind/jsontype/BasicPolymorphicTypeValidator$NameMatcher;[Lcom/fasterxml/jackson/databind/jsontype/BasicPolymorphicTypeValidator$TypeMatcher;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebGoat#restTemplate()Lorg/springframework/web/client/RestTemplate;` → `org.springframework.web.client.RestTemplate#<init>()V`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Security's SecurityJackson2Modules really does construct a BasicPolymorphicTypeValidator (via createAllowlistedDefaultTyping/AllowlistTypeResolverBuilder) that uses allowIfBaseType/allowIfSubTypeIsArray, and Jackson2ObjectMapperBuilder.build() really registers CoreJackson2Module, so this is a genuine framework wiring path to the vulnerable constructor.


### CVE-2026-77310

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#_deserialize(Ljava/lang/String;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from AbstractAutowireCapableBeanFactory#invokeInitMethods to DataSourceInitializerInvoker#afterPropertiesSet is a classic CHA over-approximation (any InitializingBean could appear there), and the subsequent chain through event multicasting to a repository populator is unrelated wiring. Moreover, Jackson2ResourceReader uses readTree into a JsonNode, which would not invoke FromStringDeserializer for InetAddress at all — the vulnerable branch requires an InetAddress-typed target, so this path does not correspond to a real call chain.


### CVE-2026-89425

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#nextToken()Lcom/fasterxml/jackson/core/JsonToken;
 → com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_handleUnexpectedValue(I)Lcom/fasterxml/jackson/core/JsonToken;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from AbstractAutowireCapableBeanFactory#invokeInitMethods to DataSourceInitializerInvoker#afterPropertiesSet is a classic CHA artifact — initializing the authenticationManager bean would not run afterPropertiesSet on an unrelated DataSourceInitializerInvoker bean. Additionally, Jackson2ResourceReader reads from an InputStream (ObjectReader#readTree(InputStream)), which produces UTF8StreamJsonParser, not the DataInput-backed UTF8DataInputJsonParser that is the sole vulnerable implementation.


### CVE-2026-91776

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.impl.TypeWrappedDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.JsonDeserializer#deserializeWithType(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;Lcom/fasterxml/jackson/databind/jsontype/TypeDeserializer;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.jsontype.impl.AsArrayTypeDeserializer#deserializeTypedFromAny(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.jsontype.impl.AsArrayTypeDeserializer#_deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.jsontype.impl.AsArrayTypeDeserializer#_findDeserializer(Lcom/fasterxml/jackson/databind/DeserializationContext;Ljava/lang/String;)Lcom/fasterxml/jackson/databind/JsonDeserializer;
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows real Spring mechanisms: bean initialization (invokeInitMethods → InitializingBean.afterPropertiesSet), DataSourceInitializerInvoker publishing a ContextRefreshedEvent, and Spring Data's ResourceReaderRepositoryPopulator/Jackson2ResourceReader, which genuinely uses Jackson's ObjectReader to deserialize repository-population JSON with polymorphic typing. The only CHA-style hop is the afterPropertiesSet dispatch, but the target class is a real Spring Boot bean, and the Jackson tail (readTree → TypeWrappedDeserializer → AsArrayTypeDeserializer._findDeserializer) is the actual polymorphic deserialization path.


### CVE-2026-91777

```
org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#getAuthenticationManagerBean()Lorg/springframework/security/authentication/AuthenticationManager;
 → org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration#lazyBean(Ljava/lang/Class;)Ljava/lang/Object;
 → org.springframework.security.config.annotation.configuration.AutowireBeanFactoryObjectPostProcessor#postProcess(Ljava/lang/Object;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#afterPropertiesSet()V
 → org.springframework.boot.autoconfigure.jdbc.DataSourceInitializerInvoker#initialize(Lorg/springframework/boot/autoconfigure/jdbc/DataSourceInitializer;)V
 → org.springframework.context.support.AbstractApplicationContext#publishEvent(Ljava/lang/Object;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.data.repository.init.AbstractRepositoryPopulatorFactoryBean#onApplicationEvent(Lorg/springframework/context/event/ContextRefreshedEvent;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#populate(Lorg/springframework/data/repository/support/Repositories;)V
 → org.springframework.data.repository.init.ResourceReaderRepositoryPopulator#readObjectFrom(Lorg/springframework/core/io/Resource;)Ljava/lang/Object;
 → org.springframework.data.repository.init.Jackson2ResourceReader#readFrom(Lorg/springframework/core/io/Resource;Ljava/lang/ClassLoader;)Ljava/lang/Object;
 → com.fasterxml.jackson.databind.ObjectReader#readTree(Ljava/io/InputStream;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAndCloseAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.ObjectReader#_bindAsTree(Lcom/fasterxml/jackson/core/JsonParser;)Lcom/fasterxml/jackson/databind/JsonNode;
 → com.fasterxml.jackson.databind.deser.std.MapDeserializer#deserialize(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;)Ljava/util/Map;
 → com.fasterxml.jackson.databind.deser.std.MapDeserializer#_readAndBindStringKeyMap(Lcom/fasterxml/jackson/core/JsonParser;Lcom/fasterxml/jackson/databind/DeserializationContext;Ljava/util/Map;)V
 → com.fasterxml.jackson.databind.deser.std.MapDeserializer#handleUnresolvedReference(Lcom/fasterxml/jackson/databind/DeserializationContext;Lcom/fasterxml/jackson/databind/deser/std/MapDeserializer$MapReferringAccumulator;Ljava/lang/Object;Lcom/fasterxml/jackson/databind/deser/UnresolvedForwardReference;)V
```

**Entry call (app → dependency):** `org.owasp.webgoat.WebSecurityConfig#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;` → `org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter#authenticationManager()Lorg/springframework/security/authentication/AuthenticationManager;`


**Path plausibility (z-ai/glm-5.3):** artifact — The final segment is implausible: ObjectReader#readTree deserializes into JsonNode via JsonNodeDeserializer, and JsonNode maps have no @JsonIdentityInfo, so MapDeserializer#handleUnresolvedReference (which requires an identity-enabled POJO map with forward references) cannot be reached this way. Additionally, the hop from AbstractAutowireCapableBeanFactory#invokeInitMethods to DataSourceInitializerInvoker#afterPropertiesSet is a generic InitializingBean CHA edge, indicating the chain was stitched through over-approximated framework dispatch.




## Unidentified dependency jars

No Maven coordinates could be resolved for these (no `pom.properties` inside the jar
and no matching entry in `deps.txt`), so they were not checked against OSV at all.

None.


## Limitations

- **Reflection, dynamic proxies (AOP/CGLIB, Spring Data repositories, JDK `Proxy`),
  and `invokedynamic` sites other than lambdas/method references** are invisible to a
  static call graph. A Spring Data repository interface, for instance, has no bytecode
  implementation at all — Spring generates it at runtime — so a call to one can only be
  seen as reaching the interface method itself, never the generated proxy's body.
- **Serialization gadgets and config-driven behavior** (XML wiring, `web.xml`,
  `ServiceLoader`) are not modeled; a REACHABLE/FRAMEWORK/NOT_REACHABLE verdict only
  reflects direct and virtual/interface (CHA-expanded) method calls found in bytecode.
- **Shaded/relocated dependencies**: a jar bundling relocated classes under a different
  package than its own Maven coordinates is not detected as such.
- **CHA over-approximates**: a virtual/interface call is treated as reaching every
  subtype that declares the same-named method, whether or not that specific override is
  ever actually invoked at runtime for that call site. This produces false positives,
  never false negatives, for the method-dispatch part of reachability.
- **Reflection under-approximates**: anything invoked purely via `java.lang.reflect`,
  `MethodHandle` lookups outside a recognized `LambdaMetafactory` call, or a
  dependency-injection framework wiring beans by name/type at runtime, is invisible to
  this analysis and can produce false negatives (a real path exists that ariadne cannot
  see).
- **LLM-assisted mappings (`llm` in the Source column) are unverified, machine-suggested
  method lists.** Each method was checked to exist in the analyzed bytecode, but not that
  it is the vulnerable one. When the model saw only the advisory text (`llm (advisory)`),
  a NOT_REACHABLE verdict is reported as UNKNOWN instead, since that is too weak to call
  safe.

## Build warnings

- class com.sun.activation.viewers.ImageViewerCanvas found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.viewers.ImageViewer found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.viewers.TextViewer found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.viewers.TextEditor found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.LineTokenizer found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.MimeTypeFile found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.MimeTypeEntry found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.MailcapParseException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.MailcapFile found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.MailcapTokenizer found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class com.sun.activation.registries.LogSupport found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParseException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandObject found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSourceDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.FileDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$5 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParameterList found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$3 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.UnsupportedDataTypeException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.FileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandlerDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.ActivationDataFlavor found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$2 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$4 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.ObjectDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandlerFactory found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeType found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.URLDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo$Beans$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParseException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandObject found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSourceDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.FileDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$5 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParameterList found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$3 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.UnsupportedDataTypeException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo$Beans found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.FileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandlerDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.ActivationDataFlavor found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$2 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$4 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.ObjectDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandlerFactory found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeType found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.URLDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataHandlerDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.UnsupportedDataTypeException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport$4 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.FileTypeMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport$5 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.CommandObject found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.ObjectDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.URLDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MimeTypeParseException found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.CommandInfo found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.CommandInfo$Beans$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.ActivationDataFlavor found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.CommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.FileDataSource found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataHandler$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MimeTypeParameterList found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport$1 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataContentHandlerFactory found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataSourceDataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.DataContentHandler found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.CommandInfo$Beans found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport$2 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap$1 found in both work/WebGoat/app-dependency/jakarta.activation-1.2.2.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.MimeType found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.activation.SecuritySupport$3 found in both work/WebGoat/app-dependency/activation-1.1.1.jar and work/WebGoat/app-dependency/javax.activation-api-1.2.0.jar; keeping the first one seen
- class javax.annotation.sql.DataSourceDefinitions found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.sql.DataSourceDefinition found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.Resources found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.Resource found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.Resource$AuthenticationType found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.ManagedBean found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.Generated found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.PreDestroy found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.Priority found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.PostConstruct found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.security.RolesAllowed found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.security.PermitAll found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.security.DeclareRoles found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.security.RunAs found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.annotation.security.DenyAll found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/javax.annotation-api-1.3.2.jar; keeping the first one seen
- class javax.xml.bind.JAXBElement found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.util.JAXBSource found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlType$DEFAULT found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder$2 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder$4 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlType found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlID found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ServiceLoaderUtil found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.DataBindingException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.util.ValidationEventCollector found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.attachment.AttachmentUnmarshaller found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Marshaller found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.XmlAdapter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.CollapsedStringAdapter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.attachment.AttachmentMarshaller found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.PrintConversionEvent found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.XmlJavaTypeAdapter$DEFAULT found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBPermission found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.DatatypeConverter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.HexBinaryAdapter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.XmlJavaTypeAdapters found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.WhiteSpaceProcessor found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ValidationEventLocator found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.TypeConstraintException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlInlineBinaryData found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.PrintConversionEventImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Marshaller$Listener found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.NotIdentifiableEvent found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.DatatypeConverterInterface found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.W3CDomHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.XmlJavaTypeAdapter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.ValidationEventLocatorImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.SchemaOutputResolver found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXB found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.PropertyException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBContext$1 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAnyElement found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.NotIdentifiableEventImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.AbstractUnmarshallerImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.UnmarshallerHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder$3 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAccessOrder found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlList found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.Messages found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ValidationException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBContextFactory found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlSeeAlso found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.DomHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlNs found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.ValidationEventImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ModuleUtil found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlSchemaType found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementRefs found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.ParseConversionEventImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Binder found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.DefaultValidationEventHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Validator found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementDecl$GLOBAL found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlSchemaTypes found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.helpers.AbstractMarshallerImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXB$Cache found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlSchemaType$DEFAULT found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlEnumValue found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlSchema found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElements found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementDecl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlTransient found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ServiceLoaderUtil$ExceptionHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlMimeType found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Unmarshaller found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder$5 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElement$DEFAULT found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ValidationEventHandler found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlIDREF found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.DatatypeConverterImpl$CalendarFormatter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementWrapper found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBIntrospector found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlRootElement found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlMixed found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAnyAttribute found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElement found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAccessorType found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Unmarshaller$Listener found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAttachmentRef found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder$1 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlValue found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlNsForm found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBElement$GlobalScope found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlEnum found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementRef$DEFAULT found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.DatatypeConverterImpl found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAccessorOrder found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlElementRef found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ParseConversionEvent found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAttribute found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlRegistry found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ContextFinder found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Messages found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.XmlAccessType found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.MarshalException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.Element found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.util.JAXBSource$1 found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.util.Messages found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.UnmarshalException found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.GetPropertyAction found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.annotation.adapters.NormalizedStringAdapter found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.ValidationEvent found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.util.JAXBResult found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.xml.bind.JAXBContext found in both work/WebGoat/app-dependency/jakarta.xml.bind-api-2.3.3.jar and work/WebGoat/app-dependency/jaxb-api-2.3.1.jar; keeping the first one seen
- class javax.annotation.security.RolesAllowed found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.Resource found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.Priority found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.ManagedBean found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.Resource$AuthenticationType found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.Resources found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.Generated found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.sql.DataSourceDefinition found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.security.DeclareRoles found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.PreDestroy found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.sql.DataSourceDefinitions found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.security.RunAs found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.PostConstruct found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.security.PermitAll found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class javax.annotation.security.DenyAll found in both work/WebGoat/app-dependency/jakarta.annotation-api-1.3.5.jar and work/WebGoat/app-dependency/jboss-annotations-api_1.3_spec-2.0.1.Final.jar; keeping the first one seen
- class org.xmlpull.v1.XmlPullParser found in both work/WebGoat/app-dependency/xmlpull-1.1.3.1.jar and work/WebGoat/app-dependency/xpp3_min-1.1.4c.jar; keeping the first one seen
- class org.xmlpull.v1.XmlPullParserException found in both work/WebGoat/app-dependency/xmlpull-1.1.3.1.jar and work/WebGoat/app-dependency/xpp3_min-1.1.4c.jar; keeping the first one seen


