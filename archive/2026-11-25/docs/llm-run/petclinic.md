# ariadne reachability report

- **App:** work/spring-petclinic/target/classes
- **Generated:** 2026-10-14 13:29:01 CEST
- **ariadne version:** v0.0.0-20261014100226-85b4009bd1c9+dirty
- **Profile:** spring+servlet
- **Max bytecode version seen:** 52
- **Call graph:** 273180 nodes, 1230366 edges (598850 direct, 630750 added by CHA)
- **Unresolved `invokedynamic` sites:** 0
- **Dependencies:** 87 total, 87 identified
- **LLM-assisted mapping:** z-ai/glm-5.3 (5 CVEs mapped, 0 skipped)
- **Exploit-check:** z-ai/glm-5.3 (3 REACHABLE CVEs checked, 0 skipped)
- **Framework-check:** z-ai/glm-5.3 (61 FRAMEWORK CVEs checked, 0 skipped)

## Summary

**148 CVEs analyzed.**

| Status | Count |
|---|---|
| REACHABLE | 3 |
| FRAMEWORK | 61 |
| NOT_REACHABLE | 31 |
| UNKNOWN | 47 |
| NOT_APPLICABLE | 6 |

NOT_APPLICABLE means every affected dependency is a WebJar: client-side JavaScript/CSS
with no JVM bytecode, which a call graph cannot analyze. It is **not** a verdict that the
app is safe: the vulnerability runs in the browser, so check whether the app's pages use
the affected component.

| Severity | Count |
|---|---|
| CRITICAL | 18 |
| HIGH | 57 |
| LOW | 13 |
| MODERATE | 60 |


## CVEs

| CVE | Package | Version | Severity | Status | Vulnerable method(s) | Source |
|---|---|---|---|---|---|---|
| CVE-2016-1000027 | org.springframework:spring-web | 5.3.6 | CRITICAL | NOT_REACHABLE | `org.springframework.remoting.httpinvoker.HttpInvokerServiceExporter#handleRequest`, `org.springframework.remoting.httpinvoker.SimpleHttpInvokerServiceExporter#handle` | llm (diff) |
| CVE-2016-10735 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2018-14040 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2018-14042 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2018-20676 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2018-20677 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2019-8331 | org.webjars:bootstrap | 3.3.6 | MODERATE | NOT_APPLICABLE |  |  |
| CVE-2020-36518 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#createContextual`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapArrayToArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializer#mapObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#<init>`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeAnyScalar`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeNR`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_deserializeObjectAtName`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_mapObjectWithDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_putValueHandleDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#_squashDups`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#addValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#childArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#childObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#deserialize`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#deserializeWithType`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#emptyList`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#emptyMap`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishBranchArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishBranchObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishRootArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#finishRootObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#instance`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#isObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#logicalType`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#putDeferredValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#putValue`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#resetAsArray`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#resetAsObject`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#rootArrayScope`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#rootObjectScope`, `com.fasterxml.jackson.databind.deser.std.UntypedObjectDeserializerNR#supportsUpdate` | diff |
| CVE-2021-22060 | org.springframework:spring-core | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2021-22096 | org.springframework:spring-core | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2021-22118 | org.springframework:spring-web | 5.3.6 | HIGH | FRAMEWORK | `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#getBlockingOperationScheduler`, `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#read`, `org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#setFileStorageDirectory`, `org.springframework.http.codec.multipart.FileStorage#<init>`, `org.springframework.http.codec.multipart.FileStorage#createNewDirectoryIfDeleted`, `org.springframework.http.codec.multipart.FileStorage#directory`, `org.springframework.http.codec.multipart.FileStorage#fromPath`, `org.springframework.http.codec.multipart.FileStorage#tempDirectory`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#<init>`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#accept`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#createTempDirectory`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#read`, `org.springframework.http.codec.multipart.SynchronossPartHttpMessageReader#setFileStorageDirectory` | diff |
| CVE-2021-23463 | com.h2database:h2 | 1.4.200 | HIGH | NOT_REACHABLE | `org.h2.jdbc.JdbcSQLXML#getSource`, `org.h2.jdbc.JdbcSQLXML#setResult` | diff |
| CVE-2021-2471 | mysql:mysql-connector-java | 8.0.23 | MODERATE | UNKNOWN |  |  |
| CVE-2021-28170 | org.glassfish:jakarta.el | 3.0.3 | MODERATE | UNKNOWN |  |  |
| CVE-2021-42392 | com.h2database:h2 | 1.4.200 | CRITICAL | UNKNOWN |  |  |
| CVE-2021-42550 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | FRAMEWORK | `ch.qos.logback.classic.selector.ContextJNDISelector#findConfigFileURL`, `ch.qos.logback.classic.util.JNDIUtil#lookup`, `ch.qos.logback.core.util.OptionHelper#isEmpty` | diff |
| CVE-2021-43466 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2021-46877 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.databind.node.NodeSerialization#_read`, `com.fasterxml.jackson.databind.node.NodeSerialization#readExternal` | diff |
| CVE-2021-47621 | io.github.classgraph:classgraph | 4.8.69 | MODERATE | NOT_REACHABLE | `nonapi.io.github.classgraph.utils.VersionFinder#getSecureDocumentBuilderFactory`, `nonapi.io.github.classgraph.utils.VersionFinder#getSecureXPathFactory`, `nonapi.io.github.classgraph.utils.VersionFinder#getVersion` | diff |
| CVE-2022-1471 | org.yaml:snakeyaml | 1.27 | HIGH | FRAMEWORK | `org.yaml.snakeyaml.Yaml#load`, `org.yaml.snakeyaml.Yaml#loadAll`, `org.yaml.snakeyaml.Yaml#loadAs`, `org.yaml.snakeyaml.constructor.Constructor#<init>` | manual |
| CVE-2022-21363 | mysql:mysql-connector-java | 8.0.23 | MODERATE | UNKNOWN |  |  |
| CVE-2022-22950 | org.springframework:spring-expression | 5.3.6 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ast.ConstructorReference#checkNumElements`, `org.springframework.expression.spel.ast.ConstructorReference#createArray` | diff |
| CVE-2022-22965 | org.springframework.boot:spring-boot-starter-web | 2.4.5 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22965 | org.springframework:spring-beans | 5.3.6 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22965 | org.springframework:spring-webmvc | 5.3.6 | CRITICAL | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces` | diff |
| CVE-2022-22968 | org.springframework:spring-context | 5.3.6 | HIGH | REACHABLE | `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields` | diff |
| CVE-2022-22970 | org.springframework:spring-beans | 5.3.6 | HIGH | FRAMEWORK | `org.springframework.beans.CachedIntrospectionResults#<init>`, `org.springframework.beans.CachedIntrospectionResults#introspectInterfaces`, `org.springframework.beans.CachedIntrospectionResults#isInvalidReadOnlyPropertyType`, `org.springframework.beans.CachedIntrospectionResults#isPlainAccessor` | diff |
| CVE-2022-23221 | com.h2database:h2 | 1.4.200 | CRITICAL | UNKNOWN |  |  |
| CVE-2022-25857 | org.yaml:snakeyaml | 1.27 | HIGH | FRAMEWORK | `org.yaml.snakeyaml.LoaderOptions#getNestingDepthLimit`, `org.yaml.snakeyaml.LoaderOptions#setNestingDepthLimit`, `org.yaml.snakeyaml.composer.Composer#<init>`, `org.yaml.snakeyaml.composer.Composer#composeNode`, `org.yaml.snakeyaml.composer.Composer#decreaseNestingDepth`, `org.yaml.snakeyaml.composer.Composer#increaseNestingDepth` | diff |
| CVE-2022-38749 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38750 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38751 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-38752 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-41854 | org.yaml:snakeyaml | 1.27 | MODERATE | UNKNOWN |  |  |
| CVE-2022-42003 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.BeanDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.BeanDeserializerBase#_delegateDeserializer`, `com.fasterxml.jackson.databind.deser.BeanDeserializerBase#deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.DateDeserializers#_parseDate`, `com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.MapEntryDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.NumberDeserializers#deserialize`, `com.fasterxml.jackson.databind.deser.std.StackTraceElementDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeFromEmpty`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeWrappedValue`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBoolean`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBooleanPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseByte`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBytePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDate`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDateFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDouble`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDoublePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloat`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloatPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseIntPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseInteger`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLong`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLongPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShort`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShortPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseString`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#handleNestedArrayForSingle`, `com.fasterxml.jackson.databind.deser.std.StringDeserializer#deserialize` | diff |
| CVE-2022-42004 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.BeanDeserializer#_deserializeFromArray`, `com.fasterxml.jackson.databind.deser.std.DateDeserializers#_parseDate`, `com.fasterxml.jackson.databind.deser.std.EnumDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.NumberDeserializers#deserialize`, `com.fasterxml.jackson.databind.deser.std.StackTraceElementDeserializer#deserialize`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_deserializeWrappedValue`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBoolean`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBooleanPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseByte`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseBytePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDate`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDateFromArray`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDouble`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseDoublePrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloat`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseFloatPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseIntPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseInteger`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLong`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseLongPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShort`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseShortPrimitive`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#_parseString`, `com.fasterxml.jackson.databind.deser.std.StdDeserializer#handleNestedArrayForSingle`, `com.fasterxml.jackson.databind.deser.std.StringDeserializer#deserialize` | diff |
| CVE-2022-42252 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.coyote.http11.Http11InputBuffer#parseHeader`, `org.apache.coyote.http11.Http11InputBuffer#skipLine` | diff |
| CVE-2022-45143 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.valves.JsonErrorReportValve#report`, `org.apache.tomcat.util.json.JSONFilter#<init>`, `org.apache.tomcat.util.json.JSONFilter#escape` | diff |
| CVE-2022-45868 | com.h2database:h2 | 1.4.200 | HIGH | NOT_REACHABLE | `org.h2.server.web.WebServer#encodeAdminPassword`, `org.h2.server.web.WebServer#setAdminPassword`, `org.h2.tools.Server#main`, `org.h2.tools.Server#runTool`, `org.h2.tools.Server#verifyArgs` | diff |
| CVE-2023-1932 | org.hibernate.validator:hibernate-validator | 6.1.7.Final | MODERATE | UNKNOWN | `org.hibernate.validator.internal.constraintvalidators.hv.SafeHtmlValidator#isValid` | llm (advisory) |
| CVE-2023-20860 | org.springframework:spring-webmvc | 5.3.6 | CRITICAL | FRAMEWORK | `org.springframework.web.servlet.handler.HandlerMappingIntrospector#<init>`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#afterPropertiesSet`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#doWithHandlerMapping`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#doWithHandlerMappingIgnoringException`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#getCorsConfiguration`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#getMatchableHandlerMapping`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#initFallback`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#initHandlerMappings`, `org.springframework.web.servlet.handler.HandlerMappingIntrospector#match`, `org.springframework.web.servlet.handler.PathPatternMatchableHandlerMapping#<init>` | diff |
| CVE-2023-20861 | org.springframework:spring-expression | 5.3.6 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ast.OperatorMatches#<init>`, `org.springframework.expression.spel.ast.OperatorMatches#getValueInternal`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#eatRelationalExpression` | diff |
| CVE-2023-20863 | org.springframework:spring-expression | 5.3.6 | HIGH | FRAMEWORK | `org.springframework.expression.spel.standard.InternalSpelExpressionParser#checkExpressionLength`, `org.springframework.expression.spel.standard.InternalSpelExpressionParser#doParseExpression` | diff |
| CVE-2023-20873 | org.springframework.boot:spring-boot-actuator-autoconfigure | 2.4.5 | CRITICAL | NOT_REACHABLE | `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.CloudFoundryWebFluxEndpointHandlerMapping#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.CloudFoundryWebFluxEndpointHandlerMapping#getAllEndpoints`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#cloudFoundryWebFluxEndpointHandlerMapping`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#getPaths`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#postProcess`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#postProcessAfterInitialization`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.reactive.ReactiveCloudFoundryActuatorAutoConfiguration#webFilterChainPostProcessor`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#cloudFoundryWebEndpointServletHandlerMapping`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#customize`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryActuatorAutoConfiguration#ignoreCloudFoundryPathsWebSecurityCustomizer`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryWebEndpointServletHandlerMapping#<init>`, `org.springframework.boot.actuate.autoconfigure.cloudfoundry.servlet.CloudFoundryWebEndpointServletHandlerMapping#getAllEndpoints` | diff |
| CVE-2023-20883 | org.springframework.boot:spring-boot-autoconfigure | 2.4.5 | HIGH | NOT_REACHABLE | `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#createWelcomePageHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#getIndexHtmlResource`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#welcomePageHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WebMvcAutoConfiguration#welcomePageNotAcceptableHandlerMapping`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#getViewName`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#isTemplated`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePage#resolve`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#getHandlerInternal`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageHandlerMapping#isHtmlTextAccepted`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#<init>`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#getHandlerInternal`, `org.springframework.boot.autoconfigure.web.servlet.WelcomePageNotAcceptableHandlerMapping#handleRequest` | diff |
| CVE-2023-22102 | mysql:mysql-connector-java | 8.0.23 | HIGH | UNKNOWN |  |  |
| CVE-2023-24998 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.connector.Request#parseParts`, `org.apache.commons.fileupload.FileCountLimitExceededException#<init>`, `org.apache.commons.fileupload.FileCountLimitExceededException#getLimit`, `org.apache.commons.fileupload.FileUploadBase#getFileCountMax`, `org.apache.commons.fileupload.FileUploadBase#parseRequest`, `org.apache.commons.fileupload.FileUploadBase#setFileCountMax`, `org.apache.tomcat.util.http.Parameters#addParameter`, `org.apache.tomcat.util.http.Parameters#size`, `org.apache.tomcat.util.http.fileupload.FileUploadBase#getFileCountMax`, `org.apache.tomcat.util.http.fileupload.FileUploadBase#parseRequest`, `org.apache.tomcat.util.http.fileupload.FileUploadBase#setFileCountMax`, `org.apache.tomcat.util.http.fileupload.impl.FileCountLimitExceededException#<init>`, `org.apache.tomcat.util.http.fileupload.impl.FileCountLimitExceededException#getLimit` | diff |
| CVE-2023-34055 | org.springframework.boot:spring-boot-actuator | 2.4.5 | MODERATE | FRAMEWORK | `org.springframework.boot.actuate.metrics.web.reactive.server.WebFluxTags#method`, `org.springframework.boot.actuate.metrics.web.servlet.WebMvcTags#method` | diff |
| CVE-2023-41080 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.catalina.authenticator.FormAuthenticator#savedRequestURL` | diff |
| CVE-2023-42795 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.catalina.connector.Request#recycle`, `org.apache.catalina.core.ApplicationHttpRequest#recycle`, `org.apache.tomcat.util.buf.B2CConverter#recycle`, `org.apache.tomcat.util.buf.C2BConverter#recycle` | diff |
| CVE-2023-44487 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `io.netty.handler.codec.http2.AbstractHttp2ConnectionHandlerBuilder#buildFromCodec`, `io.netty.handler.codec.http2.AbstractHttp2ConnectionHandlerBuilder#decoderEnforceMaxRstFramesPerWindow`, `io.netty.handler.codec.http2.Http2FrameCodecBuilder#decoderEnforceMaxRstFramesPerWindow`, `io.netty.handler.codec.http2.Http2MaxRstFrameDecoder#<init>`, `io.netty.handler.codec.http2.Http2MaxRstFrameDecoder#frameListener`, `io.netty.handler.codec.http2.Http2MaxRstFrameDecoder#frameListener0`, `io.netty.handler.codec.http2.Http2MaxRstFrameListener#<init>`, `io.netty.handler.codec.http2.Http2MaxRstFrameListener#onRstStreamRead`, `io.netty.handler.codec.http2.Http2MultiplexCodecBuilder#decoderEnforceMaxRstFramesPerWindow`, `org.apache.coyote.http2.Http2Protocol#getOverheadResetFactor`, `org.apache.coyote.http2.Http2Protocol#setOverheadResetFactor`, `org.apache.coyote.http2.Http2UpgradeHandler#priorityUpdate`, `org.apache.coyote.http2.Http2UpgradeHandler#reset` | diff |
| CVE-2023-45648 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.coyote.http11.Http11InputBuffer#parseHeader`, `org.apache.coyote.http11.filters.ChunkedInputFilter#parseHeader` | diff |
| CVE-2023-46589 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.connector.InputBuffer#realReadBytes`, `org.apache.catalina.connector.Request#doParseParameters`, `org.apache.catalina.connector.Request#getParameter`, `org.apache.catalina.connector.Request#getParameterNames`, `org.apache.catalina.connector.Request#getParameterValues`, `org.apache.catalina.connector.Request#parseParameters`, `org.apache.catalina.connector.Request#parseParts`, `org.apache.catalina.connector.Request#readChunkedPostBody`, `org.apache.catalina.connector.Request#recycle`, `org.apache.catalina.core.StandardWrapperValve#exception`, `org.apache.catalina.core.StandardWrapperValve#invoke`, `org.apache.tomcat.util.http.InvalidParameterException#<init>`, `org.apache.tomcat.util.http.InvalidParameterException#getErrorCode`, `org.apache.tomcat.util.http.Parameters#addParameter`, `org.apache.tomcat.util.http.Parameters#processParameters` | diff |
| CVE-2023-6378 | ch.qos.logback:logback-classic | 1.2.3 | HIGH | FRAMEWORK | `ch.qos.logback.classic.spi.LoggingEventVO#readObject`, `ch.qos.logback.core.net.HardenedObjectInputStream#<init>`, `ch.qos.logback.core.net.HardenedObjectInputStream#initObjectFilter`, `ch.qos.logback.core.util.EnvUtil#getJDKVersion`, `ch.qos.logback.core.util.EnvUtil#isJDK9OrHigher`, `ch.qos.logback.core.util.EnvUtil#isJDK_N_OrHigher` | diff |
| CVE-2023-6378 | ch.qos.logback:logback-core | 1.2.3 | HIGH | FRAMEWORK | `ch.qos.logback.classic.spi.LoggingEventVO#readObject`, `ch.qos.logback.core.net.HardenedObjectInputStream#<init>`, `ch.qos.logback.core.net.HardenedObjectInputStream#initObjectFilter`, `ch.qos.logback.core.util.EnvUtil#getJDKVersion`, `ch.qos.logback.core.util.EnvUtil#isJDK9OrHigher`, `ch.qos.logback.core.util.EnvUtil#isJDK_N_OrHigher` | diff |
| CVE-2024-12798 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | FRAMEWORK | `ch.qos.logback.classic.boolex.JaninoEventEvaluator#evaluate`, `ch.qos.logback.classic.boolex.JaninoEventEvaluator#getDecoratedExpression`, `ch.qos.logback.classic.boolex.JaninoEventEvaluator#start`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#evaluate`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#setExpression`, `ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#start` | llm (diff) |
| CVE-2024-12801 | ch.qos.logback:logback-core | 1.2.3 | LOW | NOT_REACHABLE | `ch.qos.logback.core.joran.event.SaxEventRecorder#buildSaxParser`, `ch.qos.logback.core.joran.event.SaxEventRecorder#error`, `ch.qos.logback.core.joran.event.SaxEventRecorder#fatalError`, `ch.qos.logback.core.joran.event.SaxEventRecorder#recordEvents`, `ch.qos.logback.core.joran.event.SaxEventRecorder#resolveEntity`, `ch.qos.logback.core.joran.event.SaxEventRecorder#warning` | diff |
| CVE-2024-22243 | org.springframework:spring-web | 5.3.6 | HIGH | UNKNOWN |  |  |
| CVE-2024-22259 | org.springframework:spring-web | 5.3.6 | HIGH | FRAMEWORK | `org.springframework.web.util.UriComponentsBuilder#checkSchemeAndHost`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UriComponentsBuilder#fromOriginHeader`, `org.springframework.web.util.UriComponentsBuilder#fromUriString` | diff |
| CVE-2024-22262 | org.springframework:spring-web | 5.3.6 | HIGH | UNKNOWN |  |  |
| CVE-2024-23672 | org.apache.tomcat.embed:tomcat-embed-websocket | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.tomcat.websocket.WsSession#checkCloseTimeout`, `org.apache.tomcat.websocket.WsSession#closeConnection`, `org.apache.tomcat.websocket.WsSession#doClose`, `org.apache.tomcat.websocket.WsSession#getSessionCloseTimeout`, `org.apache.tomcat.websocket.WsSession#getUserPrincipal`, `org.apache.tomcat.websocket.WsSession#getUserPrincipalInternal`, `org.apache.tomcat.websocket.WsSession#onClose`, `org.apache.tomcat.websocket.WsSession#sendCloseMessage`, `org.apache.tomcat.websocket.WsWebSocketContainer#backgroundProcess`, `org.apache.tomcat.websocket.WsWebSocketContainer#getOpenSessions`, `org.apache.tomcat.websocket.server.WsServerContainer#unregisterSession` | diff |
| CVE-2024-24549 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.coyote.http2.Http2Parser#readContinuationFrame`, `org.apache.coyote.http2.Http2Parser#readHeadersFrame` | diff |
| CVE-2024-34750 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.coyote.http2.Http2AsyncUpgradeHandler#sendStreamReset`, `org.apache.coyote.http2.Http2UpgradeHandler#decrementActiveRemoteStreamCount`, `org.apache.coyote.http2.Http2UpgradeHandler#getStream`, `org.apache.coyote.http2.Http2UpgradeHandler#headersEnd`, `org.apache.coyote.http2.Http2UpgradeHandler#headersStart`, `org.apache.coyote.http2.Http2UpgradeHandler#receivedEndOfStream`, `org.apache.coyote.http2.Http2UpgradeHandler#reset`, `org.apache.coyote.http2.Http2UpgradeHandler#sendStreamReset`, `org.apache.coyote.http2.Http2UpgradeHandler#sentEndOfStream`, `org.apache.coyote.http2.Stream#decrementAndGetActiveRemoteStreamCount` | diff |
| CVE-2024-38286 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.tomcat.util.net.SecureNio2Channel#completed`, `org.apache.tomcat.util.net.SecureNio2Channel#unwrap`, `org.apache.tomcat.util.net.SecureNio2Channel#wrap`, `org.apache.tomcat.util.net.SecureNio2Channel#write`, `org.apache.tomcat.util.net.SecureNioChannel#read`, `org.apache.tomcat.util.net.SecureNioChannel#write` | diff |
| CVE-2024-38808 | org.springframework:spring-expression | 5.3.6 | MODERATE | FRAMEWORK | `org.springframework.expression.EvaluationContext#isAssignmentEnabled`, `org.springframework.expression.spel.ast.Assign#getValueInternal`, `org.springframework.expression.spel.ast.ConstructorReference#createArray`, `org.springframework.expression.spel.ast.OpDec#getValueInternal`, `org.springframework.expression.spel.ast.OpInc#getValueInternal`, `org.springframework.expression.spel.support.SimpleEvaluationContext#<init>`, `org.springframework.expression.spel.support.SimpleEvaluationContext#build`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forPropertyAccessors`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forReadOnlyDataBinding`, `org.springframework.expression.spel.support.SimpleEvaluationContext#forReadWriteDataBinding`, `org.springframework.expression.spel.support.SimpleEvaluationContext#isAssignmentEnabled` | diff |
| CVE-2024-38809 | org.springframework:spring-web | 5.3.6 | MODERATE | NOT_REACHABLE | `org.springframework.http.ETag#<init>`, `org.springframework.http.ETag#formattedTag`, `org.springframework.http.ETag#isWildcard`, `org.springframework.http.ETag#parse`, `org.springframework.http.ETag#tag`, `org.springframework.http.ETag#toString`, `org.springframework.http.ETag#weak`, `org.springframework.http.HttpHeaders#getETagValuesAsList`, `org.springframework.web.context.request.ServletWebRequest#matchRequestedETags`, `org.springframework.web.context.request.ServletWebRequest#validateIfNoneMatch` | diff |
| CVE-2024-38816 | org.springframework:spring-webmvc | 5.3.6 | HIGH | NOT_REACHABLE | `org.springframework.web.reactive.function.server.PathResourceLookupFunction#apply`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#cleanDuplicateSlashes`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#cleanLeadingSlash`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidEncodedInputPath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidEncodedResourcePath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isResourceUnderLocation`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#processPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#apply`, `org.springframework.web.servlet.function.PathResourceLookupFunction#cleanDuplicateSlashes`, `org.springframework.web.servlet.function.PathResourceLookupFunction#cleanLeadingSlash`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidEncodedInputPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidEncodedResourcePath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isResourceUnderLocation`, `org.springframework.web.servlet.function.PathResourceLookupFunction#processPath` | diff |
| CVE-2024-38819 | org.springframework:spring-webmvc | 5.3.6 | HIGH | NOT_REACHABLE | `org.springframework.web.reactive.function.server.PathResourceLookupFunction#decode`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#isInvalidPath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#normalizePath`, `org.springframework.web.reactive.function.server.PathResourceLookupFunction#processPath`, `org.springframework.web.reactive.resource.ResourceWebHandler#decode`, `org.springframework.web.reactive.resource.ResourceWebHandler#isInvalidPath`, `org.springframework.web.reactive.resource.ResourceWebHandler#normalizePath`, `org.springframework.web.reactive.resource.ResourceWebHandler#processPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#decode`, `org.springframework.web.servlet.function.PathResourceLookupFunction#isInvalidPath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#normalizePath`, `org.springframework.web.servlet.function.PathResourceLookupFunction#processPath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#decode`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#isInvalidPath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#normalizePath`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#processPath` | diff |
| CVE-2024-38820 | org.springframework:spring-context | 5.3.6 | MODERATE | REACHABLE | `org.springframework.docs.integration.observability.config.conventions.CustomServerRequestObservationConvention#getContextualName`, `org.springframework.http.HttpHeaders#<init>`, `org.springframework.http.HttpHeaders#setAcceptCharset`, `org.springframework.http.MediaTypeFactory#getMediaTypes`, `org.springframework.http.MediaTypeFactory#parseMimeTypes`, `org.springframework.http.client.JdkClientHttpRequest#buildRequest`, `org.springframework.http.client.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.client.observation.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.http.client.reactive.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.server.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.http.server.reactive.ServletServerHttpRequest#createDefaultHttpHeaders`, `org.springframework.http.server.reactive.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.mock.web.MockHttpServletRequest#setContentType`, `org.springframework.mock.web.MockHttpServletRequest#updateContentTypeHeader`, `org.springframework.mock.web.MockHttpServletResponse#setContentType`, `org.springframework.mock.web.MockHttpServletResponse#setExplicitCharacterEncoding`, `org.springframework.mock.web.MockHttpServletResponse#updateContentTypePropertyAndHeader`, `org.springframework.util.MimeType#<init>`, `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields`, `org.springframework.web.accept.ContentNegotiationManagerFactoryBean#addMediaType`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#<init>`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#lookupMediaType`, `org.springframework.web.accept.PathExtensionContentNegotiationStrategy#getMediaTypeKey`, `org.springframework.web.client.RestClientResponseException#copyHeaders`, `org.springframework.web.filter.ForwardedHeaderFilter#headerNames`, `org.springframework.web.filter.HiddenHttpMethodFilter#doFilterInternal`, `org.springframework.web.filter.reactive.HiddenHttpMethodFilter#mapExchange`, `org.springframework.web.multipart.support.MultipartResolutionDelegate#isMultipartContent`, `org.springframework.web.multipart.support.StandardMultipartHttpServletRequest#handleParseFailure`, `org.springframework.web.reactive.accept.ParameterContentTypeResolver#formatKey`, `org.springframework.web.reactive.function.client.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.web.reactive.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.reactive.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.reactive.resource.ResourceWebHandler#getMediaType`, `org.springframework.web.reactive.resource.ResourceWebHandler#setMediaTypes`, `org.springframework.web.reactive.result.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.method.annotation.AbstractMessageConverterMethodProcessor#safeExtension`, `org.springframework.web.servlet.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.servlet.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#getMediaType`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#setMediaTypes`, `org.springframework.web.servlet.view.xslt.XsltView#configureResponse`, `org.springframework.web.socket.WebSocketExtension#<init>`, `org.springframework.web.socket.adapter.standard.StandardToWebSocketExtensionAdapter#initParameters`, `org.springframework.web.socket.client.standard.StandardWebSocketClient#getPort`, `org.springframework.web.util.DisconnectedClientHelper#isClientDisconnectedException`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UrlPathHelper#removeJsessionid` | diff |
| CVE-2024-38820 | org.springframework:spring-web | 5.3.6 | MODERATE | REACHABLE | `org.springframework.docs.integration.observability.config.conventions.CustomServerRequestObservationConvention#getContextualName`, `org.springframework.http.HttpHeaders#<init>`, `org.springframework.http.HttpHeaders#setAcceptCharset`, `org.springframework.http.MediaTypeFactory#getMediaTypes`, `org.springframework.http.MediaTypeFactory#parseMimeTypes`, `org.springframework.http.client.JdkClientHttpRequest#buildRequest`, `org.springframework.http.client.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.client.observation.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.http.client.reactive.JdkClientHttpResponse#adaptHeaders`, `org.springframework.http.server.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.http.server.reactive.ServletServerHttpRequest#createDefaultHttpHeaders`, `org.springframework.http.server.reactive.observation.DefaultServerRequestObservationConvention#getContextualName`, `org.springframework.mock.web.MockHttpServletRequest#setContentType`, `org.springframework.mock.web.MockHttpServletRequest#updateContentTypeHeader`, `org.springframework.mock.web.MockHttpServletResponse#setContentType`, `org.springframework.mock.web.MockHttpServletResponse#setExplicitCharacterEncoding`, `org.springframework.mock.web.MockHttpServletResponse#updateContentTypePropertyAndHeader`, `org.springframework.util.MimeType#<init>`, `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields`, `org.springframework.web.accept.ContentNegotiationManagerFactoryBean#addMediaType`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#<init>`, `org.springframework.web.accept.MappingMediaTypeFileExtensionResolver#lookupMediaType`, `org.springframework.web.accept.PathExtensionContentNegotiationStrategy#getMediaTypeKey`, `org.springframework.web.client.RestClientResponseException#copyHeaders`, `org.springframework.web.filter.ForwardedHeaderFilter#headerNames`, `org.springframework.web.filter.HiddenHttpMethodFilter#doFilterInternal`, `org.springframework.web.filter.reactive.HiddenHttpMethodFilter#mapExchange`, `org.springframework.web.multipart.support.MultipartResolutionDelegate#isMultipartContent`, `org.springframework.web.multipart.support.StandardMultipartHttpServletRequest#handleParseFailure`, `org.springframework.web.reactive.accept.ParameterContentTypeResolver#formatKey`, `org.springframework.web.reactive.function.client.DefaultClientRequestObservationConvention#getContextualName`, `org.springframework.web.reactive.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.reactive.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.reactive.resource.ResourceWebHandler#getMediaType`, `org.springframework.web.reactive.resource.ResourceWebHandler#setMediaTypes`, `org.springframework.web.reactive.result.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.condition.AbstractNameValueExpression#hashCode`, `org.springframework.web.servlet.mvc.method.annotation.AbstractMessageConverterMethodProcessor#safeExtension`, `org.springframework.web.servlet.resource.CachingResourceResolver#getContentCodingKey`, `org.springframework.web.servlet.resource.EncodedResourceResolver#getAcceptEncoding`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#getMediaType`, `org.springframework.web.servlet.resource.ResourceHttpRequestHandler#setMediaTypes`, `org.springframework.web.servlet.view.xslt.XsltView#configureResponse`, `org.springframework.web.socket.WebSocketExtension#<init>`, `org.springframework.web.socket.adapter.standard.StandardToWebSocketExtensionAdapter#initParameters`, `org.springframework.web.socket.client.standard.StandardWebSocketClient#getPort`, `org.springframework.web.util.DisconnectedClientHelper#isClientDisconnectedException`, `org.springframework.web.util.UriComponentsBuilder#fromHttpUrl`, `org.springframework.web.util.UrlPathHelper#removeJsessionid` | diff |
| CVE-2024-38828 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2024-50379 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.WebResourceLockSet#<init>`, `org.apache.catalina.webresources.DirResourceSet#getLockKey`, `org.apache.catalina.webresources.DirResourceSet#getResource`, `org.apache.catalina.webresources.DirResourceSet#initInternal`, `org.apache.catalina.webresources.DirResourceSet#isCaseSensitive`, `org.apache.catalina.webresources.DirResourceSet#lockForRead`, `org.apache.catalina.webresources.DirResourceSet#lockForWrite`, `org.apache.catalina.webresources.DirResourceSet#unlockForRead`, `org.apache.catalina.webresources.DirResourceSet#unlockForWrite`, `org.apache.catalina.webresources.DirResourceSet#write`, `org.apache.catalina.webresources.FileResource#<init>`, `org.apache.catalina.webresources.FileResource#delete` | diff |
| CVE-2024-56337 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | UNKNOWN |  |  |
| CVE-2025-11226 | ch.qos.logback:logback-core | 1.2.3 | MODERATE | NOT_REACHABLE | `ch.qos.logback.core.model.processor.conditional.IfModelHandler#handle`, `ch.qos.logback.core.model.processor.conditional.IfModelHandler#hasNew` | diff |
| CVE-2025-22233 | org.springframework:spring-context | 5.3.6 | LOW | REACHABLE | `org.springframework.validation.DataBinder#isAllowed`, `org.springframework.validation.DataBinder#setDisallowedFields` | diff |
| CVE-2025-22235 | org.springframework.boot:spring-boot | 2.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2025-24813 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | NOT_REACHABLE | `org.apache.catalina.servlets.DefaultServlet#doPut`, `org.apache.catalina.servlets.DefaultServlet#executePartialPut` | diff |
| CVE-2025-35036 | org.hibernate.validator:hibernate-validator | 6.1.7.Final | MODERATE | NOT_REACHABLE | `org.hibernate.validator.constraintvalidation.HibernateConstraintViolationBuilder#enableExpressionLanguage`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#constraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#customViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.AbstractConfigurationImpl#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#<init>`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#getExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#isCustomViolation`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#isExpressionLanguageEnabled`, `org.hibernate.validator.internal.engine.MessageInterpolatorContext#toString`, `org.hibernate.validator.internal.engine.PredefinedScopeValidatorFactoryImpl#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineBeanMetaDataClassNormalizer`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineGetterPropertySelectionStrategy`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determineLocaleResolver`, `org.hibernate.validator.internal.engine.ValidatorFactoryConfigurationHelper#determinePropertyNodeNameProvider`, `org.hibernate.validator.internal.engine.ValidatorFactoryImpl#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#<init>`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#build`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#setConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.ValidatorFactoryScopedContext#setCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addBeanNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addConstraintViolation`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addContainerElementNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#addPropertyNode`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#buildConstraintViolationWithTemplate`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#enableExpressionLanguage`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintValidatorContextImpl#getDefaultConstraintViolationCreationContext`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#getExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#isCustomViolation`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#isExpressionLanguageEnabled`, `org.hibernate.validator.internal.engine.constraintvalidation.ConstraintViolationCreationContext#toString`, `org.hibernate.validator.internal.engine.constraintvalidation.CrossParameterConstraintValidatorContextImpl#<init>`, `org.hibernate.validator.internal.engine.constraintvalidation.CrossParameterConstraintValidatorContextImpl#buildConstraintViolationWithTemplate`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#addVariablesToElContext`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#bindContextValues`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#getElContext`, `org.hibernate.validator.internal.engine.messageinterpolation.ElTermResolver#interpolate`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanMethodsELContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesELResolver#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesELResolver#invoke`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#addELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.BeanPropertiesElContext#getELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.DisabledFeatureELException#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getCommonPropertyType`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getFeatureDescriptors`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getType`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#getValue`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#invoke`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#isReadOnly`, `org.hibernate.validator.internal.engine.messageinterpolation.el.NoOpElResolver#setValue`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#<init>`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#addELResolver`, `org.hibernate.validator.internal.engine.messageinterpolation.el.VariablesELContext#getELResolver`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#addConstraintFailure`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#createConstraintValidatorContextFor`, `org.hibernate.validator.internal.engine.validationcontext.AbstractValidationContext#interpolate`, `org.hibernate.validator.internal.engine.validationcontext.ParameterExecutableValidationContext#createConstraintValidatorContextFor`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#<init>`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#getConstraintExpressionLanguageFeatureLevel`, `org.hibernate.validator.internal.engine.validationcontext.ValidatorScopedContext#getCustomViolationExpressionLanguageFeatureLevel`, `org.hibernate.validator.messageinterpolation.AbstractMessageInterpolator#interpolateMessage`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#<init>`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#interpretDefaultForConstraints`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#interpretDefaultForCustomViolations`, `org.hibernate.validator.messageinterpolation.ExpressionLanguageFeatureLevel#of`, `org.hibernate.validator.messageinterpolation.ParameterMessageInterpolator#interpolate`, `org.hibernate.validator.messageinterpolation.ResourceBundleMessageInterpolator#interpolate` | diff |
| CVE-2025-41242 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2025-41249 | org.springframework:spring-core | 5.3.6 | HIGH | FRAMEWORK | `org.springframework.core.annotation.AnnotatedMethod#isOverrideFor`, `org.springframework.core.annotation.AnnotationsScanner#hasSameGenericTypeParameters` | diff |
| CVE-2025-46701 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | LOW | NOT_REACHABLE | `org.apache.catalina.servlets.CGIServlet#findCGI`, `org.apache.catalina.servlets.CGIServlet#init`, `org.apache.catalina.servlets.CGIServlet#run`, `org.apache.catalina.servlets.CGIServlet#setCGIEnvironment` | diff |
| CVE-2025-48988 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.connector.Connector#getMaxPartCount`, `org.apache.catalina.connector.Connector#getMaxPartHeaderSize`, `org.apache.catalina.connector.Connector#setMaxPartCount`, `org.apache.catalina.connector.Connector#setMaxPartHeaderSize`, `org.apache.catalina.connector.Request#parseParts` | diff |
| CVE-2025-48989 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.coyote.http2.Http2AsyncUpgradeHandler#sendStreamReset`, `org.apache.coyote.http2.Http2UpgradeHandler#headersContinue`, `org.apache.coyote.http2.Http2UpgradeHandler#increaseOverheadCount`, `org.apache.coyote.http2.Http2UpgradeHandler#reduceOverheadCount`, `org.apache.coyote.http2.Http2UpgradeHandler#sendStreamReset` | diff |
| CVE-2025-49124 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | UNKNOWN |  |  |
| CVE-2025-49125 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.catalina.webresources.AbstractArchiveResourceSet#getResource`, `org.apache.catalina.webresources.AbstractArchiveResourceSet#list`, `org.apache.catalina.webresources.AbstractArchiveResourceSet#listWebAppPaths`, `org.apache.catalina.webresources.AbstractResourceSet#isPathMounted`, `org.apache.catalina.webresources.DirResourceSet#getResource`, `org.apache.catalina.webresources.DirResourceSet#list`, `org.apache.catalina.webresources.DirResourceSet#listWebAppPaths`, `org.apache.catalina.webresources.DirResourceSet#mkdir`, `org.apache.catalina.webresources.DirResourceSet#write` | diff |
| CVE-2025-49128 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.core.JsonLocation#_appendSourceDesc`, `com.fasterxml.jackson.core.JsonLocation#_truncate`, `com.fasterxml.jackson.core.json.ReaderBasedJsonParser#<init>`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_reportInvalidToken` | diff |
| CVE-2025-52434 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.tomcat.util.net.AprEndpoint#close`, `org.apache.tomcat.util.net.AprEndpoint#populateLocalAddr`, `org.apache.tomcat.util.net.AprEndpoint#populateLocalName`, `org.apache.tomcat.util.net.AprEndpoint#populateLocalPort`, `org.apache.tomcat.util.net.AprEndpoint#populateRemoteAddr`, `org.apache.tomcat.util.net.AprEndpoint#populateRemoteHost`, `org.apache.tomcat.util.net.AprEndpoint#populateRemotePort`, `org.apache.tomcat.util.net.AprEndpoint#stopInternal` | diff |
| CVE-2025-52520 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.connector.Request#parseParts` | diff |
| CVE-2025-52999 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | UNKNOWN |  |  |
| CVE-2025-53506 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.coyote.http2.ConnectionSettingsBase#set`, `org.apache.coyote.http2.ConnectionSettingsLocal#set`, `org.apache.coyote.http2.Http2UpgradeHandler#<init>` | diff |
| CVE-2025-55668 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.catalina.connector.Request#recycleSessionInfo`, `org.apache.catalina.valves.rewrite.RewriteValve#invoke` | diff |
| CVE-2025-55752 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.valves.rewrite.RewriteValve#invoke`, `org.apache.catalina.valves.rewrite.RewriteValve#parse` | diff |
| CVE-2025-55754 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | LOW | NOT_REACHABLE | `org.apache.juli.JdkLoggerFormatter#format`, `org.apache.juli.LogUtil#<init>`, `org.apache.juli.LogUtil#escape`, `org.apache.juli.OneLineFormatter#format` | diff |
| CVE-2025-61795 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | LOW | FRAMEWORK | `org.apache.catalina.connector.Request#parseParts` | diff |
| CVE-2025-66614 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.coyote.AbstractProtocol#checkSni`, `org.apache.coyote.http11.AbstractHttp11Protocol#checkSni`, `org.apache.coyote.http11.Http11Processor#prepareRequest`, `org.apache.coyote.http2.Http2UpgradeHandler#getSniHostName`, `org.apache.coyote.http2.Stream#parseAuthority`, `org.apache.tomcat.util.net.AbstractEndpoint#checkSni`, `org.apache.tomcat.util.net.AbstractEndpoint#getStrictSni`, `org.apache.tomcat.util.net.AbstractEndpoint#setStrictSni`, `org.apache.tomcat.util.net.AprEndpoint#getSslContext`, `org.apache.tomcat.util.net.AprEndpoint#setSocketOptions`, `org.apache.tomcat.util.net.SecureNio2Channel#processSNI`, `org.apache.tomcat.util.net.SecureNioChannel#processSNI`, `org.apache.tomcat.util.net.SocketWrapperBase#getSniHostName`, `org.apache.tomcat.util.net.SocketWrapperBase#setSniHostName` | diff |
| CVE-2026-10532 | ch.qos.logback:logback-core | 1.2.3 | LOW | UNKNOWN |  |  |
| CVE-2026-1225 | ch.qos.logback:logback-core | 1.2.3 | LOW | NOT_REACHABLE | `ch.qos.logback.classic.joran.ModelClassToModelHandlerLinker#link`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#<init>`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#addAppenderDeclaration`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#getAppenderNameSet`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#getSupportedModelClass`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#handle`, `ch.qos.logback.core.model.processor.AppenderDeclarationAnalyser#isAppenderDeclared`, `ch.qos.logback.core.model.processor.AppenderRefModelHandler#attachReferencedAppenders`, `ch.qos.logback.core.model.processor.DefaultProcessor#allDependenciesStarted`, `ch.qos.logback.core.model.processor.DefaultProcessor#dependencyIsADirectSubmodel`, `ch.qos.logback.core.model.processor.DependencyDefinition#<init>`, `ch.qos.logback.core.model.processor.DependencyDefinition#getDependency`, `ch.qos.logback.core.model.processor.DependencyDefinition#toString`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#getDependencyNamesForModel`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#hasDependers`, `ch.qos.logback.core.model.processor.ModelInterpretationContext#isNamedDependemcyStarted` | diff |
| CVE-2026-19032 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.databind.ext.NioPathDeserializer#<init>`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#_allowedSchemesDesc`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#_isSchemeAllowed`, `com.fasterxml.jackson.databind.ext.NioPathDeserializer#deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#<init>`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_allowedSchemesDesc`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#_isSchemeAllowed`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#deserialize`, `tools.jackson.databind.deser.jdk.JDKFromStringDeserializer#findDeserializer` | diff |
| CVE-2026-22733 | org.springframework.boot:spring-boot-starter-actuator | 2.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2026-22735 | org.springframework:spring-webmvc | 5.3.6 | LOW | UNKNOWN |  |  |
| CVE-2026-22737 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2026-22741 | org.springframework:spring-webmvc | 5.3.6 | LOW | UNKNOWN |  |  |
| CVE-2026-22745 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2026-24880 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.coyote.http11.filters.ChunkedInputFilter#parseChunkHeader`, `org.apache.coyote.http11.filters.ChunkedInputFilter#recycle`, `org.apache.coyote.http11.filters.ChunkedInputFilter#skipChunkHeader`, `org.apache.tomcat.util.http.parser.ChunkExtension#<init>`, `org.apache.tomcat.util.http.parser.ChunkExtension#parse`, `org.apache.tomcat.util.http.parser.HttpParser#isWhiteSpace` | diff |
| CVE-2026-25854 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | MODERATE | FRAMEWORK | `org.apache.catalina.valves.LoadBalancerDrainingValve#invoke` | llm (advisory) |
| CVE-2026-34483 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | UNKNOWN |  |  |
| CVE-2026-34487 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | NOT_REACHABLE | `org.apache.catalina.tribes.membership.cloud.AbstractStreamProvider#openConnection`, `org.apache.catalina.tribes.membership.cloud.TokenStreamProvider#openStream` | diff |
| CVE-2026-40477 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40477 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40478 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40478 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-40972 | org.springframework.boot:spring-boot-devtools | 2.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2026-40973 | org.springframework.boot:spring-boot | 2.4.5 | HIGH | UNKNOWN |  |  |
| CVE-2026-40984 | io.micrometer:micrometer-core | 1.6.6 | HIGH | NOT_REACHABLE | `io.micrometer.core.instrument.binder.http.HttpJakartaServletRequestTags#method`, `io.micrometer.core.instrument.binder.http.HttpMethods#<init>`, `io.micrometer.core.instrument.binder.http.HttpMethods#isStandard`, `io.micrometer.core.instrument.binder.http.HttpRequestTags#method`, `io.micrometer.core.instrument.binder.jersey.server.DefaultJerseyObservationConvention#getContextualName`, `io.micrometer.core.instrument.binder.jersey.server.HttpMethods#<init>`, `io.micrometer.core.instrument.binder.jersey.server.HttpMethods#isStandard`, `io.micrometer.core.instrument.binder.jersey.server.JerseyKeyValues#method`, `io.micrometer.core.instrument.binder.jersey.server.JerseyTags#method`, `io.micrometer.jetty12.server.DefaultJettyCoreRequestTagsProvider#method`, `io.micrometer.jetty12.server.HttpMethods#<init>`, `io.micrometer.jetty12.server.HttpMethods#isStandard` | diff |
| CVE-2026-41284 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | NOT_REACHABLE | `org.apache.catalina.servlets.WebdavServlet#<init>`, `org.apache.catalina.servlets.WebdavServlet#doLock`, `org.apache.catalina.servlets.WebdavServlet#doPropfind`, `org.apache.catalina.servlets.WebdavServlet#init`, `org.apache.catalina.servlets.WebdavServlet#write` | diff |
| CVE-2026-41293 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | FRAMEWORK | `org.apache.coyote.http2.HPackHuffman#decode`, `org.apache.coyote.http2.HpackDecoder#decode`, `org.apache.coyote.http2.HpackDecoder#readHeaderName`, `org.apache.coyote.http2.HpackDecoder#readHpackString`, `org.apache.coyote.http2.HpackDecoder#readHuffmanString`, `org.apache.coyote.http2.HpackEncoder#encode`, `org.apache.coyote.http2.Http2Parser#readHeaderPayload`, `org.apache.coyote.http2.Stream#emitHeader`, `org.apache.coyote.http2.StreamException#<init>`, `org.apache.tomcat.util.http.parser.HttpParser#isControl`, `org.apache.tomcat.util.http.parser.HttpParser#isFieldContent`, `org.apache.tomcat.util.http.parser.HttpParser#isFieldVChar` | diff |
| CVE-2026-41711 | org.springframework.data:spring-data-commons | 2.4.8 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41716 | org.springframework.data:spring-data-commons | 2.4.8 | HIGH | UNKNOWN |  |  |
| CVE-2026-41721 | org.springframework.data:spring-data-commons | 2.4.8 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41841 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41842 | org.springframework:spring-webmvc | 5.3.6 | HIGH | UNKNOWN |  |  |
| CVE-2026-41843 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41844 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | NOT_REACHABLE | `org.springframework.web.reactive.result.view.ViewResolutionResultHandler#getDefaultViewName`, `org.springframework.web.reactive.result.view.ViewResolutionResultHandler#handleResult`, `org.springframework.web.servlet.view.DefaultRequestToViewNameTranslator#getViewName`, `org.springframework.web.servlet.view.DefaultRequestToViewNameTranslator#transformPath` | diff |
| CVE-2026-41845 | org.springframework:spring-webmvc | 5.3.6 | HIGH | NOT_REACHABLE | `org.springframework.web.util.JavaScriptUtils#javaScriptEscape` | diff |
| CVE-2026-41846 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | NOT_REACHABLE | `org.springframework.web.servlet.tags.form.AbstractHtmlElementTag#resolveCssClass`, `org.springframework.web.servlet.tags.form.AbstractHtmlElementTag#writeOptionalAttributes`, `org.springframework.web.servlet.tags.form.FormTag#resolveCssClass` | diff |
| CVE-2026-41848 | org.springframework:spring-core | 5.3.6 | LOW | FRAMEWORK | `org.springframework.util.AntPathMatcher#<init>`, `org.springframework.util.AntPathMatcher#charAt`, `org.springframework.util.AntPathMatcher#isEmpty`, `org.springframework.util.AntPathMatcher#length`, `org.springframework.util.AntPathMatcher#matchStrings`, `org.springframework.util.AntPathMatcher#subSequence`, `org.springframework.util.AntPathMatcher#toString` | diff |
| CVE-2026-41849 | org.springframework:spring-expression | 5.3.6 | HIGH | UNKNOWN |  |  |
| CVE-2026-41850 | org.springframework:spring-expression | 5.3.6 | HIGH | UNKNOWN |  |  |
| CVE-2026-41851 | org.springframework:spring-expression | 5.3.6 | MODERATE | FRAMEWORK | `org.springframework.expression.spel.ExpressionState#assignVariable`, `org.springframework.expression.spel.ExpressionState#lookupVariable`, `org.springframework.expression.spel.ExpressionState#operate`, `org.springframework.expression.spel.ExpressionState#trackOperation`, `org.springframework.expression.spel.SpelParserConfiguration#<init>`, `org.springframework.expression.spel.SpelParserConfiguration#getMaximumOperations`, `org.springframework.expression.spel.SpelParserConfiguration#retrieveMaxOperations`, `org.springframework.expression.spel.ast.Assign#getValueInternal`, `org.springframework.expression.spel.ast.BeanReference#getValueInternal`, `org.springframework.expression.spel.ast.ConstructorReference#getValueInternal`, `org.springframework.expression.spel.ast.Elvis#getValueInternal`, `org.springframework.expression.spel.ast.FunctionReference#getValueInternal`, `org.springframework.expression.spel.ast.Indexer#<init>`, `org.springframework.expression.spel.ast.Indexer#getArrayElement`, `org.springframework.expression.spel.ast.Indexer#getValue`, `org.springframework.expression.spel.ast.Indexer#getValueRef`, `org.springframework.expression.spel.ast.Indexer#growCollectionIfNecessary`, `org.springframework.expression.spel.ast.Indexer#setArrayElement`, `org.springframework.expression.spel.ast.Indexer#setValue`, `org.springframework.expression.spel.ast.InlineList#<init>`, `org.springframework.expression.spel.ast.InlineList#createList`, `org.springframework.expression.spel.ast.InlineList#determineIfConstant`, `org.springframework.expression.spel.ast.InlineList#getConstantValue`, `org.springframework.expression.spel.ast.InlineList#getValueInternal`, `org.springframework.expression.spel.ast.InlineList#isCompilable`, `org.springframework.expression.spel.ast.InlineList#isConstant`, `org.springframework.expression.spel.ast.InlineMap#<init>`, `org.springframework.expression.spel.ast.InlineMap#createMap`, `org.springframework.expression.spel.ast.InlineMap#determineIfConstant`, `org.springframework.expression.spel.ast.InlineMap#getConstantValue`, `org.springframework.expression.spel.ast.InlineMap#getValueInternal`, `org.springframework.expression.spel.ast.InlineMap#isConstant`, `org.springframework.expression.spel.ast.MethodReference#<init>`, `org.springframework.expression.spel.ast.MethodReference#getValue`, `org.springframework.expression.spel.ast.MethodReference#getValueInternal`, `org.springframework.expression.spel.ast.OpAnd#getValueInternal`, `org.springframework.expression.spel.ast.OpDec#getValueInternal`, `org.springframework.expression.spel.ast.OpDivide#getValueInternal`, `org.springframework.expression.spel.ast.OpEQ#getValueInternal`, `org.springframework.expression.spel.ast.OpGE#getValueInternal`, `org.springframework.expression.spel.ast.OpGT#getValueInternal`, `org.springframework.expression.spel.ast.OpInc#getValueInternal`, `org.springframework.expression.spel.ast.OpLE#getValueInternal`, `org.springframework.expression.spel.ast.OpLT#getValueInternal`, `org.springframework.expression.spel.ast.OpMinus#<init>`, `org.springframework.expression.spel.ast.OpMinus#getValueInternal`, `org.springframework.expression.spel.ast.OpModulus#getValueInternal`, `org.springframework.expression.spel.ast.OpMultiply#getValueInternal`, `org.springframework.expression.spel.ast.OpNE#getValueInternal`, `org.springframework.expression.spel.ast.OpOr#getValueInternal`, `org.springframework.expression.spel.ast.OpPlus#concatenate`, `org.springframework.expression.spel.ast.OpPlus#getValueInternal`, `org.springframework.expression.spel.ast.OperatorBetween#getValueInternal`, `org.springframework.expression.spel.ast.OperatorInstanceof#getValueInternal`, `org.springframework.expression.spel.ast.OperatorMatches#getValueInternal`, `org.springframework.expression.spel.ast.OperatorNot#getValueInternal`, `org.springframework.expression.spel.ast.OperatorPower#getValueInternal`, `org.springframework.expression.spel.ast.Projection#getValueRef`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#<init>`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValue`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValueInternal`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#getValueRef`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#readProperty`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#setValue`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#setValueInternal`, `org.springframework.expression.spel.ast.PropertyOrFieldReference#writeProperty`, `org.springframework.expression.spel.ast.Selection#getValueRef`, `org.springframework.expression.spel.ast.Ternary#getValueInternal`, `org.springframework.expression.spel.ast.TypeReference#getValueInternal` | diff |
| CVE-2026-41852 | org.springframework:spring-expression | 5.3.6 | LOW | UNKNOWN |  |  |
| CVE-2026-41853 | org.springframework:spring-webmvc | 5.3.6 | MODERATE | UNKNOWN |  |  |
| CVE-2026-41901 | org.thymeleaf:thymeleaf | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-41901 | org.thymeleaf:thymeleaf-spring5 | 3.0.12.RELEASE | CRITICAL | UNKNOWN |  |  |
| CVE-2026-42498 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | NOT_REACHABLE | `org.apache.tomcat.websocket.Authenticator#getAuthorization`, `org.apache.tomcat.websocket.BasicAuthenticator#getAuthorization`, `org.apache.tomcat.websocket.DigestAuthenticator#calculateRequestDigest`, `org.apache.tomcat.websocket.DigestAuthenticator#getAuthorization`, `org.apache.tomcat.websocket.WsWebSocketContainer#connectToServerRecursive`, `org.apache.tomcat.websocket.WsWebSocketContainer#processAuthenticationChallenge` | diff |
| CVE-2026-43512 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | FRAMEWORK | `org.apache.catalina.realm.RealmBase#getDigest` | diff |
| CVE-2026-43513 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | HIGH | FRAMEWORK | `org.apache.catalina.realm.LockOutRealm#getCaseSensitive`, `org.apache.catalina.realm.LockOutRealm#isLocked`, `org.apache.catalina.realm.LockOutRealm#registerAuthFailure`, `org.apache.catalina.realm.LockOutRealm#registerAuthSuccess`, `org.apache.catalina.realm.LockOutRealm#setCaseSensitive` | diff |
| CVE-2026-43514 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | LOW | FRAMEWORK | `org.apache.catalina.realm.DigestCredentialHandlerBase#equals`, `org.apache.coyote.ajp.AjpProcessor#prepareRequest`, `org.apache.tomcat.util.security.ConstantTime#<init>`, `org.apache.tomcat.util.security.ConstantTime#equals` | diff |
| CVE-2026-43515 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | FRAMEWORK | `org.apache.catalina.realm.RealmBase#findSecurityConstraints` | diff |
| CVE-2026-47884 | org.springframework:spring-webmvc | 5.3.6 | CRITICAL | FRAMEWORK | `org.springframework.web.servlet.view.xslt.XsltView#getStylesheetSource` | diff |
| CVE-2026-49844 | org.apache.logging.log4j:log4j-api | 2.13.3 | MODERATE | FRAMEWORK | `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatDouble`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatDoubleArray`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatFloat`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatFloatArray`, `org.apache.logging.log4j.message.MapMessageJsonFormatter#formatNumber` | diff |
| CVE-2026-50193 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.node.InternalNodeMapper#_serializeNonRecursive`, `com.fasterxml.jackson.databind.node.InternalNodeMapper#popOrNull`, `com.fasterxml.jackson.databind.node.InternalNodeMapper#push` | diff |
| CVE-2026-54512 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.databind.DatabindContext#_resolveAndValidateGeneric`, `com.fasterxml.jackson.databind.DatabindContext#_validateTypeParameter`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType` | diff |
| CVE-2026-54513 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#<init>`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#allowIfSubTypeIsArray`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#build`, `com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#<init>`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#allowIfSubTypeIsArray`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#build`, `tools.jackson.databind.jsontype.BasicPolymorphicTypeValidator#validateSubType` | diff |
| CVE-2026-54514 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#_inetSocketAddress` | diff |
| CVE-2026-54515 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `tools.jackson.databind.deser.bean.BeanDeserializerBase#createContextual` | diff |
| CVE-2026-65182 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | FRAMEWORK | `org.apache.catalina.realm.RealmBase#findSecurityConstraints` | diff |
| CVE-2026-65905 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | UNKNOWN |  |  |
| CVE-2026-68494 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | NOT_REACHABLE | `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_finishNumberIntegralPart`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_startNegativeNumber`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_startPositiveNumber` | diff |
| CVE-2026-68525 | org.apache.tomcat.embed:tomcat-embed-core | 9.0.45 | CRITICAL | FRAMEWORK | `org.apache.catalina.authenticator.AuthenticatorBase#<init>`, `org.apache.catalina.authenticator.AuthenticatorBase#authenticate`, `org.apache.catalina.authenticator.AuthenticatorBase#checkUserDataConstraints`, `org.apache.catalina.authenticator.AuthenticatorBase#disableCaching`, `org.apache.catalina.authenticator.AuthenticatorBase#doAuthenticateExtended`, `org.apache.catalina.authenticator.AuthenticatorBase#getAuthenticated`, `org.apache.catalina.authenticator.AuthenticatorBase#invoke`, `org.apache.catalina.authenticator.FormAuthenticator#doAuthenticate`, `org.apache.catalina.authenticator.FormAuthenticator#doAuthenticateExtended`, `org.apache.catalina.authenticator.FormAuthenticator#restoreRequest` | diff |
| CVE-2026-77310 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.FromStringDeserializer#_deserialize`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_convertDottedQuadToHex`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_ipStringToBytes`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_parseHextet`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_parseOctet`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_textToNumericFormatV4`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#_textToNumericFormatV6`, `com.fasterxml.jackson.databind.deser.std.InetAddressValidator#isInetAddress` | diff |
| CVE-2026-83557 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | MODERATE | NOT_REACHABLE | `com.fasterxml.jackson.databind.jsontype.DefaultBaseTypeLimitingValidator#validateSubType`, `com.fasterxml.jackson.databind.jsontype.DefaultBaseTypeLimitingValidator$UnsafeBaseTypes#<init>` | llm (diff) |
| CVE-2026-89425 | com.fasterxml.jackson.core:jackson-core | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.core.json.ReaderBasedJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_checkMatchEnd`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_handleUnexpectedValue`, `com.fasterxml.jackson.core.json.UTF8DataInputJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_checkMatchEnd`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_handleUnexpectedValue`, `com.fasterxml.jackson.core.json.UTF8StreamJsonParser#_reportInvalidToken`, `com.fasterxml.jackson.core.json.async.NonBlockingUtf8JsonParserBase#_finishErrorToken` | diff |
| CVE-2026-91776 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.jsontype.impl.TypeDeserializerBase#_findDeserializer` | diff |
| CVE-2026-91777 | com.fasterxml.jackson.core:jackson-databind | 2.11.4 | HIGH | FRAMEWORK | `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#add`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#handleUnresolvedReference`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#resolve`, `com.fasterxml.jackson.databind.deser.std.CollectionDeserializer#resolveForwardReference`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#handleUnresolvedReference`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#put`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#resolve`, `com.fasterxml.jackson.databind.deser.std.MapDeserializer#resolveForwardReference` | diff |
| CVE-2026-9828 | ch.qos.logback:logback-core | 1.2.3 | LOW | UNKNOWN |  |  |



### Notes on individual results

- **CVE-2016-10735** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2018-14040** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2018-14042** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2018-20676** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2018-20677** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2019-8331** (NOT_APPLICABLE): org.webjars:bootstrap 3.3.6 contains only client-side web assets (WebJar: JavaScript/CSS), no JVM bytecode; check whether the app's pages use the affected component

- **CVE-2023-1932** (UNKNOWN): LLM (z-ai/glm-5.3) suggested org.hibernate.validator.internal.constraintvalidators.hv.SafeHtmlValidator#isValid from the advisory text only (no fix diff); not reachable, but unverified


## REACHABLE — full call paths


### CVE-2022-22968

```
org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V
 → org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V` → `org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V`


**Exploitability (z-ai/glm-5.3):** likely — OwnerController#setAllowedFields is a Spring @InitBinder method invoked during binding of externally supplied HTTP form data to the Owner object, so it runs on attacker-controlled requests. It calls the vulnerable WebDataBinder#setDisallowedFields, which is exactly the case-sensitivity flaw (CVE-2022-22968) allowing disallowed fields like 'id' to be set via case variations.


### CVE-2024-38820

```
org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V
 → org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V` → `org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V`


**Exploitability (z-ai/glm-5.3):** unlikely — This is the standard Petclinic @InitBinder pattern calling setDisallowedFields with a hardcoded field name (e.g. "owner.id"), so the string reaching the vulnerable method is application-fixed configuration, not attacker-controlled input. The locale-dependent toLowerCase() weakness in isAllowed would only matter if attacker data flowed through binder field matching, which this static path does not show.


### CVE-2025-22233

```
org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V
 → org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#setAllowedFields(Lorg/springframework/web/bind/WebDataBinder;)V` → `org.springframework.web.bind.WebDataBinder#setDisallowedFields([Ljava/lang/String;)V`


**Exploitability (z-ai/glm-5.3):** likely — OwnerController#setAllowedFields is a Spring @InitBinder method invoked per HTTP request, and it configures disallowedFields on the WebDataBinder that subsequently binds attacker-controlled request parameters to the Owner object. This is exactly the configuration-plus-binding flow that CVE-2025-22233's disallowedFields bypass targets.




## FRAMEWORK — needs manual review

These CVEs' vulnerable methods are only reachable by tracing *through*
framework-package code (Spring MVC, Tomcat, Jackson's `@RequestBody`
handling, ...), which ariadne's static call graph cannot follow past —
that code is driven by reflection/dispatch tables the framework builds at
runtime. A human should check whether the app's own configuration
actually exposes this path.


### CVE-2020-36518

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is the well-known Spring Data repository initializer feature: ResourceReaderRepositoryPopulator uses Jackson2ResourceReader to read data.json via ObjectReader#readTree, which genuinely dispatches to UntypedObjectDeserializer for untyped JSON. The hops correspond to real framework wiring (ContextRefreshedEvent -> populator -> Jackson read), not CHA artifacts.


### CVE-2021-22118

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplicationRunListeners#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplicationRunListeners#lambda$contextLoaded$4(Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/boot/SpringApplicationRunListener;)V
 → org.springframework.boot.context.event.EventPublishingRunListener#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.context.support.WebApplicationObjectSupport#initApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.ContentNegotiatingViewResolver#initServletContext(Ljavax/servlet/ServletContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop ApplicationObjectSupport#setApplicationContext -> WebApplicationObjectSupport#initApplicationContext -> ContentNegotiatingViewResolver#initServletContext is a classic CHA spurious edge (generic ApplicationContextAware callback dispatched onto an unrelated view-resolver subtype), and the path then jumps into the reactive WebFlux stack (DefaultServerWebExchange, DefaultPartHttpMessageReader) which a servlet-based PetClinic app would never execute; the jump from initFormData into MultipartHttpMessageReader with a Flux/ReactiveHttpInputMessage descriptor further indicates over-broad virtual dispatch rather than a real call chain.


### CVE-2021-42550

```
org.springframework.samples.petclinic.owner.Owner#getPets()Ljava/util/List;
 → org.springframework.beans.support.PropertyComparator#sort(Ljava/util/List;Lorg/springframework/beans/support/SortDefinition;)V
 → org.springframework.beans.support.PropertyComparator#<init>(Lorg/springframework/beans/support/SortDefinition;)V
 → org.apache.commons.logging.LogFactory#getLog(Ljava/lang/Class;)Lorg/apache/commons/logging/Log;
 → org.apache.commons.logging.LogFactory#getLog(Ljava/lang/String;)Lorg/apache/commons/logging/Log;
 → org.apache.commons.logging.LogAdapter#createLog(Ljava/lang/String;)Lorg/apache/commons/logging/Log;
 → org.apache.commons.logging.LogAdapter$Slf4jAdapter#createLocationAwareLog(Ljava/lang/String;)Lorg/apache/commons/logging/Log;
 → org.slf4j.LoggerFactory#getLogger(Ljava/lang/String;)Lorg/slf4j/Logger;
 → org.slf4j.LoggerFactory#getILoggerFactory()Lorg/slf4j/ILoggerFactory;
 → org.slf4j.impl.StaticLoggerBinder#getLoggerFactory()Lorg/slf4j/ILoggerFactory;
 → ch.qos.logback.classic.selector.ContextJNDISelector#getLoggerContext()Lch/qos/logback/classic/LoggerContext;
 → ch.qos.logback.classic.selector.ContextJNDISelector#findConfigFileURL(Ljavax/naming/Context;Lch/qos/logback/classic/LoggerContext;)Ljava/net/URL;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.Owner#getPets()Ljava/util/List;` → `org.springframework.beans.support.PropertyComparator#sort(Ljava/util/List;Lorg/springframework/beans/support/SortDefinition;)V`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop StaticLoggerBinder#getLoggerFactory() -> ContextJNDISelector#getLoggerContext() is a CHA over-approximation: StaticLoggerBinder returns the default LogbackLoggerContext, and ContextJNDISelector is only installed when logback is explicitly configured with a JNDI selector, so this edge exists only because both implement ILoggerFactory. The earlier hops (getPets -> PropertyComparator.sort -> logging setup) are plausible, but the JNDI-selector connection is not a real dispatch.


### CVE-2022-1471

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot genuinely loads EnvironmentPostProcessors (including CloudFoundryVcapEnvironmentPostProcessor) via EnvironmentPostProcessorApplicationListener during startup, and its JsonParser implementations include YamlJsonParser, which really calls snakeyaml's Yaml.loadAs to parse VCAP/Application metadata. The only CHA-style hop is the JsonParser.parseMap interface dispatch to YamlJsonParser's lambda, but that is a real framework-selected implementation, not a spurious subtype edge.


### CVE-2022-22950

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#onApplicationEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#processEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#shouldHandle(Lorg/springframework/context/ApplicationEvent;[Ljava/lang/Object;)Z
 → org.springframework.context.event.EventExpressionEvaluator#condition(Ljava/lang/String;Lorg/springframework/context/ApplicationEvent;Ljava/lang/reflect/Method;Lorg/springframework/context/expression/AnnotatedElementKey;[Ljava/lang/Object;Lorg/springframework/beans/factory/BeanFactory;)Z
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.ConstructorReference#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
 → org.springframework.expression.spel.ast.ConstructorReference#createArray(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot startup really does multicast context events to @EventListener methods, and ApplicationListenerMethodAdapter evaluates the listener's condition SpEL via EventExpressionEvaluator, so reaching SpelExpression#getValue during SpringApplication.run is a genuine framework flow. The final hop to ConstructorReference#createArray only occurs if an event condition expression contains an array constructor, which is possible but configuration-dependent; the chain itself is a real dispatch path, not a CHA artifact.


### CVE-2022-22965

```
org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;
 → org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V
 → org.springframework.validation.AbstractPropertyBindingResult#getActualFieldValue(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#getPropertyValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.beans.BeanWrapperImpl#getLocalPropertyHandler(Ljava/lang/String;)Lorg/springframework/beans/BeanWrapperImpl$BeanPropertyHandler;
 → org.springframework.beans.BeanWrapperImpl#getCachedIntrospectionResults()Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#forClass(Ljava/lang/Class;)Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#<init>(Ljava/lang/Class;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;` → `org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — This is the canonical Spring4Shell sink chain: a controller's BindingResult.rejectValue triggers AbstractPropertyBindingResult.getActualFieldValue, which resolves properties via BeanWrapperImpl and CachedIntrospectionResults.forClass, exactly the introspection path (CachedIntrospectionResults#<init>/introspectInterfaces) that CVE-2022-22965 exploits. The Petclinic OwnerController.processFindForm genuinely calls rejectValue on a BindingResult, and every hop is a real Spring data-binding call, not a CHA artifact.


### CVE-2022-22970

```
org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;
 → org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V
 → org.springframework.validation.AbstractPropertyBindingResult#getActualFieldValue(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#getPropertyValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.beans.BeanWrapperImpl#getLocalPropertyHandler(Ljava/lang/String;)Lorg/springframework/beans/BeanWrapperImpl$BeanPropertyHandler;
 → org.springframework.beans.BeanWrapperImpl#getCachedIntrospectionResults()Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#forClass(Ljava/lang/Class;)Lorg/springframework/beans/CachedIntrospectionResults;
 → org.springframework.beans.CachedIntrospectionResults#<init>(Ljava/lang/Class;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;` → `org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain rejectValue → getActualFieldValue → AbstractNestablePropertyAccessor.getPropertyValue → BeanWrapperImpl.getCachedIntrospectionResults → CachedIntrospectionResults.forClass is the genuine Spring data-binding/property-access path; when a controller method takes BindingResult and validation rejects a field, Spring resolves the field value via BeanWrapper, which introspects the bean class exactly this way.


### CVE-2022-25857

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot's EnvironmentPostProcessorApplicationListener really does invoke CloudFoundryVcapEnvironmentPostProcessor on the environment-prepared event, and the VCAP_APPLICATION value is parsed through the JsonParser abstraction, for which Spring Boot's JsonParserFactory really selects YamlJsonParser when snakeyaml is on the classpath (its parseMap lambda calls Yaml#loadAs). The hops match real Spring Boot startup behavior; this is not a CHA artifact.


### CVE-2022-42003

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Data's ResourceReaderRepositoryPopulator with Jackson2ResourceReader is a real, documented mechanism that parses JSON data files via Jackson on context refresh, so the chain from SpringApplication.run through populate/readObjectFrom to ObjectReader.readTree and a deserializer is a genuine framework call path, not a CHA artifact.


### CVE-2022-42004

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is a genuine Spring Data path: AbstractRepositoryPopulatorFactoryBean listens for ContextRefreshedEvent and ResourceReaderRepositoryPopulator/Jackson2ResourceReader really do use Jackson's ObjectReader.readTree to parse repository initializer JSON (e.g., data.json), so reaching EnumDeserializer#deserialize via _bindAsTree is a real dispatch, not a CHA artifact.


### CVE-2022-42252

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http11.Http11Processor#service(Lorg/apache/tomcat/util/net/SocketWrapperBase;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http11.Http11InputBuffer#parseHeaders()Z
 → org.apache.coyote.http11.Http11InputBuffer#parseHeader()Lorg/apache/coyote/http11/Http11InputBuffer$HeaderParseStatus;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The middle segment is implausible: createWebServer/setBasic would never call LifecycleBase#stop(), and Connector#stopInternal() would not invoke UpgradeProcessorInternal#timeoutAsync — these are CHA edges stitching unrelated lifecycle and request-processing code together. The final hops (Http11Processor#service -> parseHeaders -> parseHeader) are the real vulnerable path, but the path as shown does not represent a genuine call chain from the entry point.


### CVE-2022-45143

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.valves.ErrorReportValve#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.valves.JsonErrorReportValve#report(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;Ljava/lang/Throwable;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot startup prefix (main → SpringApplication.run → TomcatServletWebServerFactory.getWebServer) is real, but the middle of the path is pure CHA noise: StandardHost#<init> → StandardPipeline#setBasic, then LifecycleBase#stop → Connector#stopInternal → AbstractProtocol#stop → UpgradeProcessorInternal#timeoutAsync is a spurious interface/virtual dispatch chain (stop() would never route to an upgrade processor's timeoutAsync), so the reported chain to JsonErrorReportValve#report does not correspond to a real call sequence.


### CVE-2023-20860

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplicationRunListeners#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplicationRunListeners#lambda$contextLoaded$4(Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/boot/SpringApplicationRunListener;)V
 → org.springframework.boot.context.event.EventPublishingRunListener#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.context.support.WebApplicationObjectSupport#initApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.ContentNegotiatingViewResolver#initServletContext(Ljavax/servlet/ServletContext;)V
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#initializeBean(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)Ljava/lang/Object;
 → org.springframework.beans.factory.support.AbstractAutowireCapableBeanFactory#invokeInitMethods(Ljava/lang/String;Ljava/lang/Object;Lorg/springframework/beans/factory/support/RootBeanDefinition;)V
 → org.springframework.web.servlet.handler.HandlerMappingIntrospector#afterPropertiesSet()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The tail of the path — AbstractAutowireCapableBeanFactory#initializeBean → invokeInitMethods → HandlerMappingIntrospector#afterPropertiesSet — is the standard Spring InitializingBean lifecycle, so the vulnerable method is genuinely invoked during context startup. The middle hops (EventPublishingRunListener → ApplicationObjectSupport#setApplicationContext) look like CHA noise, but the vulnerable endpoint itself is reached via real framework wiring.


### CVE-2023-20861

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#onApplicationEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#processEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#shouldHandle(Lorg/springframework/context/ApplicationEvent;[Ljava/lang/Object;)Z
 → org.springframework.context.event.EventExpressionEvaluator#condition(Ljava/lang/String;Lorg/springframework/context/ApplicationEvent;Ljava/lang/reflect/Method;Lorg/springframework/context/expression/AnnotatedElementKey;[Ljava/lang/Object;Lorg/springframework/beans/factory/BeanFactory;)Z
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.OperatorMatches#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/spel/support/BooleanTypedValue;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows a genuine Spring Boot startup flow: SpringApplication.run → prepareContext → bootstrap-context close firing an ApplicationEvent, which SimpleApplicationEventMulticaster dispatches to ApplicationListenerMethodAdapter, whose shouldHandle evaluates the listener's SpEL condition via EventExpressionEvaluator.condition → SpelExpression.getValue → OperatorMatches. Event-listener conditions are a real SpEL evaluation point in Spring, so reaching OperatorMatches there is a plausible runtime path, though it depends on a listener declaring a 'matches' condition expression.


### CVE-2023-20863

```
org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;
 → org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V
 → org.springframework.validation.AbstractPropertyBindingResult#getActualFieldValue(Ljava/lang/String;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#getPropertyValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.beans.AbstractNestablePropertyAccessor#setDefaultValue(Lorg/springframework/beans/AbstractNestablePropertyAccessor$PropertyTokenHolder;)Ljava/lang/Object;
 → org.springframework.data.jpa.repository.support.JpaMetamodelEntityInformation$IdentifierDerivingDirectFieldAccessFallbackBeanWrapper#setPropertyValue(Ljava/lang/String;Ljava/lang/Object;)V
 → org.springframework.data.util.ProxyUtils#getUserClass(Ljava/lang/Object;)Ljava/lang/Class;
 → org.springframework.aop.support.AopUtils#getTargetClass(Ljava/lang/Object;)Ljava/lang/Class;
 → org.springframework.aop.target.AbstractBeanFactoryBasedTargetSource#getTargetClass()Ljava/lang/Class;
 → org.springframework.beans.factory.support.AbstractBeanFactory#getType(Ljava/lang/String;Z)Ljava/lang/Class;
 → org.springframework.beans.factory.support.AbstractBeanFactory#predictBeanType(Ljava/lang/String;Lorg/springframework/beans/factory/support/RootBeanDefinition;[Ljava/lang/Class;)Ljava/lang/Class;
 → org.springframework.beans.factory.support.AbstractBeanFactory#resolveBeanClass(Lorg/springframework/beans/factory/support/RootBeanDefinition;Ljava/lang/String;[Ljava/lang/Class;)Ljava/lang/Class;
 → org.springframework.beans.factory.support.AbstractBeanFactory#doResolveBeanClass(Lorg/springframework/beans/factory/support/RootBeanDefinition;[Ljava/lang/Class;)Ljava/lang/Class;
 → org.springframework.beans.factory.support.AbstractBeanFactory#evaluateBeanDefinitionString(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanDefinition;)Ljava/lang/Object;
 → org.springframework.context.expression.StandardBeanExpressionResolver#evaluate(Ljava/lang/String;Lorg/springframework/beans/factory/config/BeanExpressionContext;)Ljava/lang/Object;
 → org.springframework.expression.common.TemplateAwareExpressionParser#parseExpression(Ljava/lang/String;Lorg/springframework/expression/ParserContext;)Lorg/springframework/expression/Expression;
 → org.springframework.expression.spel.standard.InternalSpelExpressionParser#doParseExpression(Ljava/lang/String;Lorg/springframework/expression/ParserContext;)Lorg/springframework/expression/spel/standard/SpelExpression;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.owner.OwnerController#processFindForm(Lorg/springframework/samples/petclinic/owner/Owner;Lorg/springframework/validation/BindingResult;Ljava/util/Map;)Ljava/lang/String;` → `org.springframework.validation.AbstractBindingResult#rejectValue(Ljava/lang/String;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/String;)V`


**Path plausibility (z-ai/glm-5.3):** artifact — The tail (AbstractBeanFactory#predictBeanType → evaluateBeanDefinitionString → StandardBeanExpressionResolver → SpEL parse) is a real Spring chain, but the middle hops are CHA glue: AbstractNestablePropertyAccessor#setDefaultValue does not call setPropertyValue(String,Object) on a JPA wrapper, and ProxyUtils#getUserClass → AopUtils#getTargetClass → AbstractBeanFactoryBasedTargetSource#getTargetClass connects unrelated classes only via same-named virtual dispatch. The link from a controller's rejectValue into bean-factory type prediction during request handling is not a real Spring flow.


### CVE-2023-24998

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.core.StandardEngine#logAccess(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;JZ)V
 → org.apache.catalina.valves.AbstractAccessLogValve#log(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.valves.ExtendedAccessLogValve$RequestParameterElement#addElement(Ljava/io/CharArrayWriter;Ljava/util/Date;Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.connector.Request#getParameter(Ljava/lang/String;)Ljava/lang/String;
 → org.apache.catalina.connector.Request#parseParameters()V
 → org.apache.catalina.connector.Request#parseParts(Z)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The middle hops are CHA artifacts: StandardPipeline#setBasic -> LifecycleBase#stop, AprEndpoint#stopInternal -> SocketWrapperBase#close, and AjpProcessor#recycle -> CoyoteAdapter#checkRecycled are unrelated methods spuriously connected by over-broad virtual dispatch (e.g., a generic stop()/close() call site), not a real Tomcat startup or request-handling sequence. While the tail (Request#getParameter -> parseParameters -> parseParts) is a genuine Tomcat call chain, it is not actually reachable via this fabricated prefix.


### CVE-2023-34055

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplicationRunListeners#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplicationRunListeners#lambda$contextLoaded$4(Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/boot/SpringApplicationRunListener;)V
 → org.springframework.boot.context.event.EventPublishingRunListener#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.context.support.WebApplicationObjectSupport#initApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.ContentNegotiatingViewResolver#initServletContext(Ljavax/servlet/ServletContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — Hop 8→9 is a CHA artifact: EventPublishingRunListener does not extend ApplicationObjectSupport, so contextLoaded cannot dispatch to setApplicationContext; the path then implausibly mixes servlet ContentNegotiatingViewResolver init with a reactive WebFilter chain. The tail (MetricsWebFilter→WebFluxTags#method) is a real runtime chain, but the connecting edges are spurious.


### CVE-2023-41080

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.authenticator.AuthenticatorBase#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.authenticator.FormAuthenticator#doAuthenticate(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;)Z
 → org.apache.catalina.authenticator.FormAuthenticator#savedRequestURL(Lorg/apache/catalina/Session;)Ljava/lang/String;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain from StandardHost#<init>() to StandardPipeline#setBasic to LifecycleBase#stop is implausible (a constructor does not trigger a lifecycle stop), and the path then jumps from server shutdown code (Connector#stopInternal, AbstractProtocol#stop) into live request processing (SocketProcessorBase#run, CoyoteAdapter#asyncDispatch) — these are unrelated runtime contexts stitched together by CHA over-approximation. Only the final hops (AuthenticatorBase#invoke -> FormAuthenticator#doAuthenticate -> savedRequestURL) are a genuine Tomcat dispatch, but they are reached via a spurious connection from Spring Boot startup code.


### CVE-2023-42795

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.connector.Request#recycle()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot startup portion (main -> SpringApplication.run -> refresh -> createWebServer -> TomcatServletWebServerFactory.getWebServer) is plausible, but the Tomcat portion is nonsense: Tomcat#getHost() does not construct a StandardHost, StandardPipeline#setBasic(Valve) does not call LifecycleBase#stop(), and stop() during server creation leading to AjpProcessor#recycle is a CHA-invented chain of unrelated virtual/overridden methods.


### CVE-2023-44487

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http2.Http2UpgradeHandler#upgradeDispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http2.Http2Parser#readFrame(Z)Z
 → org.apache.coyote.http2.Http2Parser#readFrame(ZLorg/apache/coyote/http2/FrameType;)Z
 → org.apache.coyote.http2.Http2Parser#readRstFrame(ILjava/nio/ByteBuffer;)V
 → org.apache.coyote.http2.Http2UpgradeHandler#reset(IJ)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain contains impossible hops: StandardHost#<init> -> StandardPipeline#setBasic -> LifecycleBase#stop (a constructor/setter leading to lifecycle stop during startup) and Connector#stopInternal -> UpgradeProcessorInternal#timeoutAsync, which are not real call relationships. While Tomcat's HTTP/2 stack genuinely processes RST_STREAM frames at runtime, this specific path is stitched together by CHA over-approximation rather than a real dispatch sequence.


### CVE-2023-45648

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#start()V
 → org.apache.catalina.util.LifecycleBase#init()V
 → org.apache.catalina.util.LifecycleMBeanBase#initInternal()V
 → org.apache.catalina.authenticator.AuthenticatorBase#register(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;Ljava/security/Principal;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;ZZ)V
 → org.apache.catalina.connector.Request#getAttribute(Ljava/lang/String;)Ljava/lang/Object;
 → org.apache.coyote.Request#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.AbstractProcessor#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.http11.Http11Processor#setRequestBody(Lorg/apache/tomcat/util/buf/ByteChunk;)V
 → org.apache.coyote.http11.Http11InputBuffer#addActiveFilter(Lorg/apache/coyote/http11/InputFilter;)V
 → org.apache.coyote.http11.filters.BufferedInputFilter#setRequest(Lorg/apache/coyote/Request;)V
 → org.apache.coyote.http11.filters.ChunkedInputFilter#doRead(Lorg/apache/tomcat/util/net/ApplicationBufferHandler;)I
 → org.apache.coyote.http11.filters.ChunkedInputFilter#parseEndChunk()V
 → org.apache.coyote.http11.filters.ChunkedInputFilter#parseHeader()Z
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from LifecycleBase#init into AuthenticatorBase#register is a CHA artifact — register is a per-request authentication callback, never invoked during container startup/init. The subsequent chain (Request#getAttribute -> coyote action -> Http11Processor#setRequestBody -> addActiveFilter -> ChunkedInputFilter) is likewise not a real call sequence; the vulnerable parseHeader is reached during request parsing, not via this path.


### CVE-2023-46589

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.connector.Request#recycle()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop StandardPipeline#setBasic -> LifecycleBase#stop is a CHA artifact: setBasic never stops the pipeline, and a chain of stop()/stopInternal() calls during web-server creation at startup makes no sense. Also Tomcat#getHost does not construct a new StandardHost, and the path never reaches any of the listed vulnerable parameter-parsing methods, only Request#recycle via an implausible recycle/log chain.


### CVE-2023-6378

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.StartupInfoLogger#logStarted(Lorg/apache/commons/logging/Log;Lorg/springframework/util/StopWatch;)V
 → org.apache.commons.logging.LogAdapter$Slf4jLog#info(Ljava/lang/Object;)V
 → ch.qos.logback.classic.Logger#info(Ljava/lang/String;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain Spring Boot startup logging → Logger#callAppenders → RollingFileAppender#rollover → RenameUtil → EnvUtil#isJDK7OrHigher/isJDK_N_OrHigher is genuine logback behavior (RenameUtil really checks JDK version when renaming across volumes), so the edges are real, not CHA artifacts. However, note this only reaches the EnvUtil version-check methods, not LoggingEventVO#readObject, so the actual receiver-based deserialization vulnerability is not exercised by this path.


### CVE-2024-12798

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.StartupInfoLogger#logStarted(Lorg/apache/commons/logging/Log;Lorg/springframework/util/StopWatch;)V
 → org.apache.commons.logging.LogAdapter$Slf4jLog#info(Ljava/lang/Object;)V
 → ch.qos.logback.classic.Logger#info(Ljava/lang/String;)V
 → ch.qos.logback.classic.Logger#filterAndLog_0_Or3Plus(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#buildLoggingEventAndAppend(Ljava/lang/String;Lorg/slf4j/Marker;Lch/qos/logback/classic/Level;Ljava/lang/String;[Ljava/lang/Object;Ljava/lang/Throwable;)V
 → ch.qos.logback.classic.Logger#callAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)V
 → ch.qos.logback.classic.Logger#appendLoopOnAppenders(Lch/qos/logback/classic/spi/ILoggingEvent;)I
 → ch.qos.logback.core.spi.AppenderAttachableImpl#appendLoopOnAppenders(Ljava/lang/Object;)I
 → ch.qos.logback.core.AppenderBase#doAppend(Ljava/lang/Object;)V
 → ch.qos.logback.core.net.SMTPAppenderBase#append(Ljava/lang/Object;)V
 → ch.qos.logback.core.boolex.JaninoEventEvaluatorBase#evaluate(Ljava/lang/Object;)Z
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain from Spring Boot's StartupInfoLogger through the SLF4J adapter into logback's Logger→callAppenders→AppenderAttachableImpl pipeline is the genuine default startup logging flow, and SMTPAppenderBase invoking its configured Evaluator#evaluate is real logback behavior — the logback manual itself documents JaninoEventEvaluator as an SMTPAppender evaluator. The final hop is logback.xml-driven wiring rather than a default (the default evaluator is OnErrorEvaluator, and PetClinic ships no SMTPAppender), but it corresponds to a real, documented appender/evaluator pairing — exactly this CVE's config-injection attack vector — not a spurious CHA connection between unrelated classes.


### CVE-2024-22259

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplicationRunListeners#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplicationRunListeners#lambda$contextLoaded$4(Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/boot/SpringApplicationRunListener;)V
 → org.springframework.boot.context.event.EventPublishingRunListener#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.context.support.WebApplicationObjectSupport#initApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.ContentNegotiatingViewResolver#initServletContext(Ljavax/servlet/ServletContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The startup segment is incoherent: EventPublishingRunListener#contextLoaded cannot dispatch into ApplicationObjectSupport#setApplicationContext, and the path then jumps from servlet-stack ContentNegotiatingViewResolver into AbstractAutowireCapableBeanFactory#initializeBean and a reactive WebFilterChain lambda — mixing bean-initialization callbacks with runtime request handling in a way Spring never performs. While CorsUtils#isSameOrigin really does call UriComponentsBuilder#fromOriginHeader, the chain reaching it (a servlet PetClinic app running reactive CloudFoundry actuator filters during contextLoaded) is a CHA over-approximation, not a real call sequence.


### CVE-2024-23672

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.tomcat.websocket.server.WsHttpUpgradeHandler#upgradeDispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.tomcat.websocket.WsSession#doClose(Ljavax/websocket/CloseReason;Ljavax/websocket/CloseReason;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from StandardHost#<init> to StandardPipeline#setBasic to LifecycleBase#stop, which is nonsensical — a constructor and a valve setter do not trigger lifecycle stop; this is CHA over-approximation on generic virtual calls. Similarly, AbstractProtocol#stop() reaching UpgradeProcessorInternal#timeoutAsync (a per-connection processor method) is not a real dispatch edge, so the path to WsSession#doClose is spurious.


### CVE-2024-24549

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http2.Http2UpgradeHandler#upgradeDispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http2.Http2Parser#readFrame(Z)Z
 → org.apache.coyote.http2.Http2Parser#readFrame(ZLorg/apache/coyote/http2/FrameType;)Z
 → org.apache.coyote.http2.Http2Parser#readContinuationFrame(IIILjava/nio/ByteBuffer;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The startup chain (Spring Boot main -> embedded Tomcat creation) is real, but the tail is bogus: Tomcat#getHost() does not construct a StandardHost, and StandardHost#<init> -> StandardPipeline#setBasic -> LifecycleBase#stop() is a nonsensical constructor-to-stop sequence typical of CHA edges into widely-overridden lifecycle methods. The vulnerable Http2Parser#readContinuationFrame is only reached while processing an actual HTTP/2 request at runtime, not during server startup wiring.


### CVE-2024-34750

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.AbstractProcessor#timeoutAsync(J)V
 → org.apache.coyote.AbstractProcessor#doTimeoutAsync()V
 → org.apache.coyote.http2.StreamProcessor#processSocketEvent(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.coyote.http2.StreamProcessor#process(Lorg/apache/tomcat/util/net/SocketEvent;)V
 → org.apache.coyote.http2.Stream#close(Lorg/apache/coyote/http2/Http2Exception;)V
 → org.apache.coyote.http2.Http2AsyncUpgradeHandler#sendStreamReset(Lorg/apache/coyote/http2/StreamException;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop chain Tomcat#getHost -> StandardHost#<init> -> StandardPipeline#setBasic -> LifecycleBase#stop is nonsensical: getHost does not construct a StandardHost, and setBasic cannot lead to a lifecycle stop during startup. The rest of the chain (stop -> Connector -> AbstractProtocol -> timeoutAsync -> StreamProcessor -> sendStreamReset) is a plausible runtime sequence, but it is only reachable via the bogus LifecycleBase#stop edge, so the overall path is a CHA artifact.


### CVE-2024-38286

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#start()V
 → org.apache.catalina.util.LifecycleBase#init()V
 → org.apache.catalina.util.LifecycleMBeanBase#initInternal()V
 → org.apache.catalina.authenticator.AuthenticatorBase#register(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;Ljava/security/Principal;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;ZZ)V
 → org.apache.catalina.connector.Request#getAttribute(Ljava/lang/String;)Ljava/lang/Object;
 → org.apache.coyote.Request#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.AbstractProcessor#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.ajp.AjpProcessor#isReadyForWrite()Z
 → org.apache.tomcat.util.net.Nio2Endpoint$Nio2SocketWrapper#isReadyForWrite()Z
 → org.apache.tomcat.util.net.Nio2Endpoint$Nio2SocketWrapper#flushNonBlockingInternal(Z)Z
 → org.apache.tomcat.util.net.SecureNio2Channel#write(Ljava/nio/ByteBuffer;)Ljava/util/concurrent/Future;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot startup portion (main -> SpringApplication.run -> createWebServer -> TomcatServletWebServerFactory) is a real embedded-Tomcat startup chain, but the tail is implausible: LifecycleBase#init does not lead to AuthenticatorBase#register (which runs per-request during authentication), and Request#getAttribute -> coyote action -> AjpProcessor#isReadyForWrite -> Nio2SocketWrapper mixes AJP processor with the NIO2 endpoint wrapper in a way that doesn't correspond to real Tomcat wiring. These hops look like CHA edges through generic lifecycle/action dispatch methods rather than a genuine call sequence.


### CVE-2024-38808

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#multicastEvent(Lorg/springframework/context/ApplicationEvent;Lorg/springframework/core/ResolvableType;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#invokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.SimpleApplicationEventMulticaster#doInvokeListener(Lorg/springframework/context/ApplicationListener;Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#onApplicationEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#processEvent(Lorg/springframework/context/ApplicationEvent;)V
 → org.springframework.context.event.ApplicationListenerMethodAdapter#shouldHandle(Lorg/springframework/context/ApplicationEvent;[Ljava/lang/Object;)Z
 → org.springframework.context.event.EventExpressionEvaluator#condition(Ljava/lang/String;Lorg/springframework/context/ApplicationEvent;Ljava/lang/reflect/Method;Lorg/springframework/context/expression/AnnotatedElementKey;[Ljava/lang/Object;Lorg/springframework/beans/factory/BeanFactory;)Z
 → org.springframework.expression.spel.standard.SpelExpression#getValue()Ljava/lang/Object;
 → org.springframework.expression.spel.ast.SpelNodeImpl#getValue(Lorg/springframework/expression/spel/ExpressionState;)Ljava/lang/Object;
 → org.springframework.expression.spel.ast.Assign#getValueInternal(Lorg/springframework/expression/spel/ExpressionState;)Lorg/springframework/expression/TypedValue;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is the real Spring Boot startup event-dispatch chain: SpringApplication.run → prepareContext → multicaster → ApplicationListenerMethodAdapter → EventExpressionEvaluator.condition, which genuinely evaluates @EventListener condition SpEL expressions. The final hop to Assign#getValueInternal only occurs if a condition expression contains an assignment, which is config-dependent, but the framework wiring itself is authentic, not a CHA artifact.


### CVE-2024-50379

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#start()V
 → org.apache.catalina.util.LifecycleBase#init()V
 → org.apache.catalina.webresources.DirResourceSet#initInternal()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot's embedded Tomcat startup genuinely constructs and initializes Catalina components (including WebResourceRoot/DirResourceSet) via LifecycleBase#init during createWebServer, so reaching DirResourceSet#initInternal from PetClinicApplication#main is a real framework call chain. The addValve->start hop is slightly imprecise but the overall Spring Boot -> Tomcat lifecycle initialization sequence is authentic.


### CVE-2025-41249

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#handleRunFailure(Lorg/springframework/context/ConfigurableApplicationContext;Ljava/lang/Throwable;Lorg/springframework/boot/SpringApplicationRunListeners;)V
 → org.springframework.boot.SpringApplication#reportFailure(Ljava/util/Collection;Ljava/lang/Throwable;)V
 → org.springframework.boot.diagnostics.FailureAnalyzers#reportException(Ljava/lang/Throwable;)Z
 → org.springframework.boot.diagnostics.FailureAnalyzers#analyze(Ljava/lang/Throwable;Ljava/util/List;)Lorg/springframework/boot/diagnostics/FailureAnalysis;
 → org.springframework.boot.autoconfigure.diagnostics.analyzer.NoSuchBeanDefinitionFailureAnalyzer#analyze(Ljava/lang/Throwable;Lorg/springframework/beans/factory/NoSuchBeanDefinitionException;Ljava/lang/String;)Lorg/springframework/boot/diagnostics/FailureAnalysis;
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain follows a genuine Spring Boot startup flow (SpringApplication.run -> failure handling -> FailureAnalyzers -> NoSuchBeanDefinitionFailureAnalyzer -> AnnotationsScanner), and the vulnerable method hasSameGenericTypeParameters is actually invoked by processMethodHierarchy/isOverride during annotation scanning, exactly the code path the CVE describes. The only caveat is that reaching it requires a startup failure (handleRunFailure), but the calls themselves are real framework wiring, not CHA artifacts.


### CVE-2025-48988

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.core.StandardEngine#logAccess(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;JZ)V
 → org.apache.catalina.valves.AbstractAccessLogValve#log(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.valves.ExtendedAccessLogValve$RequestParameterElement#addElement(Ljava/io/CharArrayWriter;Ljava/util/Date;Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.connector.Request#getParameter(Ljava/lang/String;)Ljava/lang/String;
 → org.apache.catalina.connector.Request#parseParameters()V
 → org.apache.catalina.connector.Request#parseParts(Z)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain contains obviously bogus hops: Tomcat#getHost() 'calling' StandardHost#<init>(), and createWebServer() reaching LifecycleBase#stop()/Connector#stopInternal() during server creation — these are CHA over-approximation edges, not real framework behavior. While Request#parseParts is genuinely exercised at runtime during multipart uploads, this specific path (via endpoint stop, AjpProcessor recycle, and access-log valve parameter parsing) does not correspond to a real call sequence.


### CVE-2025-48989

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.AbstractProcessor#timeoutAsync(J)V
 → org.apache.coyote.AbstractProcessor#doTimeoutAsync()V
 → org.apache.coyote.http2.StreamProcessor#processSocketEvent(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.coyote.http2.StreamProcessor#process(Lorg/apache/tomcat/util/net/SocketEvent;)V
 → org.apache.coyote.http2.Stream#close(Lorg/apache/coyote/http2/Http2Exception;)V
 → org.apache.coyote.http2.Http2AsyncUpgradeHandler#sendStreamReset(Lorg/apache/coyote/http2/StreamException;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from Tomcat#getHost to StandardHost#<init> to StandardPipeline#setBasic and then to LifecycleBase#stop(), but setBasic never calls stop(); this is a CHA-style splice of unrelated lifecycle methods. The tail (AbstractProcessor#timeoutAsync -> 'StreamProcessor#process' -> Stream#close -> sendStreamReset) also doesn't correspond to real Tomcat HTTP/2 classes/call flow during server startup.


### CVE-2025-49125

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.loader.WebappClassLoaderBase#start()V
 → org.apache.catalina.webresources.StandardRoot#getResources(Ljava/lang/String;Z)[Lorg/apache/catalina/WebResource;
 → org.apache.catalina.webresources.StandardRoot#getResourcesInternal(Ljava/lang/String;Z)[Lorg/apache/catalina/WebResource;
 → org.apache.catalina.webresources.AbstractArchiveResourceSet#getResource(Ljava/lang/String;)Lorg/apache/catalina/WebResource;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot's embedded Tomcat startup genuinely initializes the webapp class loader and its StandardRoot resource sets (WebappClassLoaderBase#start → StandardRoot#getResources), and DirResourceSet#getResource is the concrete resource lookup invoked during that initialization, matching the CVE's affected methods.


### CVE-2025-49128

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is a real Spring Boot/Spring Data chain: a RepositoryPopulator bean (Jackson2ResourceReader) genuinely parses data.json resources via ObjectReader/JsonFactory during ContextRefreshedEvent, and ByteSourceJsonBootstrapper really constructs a ReaderBasedJsonParser for non-UTF8 input. The hops match actual framework wiring rather than generic interface dispatch, so reaching the listed vulnerable constructor is plausible.


### CVE-2025-52434

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Boot's embedded Tomcat setup (SpringApplication.run -> refresh -> ServletWebServerApplicationContext.createWebServer -> TomcatServletWebServerFactory.getWebServer) is a genuine startup flow, and LifecycleBase.stop -> Connector.stopInternal -> AbstractProtocol.stop -> AbstractEndpoint.stop -> AprEndpoint.stopInternal is the real Tomcat connector shutdown chain. Some middle hops (Tomcat#getHost -> StandardHost#<init> -> StandardPipeline#setBasic) look like CHA noise, but the overall sequence corresponds to actual framework behavior, so the vulnerable method is plausibly reachable when the APR connector is configured.


### CVE-2025-52520

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.core.StandardEngine#logAccess(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;JZ)V
 → org.apache.catalina.valves.AbstractAccessLogValve#log(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.valves.ExtendedAccessLogValve$RequestParameterElement#addElement(Ljava/io/CharArrayWriter;Ljava/util/Date;Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.connector.Request#getParameter(Ljava/lang/String;)Ljava/lang/String;
 → org.apache.catalina.connector.Request#parseParameters()V
 → org.apache.catalina.connector.Request#parseParts(Z)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot startup chain (steps 1-11) reaching TomcatServletWebServerFactory#getWebServer is a real embedded-Tomcat setup path, but the tail is nonsense: Tomcat#getHost() cannot call StandardHost#<init>, a pipeline setBasic() would not lead to LifecycleBase#stop(), and the stop/recycle chain hops across unrelated classes (AprEndpoint -> SocketWrapperBase#close -> ConnectionHandler#release -> AjpProcessor#recycle -> CoyoteAdapter#checkRecycled) purely via CHA over-approximation on generic virtual calls.


### CVE-2025-53506

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.http2.Http2Protocol#getProcessor(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/coyote/Adapter;)Lorg/apache/coyote/Processor;
 → org.apache.coyote.http2.Http2Protocol#getInternalUpgradeHandler(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/coyote/Adapter;Lorg/apache/coyote/Request;)Lorg/apache/coyote/http11/upgrade/InternalHttpUpgradeHandler;
 → org.apache.coyote.http2.Http2UpgradeHandler#<init>(Lorg/apache/coyote/http2/Http2Protocol;Lorg/apache/coyote/Adapter;Lorg/apache/coyote/Request;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The middle hops are CHA noise: Tomcat#getHost() would not construct a fresh StandardHost, StandardPipeline#setBasic(Valve) would never call LifecycleBase#stop(), and Connector#stopInternal() does not dispatch into UpgradeProcessorInternal#timeoutAsync — these are unrelated virtual-call targets strung together. Only the final segment (Http2Protocol#getInternalUpgradeHandler creating an Http2UpgradeHandler) is a genuine Tomcat HTTP/2 flow, but the path as a whole does not correspond to a real call chain.


### CVE-2025-55668

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.connector.Request#recycle()V
 → org.apache.catalina.connector.Request#recycleSessionInfo()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain from Tomcat#getHost to StandardHost#<init> to StandardPipeline#setBasic to LifecycleBase#stop is a nonsensical sequence (a constructor and a setter do not lead to lifecycle stop), and the subsequent Connector/AbstractProtocol/AprEndpoint stop path then jumps into AjpProcessor#recycle and CoyoteAdapter#log, which are unrelated runtime request-processing components, not startup code. These hops are classic CHA over-approximation edges (e.g., virtual dispatch on Valve/Lifecycle/Processor types) with no real call relationship.


### CVE-2025-55752

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#start()V
 → org.apache.catalina.valves.rewrite.RewriteValve#startInternal()V
 → org.apache.catalina.valves.rewrite.RewriteValve#parse(Ljava/io/BufferedReader;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The chain Spring Boot main -> embedded Tomcat -> TomcatServletWebServerFactory#configureEngine -> StandardPipeline#addValve -> LifecycleBase#start -> RewriteValve#startInternal -> parse is a genuine framework lifecycle: startInternal really does read rewrite rules via parse(BufferedReader). The only caveat is that RewriteValve is only added if the app actually configures URL rewriting, but the call edges themselves are real, not CHA artifacts.


### CVE-2025-61795

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.ajp.AjpProcessor#recycle()V
 → org.apache.catalina.connector.CoyoteAdapter#checkRecycled(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;)V
 → org.apache.catalina.connector.CoyoteAdapter#log(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;J)V
 → org.apache.catalina.core.StandardEngine#logAccess(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;JZ)V
 → org.apache.catalina.valves.AbstractAccessLogValve#log(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.valves.ExtendedAccessLogValve$RequestParameterElement#addElement(Ljava/io/CharArrayWriter;Ljava/util/Date;Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;J)V
 → org.apache.catalina.connector.Request#getParameter(Ljava/lang/String;)Ljava/lang/String;
 → org.apache.catalina.connector.Request#parseParameters()V
 → org.apache.catalina.connector.Request#parseParts(Z)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from StandardHost#<init> to StandardPipeline#setBasic to LifecycleBase#stop, i.e., a constructor call flowing into a lifecycle shutdown, which is not a real call relationship. The rest of the path (endpoint stop -> ConnectionHandler#release -> AjpProcessor#recycle -> access-log valve -> Request#parseParts) is a chain of unrelated shutdown/recycling/logging methods stitched together by over-approximate virtual dispatch, not a genuine route to multipart parsing.


### CVE-2025-66614

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#configureEngine(Lorg/apache/catalina/Engine;)V
 → org.apache.catalina.core.StandardPipeline#addValve(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#start()V
 → org.apache.catalina.util.LifecycleBase#init()V
 → org.apache.catalina.util.LifecycleMBeanBase#initInternal()V
 → org.apache.catalina.authenticator.AuthenticatorBase#register(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;Ljava/security/Principal;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;ZZ)V
 → org.apache.catalina.connector.Request#getAttribute(Ljava/lang/String;)Ljava/lang/Object;
 → org.apache.coyote.Request#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.AbstractProcessor#action(Lorg/apache/coyote/ActionCode;Ljava/lang/Object;)V
 → org.apache.coyote.http11.Http11Processor#sslReHandShake()V
 → org.apache.tomcat.util.net.Nio2Endpoint$Nio2SocketWrapper#doClientAuth(Lorg/apache/tomcat/util/net/SSLSupport;)V
 → org.apache.tomcat.util.net.SecureNio2Channel#rehandshake()V
 → org.apache.tomcat.util.net.SecureNio2Channel#handshakeInternal(Z)I
 → org.apache.tomcat.util.net.SecureNio2Channel#processSNI()I
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The bridge from Tomcat startup code (StandardPipeline#addValve → LifecycleBase#start/init) into AuthenticatorBase#register is implausible: register() is a per-request authentication bookkeeping method, not something invoked during component initialization, so that hop exists only via over-broad virtual dispatch. The tail (doClientAuth → rehandshake → processSNI) is a genuine Tomcat chain, but the path as a whole never reaches it from this entry sequence.


### CVE-2026-19032

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — Node 6 is a signature-mismatched hop (DefaultBootstrapContext#close takes no arguments, yet is shown receiving a ConfigurableApplicationContext), and the tail is implausible: Jackson2ResourceReader uses ObjectReader#readTree, which binds to JsonNode and never invokes NioPathDeserializer (a typed Path deserializer only runs for POJO fields, not tree binding).


### CVE-2026-24880

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.tomcat.util.net.AbstractEndpoint#stop()V
 → org.apache.tomcat.util.net.AprEndpoint#stopInternal()V
 → org.apache.tomcat.util.net.SocketWrapperBase#close()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#release(Lorg/apache/coyote/Processor;)V
 → org.apache.coyote.http11.Http11Processor#recycle()V
 → org.apache.coyote.http11.Http11InputBuffer#recycle()V
 → org.apache.coyote.http11.filters.ChunkedInputFilter#recycle()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from Tomcat#getHost to StandardHost#<init> (getHost does not construct a host) and then into LifecycleBase#stop / Connector#stopInternal / endpoint stop during server *creation*, which is backwards — shutdown/recycle code is not invoked from createWebServer. The real route to ChunkedInputFilter#recycle is via runtime request processing, so this specific sequence is a CHA over-approximation stitching unrelated lifecycle methods together.


### CVE-2026-25854

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.valves.LoadBalancerDrainingValve#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop StandardHost#<init>() -> StandardPipeline#setBasic -> LifecycleBase#stop is nonsensical (a constructor leading to pipeline setup then to lifecycle stop), a classic CHA over-approximation; additionally LoadBalancerDrainingValve#invoke is only reached if that valve is explicitly configured, which Spring Boot's embedded Tomcat does not do by default.


### CVE-2026-41293

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.AbstractProcessor#timeoutAsync(J)V
 → org.apache.coyote.AbstractProcessor#doTimeoutAsync()V
 → org.apache.coyote.http2.StreamProcessor#processSocketEvent(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.coyote.http2.StreamProcessor#process(Lorg/apache/tomcat/util/net/SocketEvent;)V
 → org.apache.coyote.http2.StreamException#<init>(Ljava/lang/String;Lorg/apache/coyote/http2/Http2Error;I)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from StandardPipeline#setBasic to LifecycleBase#stop and then into Connector#stopInternal → AbstractProcessor#timeoutAsync → StreamProcessor#process, which is incoherent: setBasic does not call stop, and a stop/timeout path would never flow into StreamException construction during server startup. These are CHA-style virtual-dispatch edges connecting unrelated lifecycle and request-processing code, not a real Tomcat call sequence.


### CVE-2026-41848

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplication#load(Lorg/springframework/context/ApplicationContext;[Ljava/lang/Object;)V
 → org.springframework.boot.SpringApplication#createBeanDefinitionLoader(Lorg/springframework/beans/factory/support/BeanDefinitionRegistry;[Ljava/lang/Object;)Lorg/springframework/boot/BeanDefinitionLoader;
 → org.springframework.boot.BeanDefinitionLoader#<init>(Lorg/springframework/beans/factory/support/BeanDefinitionRegistry;[Ljava/lang/Object;)V
 → org.springframework.beans.factory.xml.XmlBeanDefinitionReader#<init>(Lorg/springframework/beans/factory/support/BeanDefinitionRegistry;)V
 → org.springframework.beans.factory.support.AbstractBeanDefinitionReader#<init>(Lorg/springframework/beans/factory/support/BeanDefinitionRegistry;)V
 → org.springframework.core.io.support.PathMatchingResourcePatternResolver#<init>()V
 → org.springframework.util.AntPathMatcher#<init>()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — PathMatchingResourcePatternResolver genuinely instantiates AntPathMatcher in its constructor, and it is plausibly created via AbstractBeanDefinitionReader/XmlBeanDefinitionReader during BeanDefinitionLoader setup in SpringApplication.run. The chain matches real Spring Boot bootstrap wiring, though it only reaches the AntPathMatcher constructor (setup), not the vulnerable match methods with attacker input.


### CVE-2026-41851

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.context.support.AbstractApplicationContext#prepareBeanFactory(Lorg/springframework/beans/factory/config/ConfigurableListableBeanFactory;)V
 → org.springframework.context.expression.StandardBeanExpressionResolver#<init>(Ljava/lang/ClassLoader;)V
 → org.springframework.expression.spel.SpelParserConfiguration#<init>(Lorg/springframework/expression/spel/SpelCompilerMode;Ljava/lang/ClassLoader;)V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — AbstractApplicationContext#refresh really does call prepareBeanFactory, which instantiates a StandardBeanExpressionResolver whose constructor creates a SpelParserConfiguration with a compiler mode and classloader — this is genuine Spring context bootstrap wiring, not a CHA artifact.


### CVE-2026-43512

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.authenticator.AuthenticatorBase#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.authenticator.BasicAuthenticator#doAuthenticate(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;)Z
 → org.apache.catalina.realm.RealmBase#authenticate(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Ljava/security/Principal;
 → org.apache.catalina.realm.RealmBase#getDigest(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain jumps from server startup (createWebServer -> Tomcat#getHost -> StandardHost#<init> -> setBasic) directly into LifecycleBase#stop and then into runtime request processing (processSocket, CoyoteAdapter#asyncDispatch), which cannot be a single synchronous call chain — startup never executes a request dispatch inline. Additionally the path reaches the digest-auth vulnerable RealmBase#getDigest via BasicAuthenticator#doAuthenticate, which uses a different authenticate overload and never calls getDigest, a classic CHA over-approximation hop.


### CVE-2026-43513

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.authenticator.AuthenticatorBase#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.authenticator.BasicAuthenticator#doAuthenticate(Lorg/apache/catalina/connector/Request;Ljavax/servlet/http/HttpServletResponse;)Z
 → org.apache.catalina.realm.LockOutRealm#authenticate(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Ljava/security/Principal;
 → org.apache.catalina.realm.LockOutRealm#filterLockedAccounts(Ljava/lang/String;Ljava/security/Principal;)Ljava/security/Principal;
 → org.apache.catalina.realm.LockOutRealm#isLocked(Ljava/lang/String;)Z
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from StandardHost#<init> to StandardPipeline#setBasic to LifecycleBase#stop() to Connector#stopInternal() during application startup, which is nonsensical — a constructor path does not lead to a lifecycle stop, and the subsequent chain through UpgradeProcessorInternal/Http2UpgradeHandler/AprEndpoint$SocketProcessor is a grab-bag of unrelated runtime dispatch targets (APR endpoint, HTTP/2 upgrade) that would never be invoked from a startup refresh. The final segment (BasicAuthenticator -> LockOutRealm#authenticate -> isLocked) is plausible in isolation, but the middle of the path is a classic CHA over-approximation stitching together unrelated virtual-call targets.


### CVE-2026-43514

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.ajp.AjpProcessor#service(Lorg/apache/tomcat/util/net/SocketWrapperBase;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.ajp.AjpProcessor#prepareRequest()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — While Spring Boot starting an embedded Tomcat (nodes 1-11) is plausible, the chain StandardPipeline#setBasic -> LifecycleBase#stop -> Connector#stopInternal -> AbstractProtocol#stop -> UpgradeProcessorInternal#timeoutAsync is a CHA artifact: setBasic does not call stop, and AbstractProtocol#stop would not dispatch into an UpgradeProcessorInternal#timeoutAsync. The path splices unrelated Tomcat lifecycle/processing methods together via over-broad virtual dispatch to reach AjpProcessor#prepareRequest.


### CVE-2026-43515

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.authenticator.AuthenticatorBase#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.realm.RealmBase#findSecurityConstraints(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/Context;)[Lorg/apache/tomcat/util/descriptor/web/SecurityConstraint;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The chain hops from StandardHost#<init> to StandardPipeline#setBasic to LifecycleBase#stop and then into Connector#stopInternal → timeoutAsync → processSocket are classic CHA artifacts: a constructor/setter does not lead to lifecycle stop, and the runtime request-processing chain (SocketProcessor → CoyoteAdapter → AuthenticatorBase#invoke → RealmBase#findSecurityConstraints) is driven by live requests, not reachable from SpringApplication#run at startup.


### CVE-2026-47884

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.SpringApplicationRunListeners#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplicationRunListeners#lambda$contextLoaded$4(Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/boot/SpringApplicationRunListener;)V
 → org.springframework.boot.context.event.EventPublishingRunListener#contextLoaded(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.ApplicationObjectSupport#setApplicationContext(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.web.servlet.view.xslt.XsltView#initApplicationContext()V
 → org.springframework.web.servlet.view.xslt.XsltView#loadTemplates()Ljavax/xml/transform/Templates;
 → org.springframework.web.servlet.view.xslt.XsltView#getStylesheetSource()Ljavax/xml/transform/Source;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop from EventPublishingRunListener#contextLoaded to ApplicationObjectSupport#setApplicationContext is a CHA artifact: EventPublishingRunListener is a SpringApplicationRunListener, not an ApplicationObjectSupport, and nothing in the startup path would call setApplicationContext on an XsltView. Additionally, XsltView#loadTemplates/getStylesheetSource only run during actual view rendering of a request, not during SpringApplication.run context preparation, so this chain does not correspond to real framework behavior.


### CVE-2026-49844

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.StartupInfoLogger#logStarted(Lorg/apache/commons/logging/Log;Lorg/springframework/util/StopWatch;)V
 → org.apache.commons.logging.LogAdapter$Log4jLog#info(Ljava/lang/Object;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — Spring Boot's StartupInfoLogger logs a plain String/SimpleMessage ('Started application in N seconds'), not a MapMessage, so the hop SimpleLogger#logMessage -> MapMessage#getFormattedMessage is a CHA over-approximation over the Message interface. Additionally, SimpleLogger does not use JsonTemplateLayout or the JSON format array, so the vulnerable formatting path would not be triggered even if a MapMessage were logged.


### CVE-2026-65182

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.connector.Connector#stopInternal()V
 → org.apache.coyote.AbstractProtocol#stop()V
 → org.apache.coyote.http11.upgrade.UpgradeProcessorInternal#timeoutAsync(J)V
 → org.apache.coyote.http2.Http2UpgradeHandler#timeoutAsync(J)V
 → org.apache.tomcat.util.net.SocketWrapperBase#processSocket(Lorg/apache/tomcat/util/net/SocketEvent;Z)V
 → org.apache.tomcat.util.net.AbstractEndpoint#processSocket(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;Z)Z
 → org.apache.tomcat.util.net.SocketProcessorBase#run()V
 → org.apache.tomcat.util.net.AprEndpoint$SocketProcessor#doRun()V
 → org.apache.coyote.AbstractProtocol$ConnectionHandler#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessorLight#process(Lorg/apache/tomcat/util/net/SocketWrapperBase;Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.coyote.AbstractProcessor#dispatch(Lorg/apache/tomcat/util/net/SocketEvent;)Lorg/apache/tomcat/util/net/AbstractEndpoint$Handler$SocketState;
 → org.apache.catalina.connector.CoyoteAdapter#asyncDispatch(Lorg/apache/coyote/Request;Lorg/apache/coyote/Response;Lorg/apache/tomcat/util/net/SocketEvent;)Z
 → org.apache.catalina.authenticator.AuthenticatorBase#invoke(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/connector/Response;)V
 → org.apache.catalina.realm.RealmBase#findSecurityConstraints(Lorg/apache/catalina/connector/Request;Lorg/apache/catalina/Context;)[Lorg/apache/tomcat/util/descriptor/web/SecurityConstraint;
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The hop StandardPipeline#setBasic -> LifecycleBase#stop and then into Connector#stopInternal -> UpgradeProcessorInternal#timeoutAsync splices unrelated lifecycle and request-processing code together via CHA; a real request path (SocketProcessor -> CoyoteAdapter -> AuthenticatorBase -> RealmBase#findSecurityConstraints) exists at runtime but is not connected by this startup sequence.


### CVE-2026-68525

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#refreshContext(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ApplicationContext;)V
 → org.springframework.boot.SpringApplication#refresh(Lorg/springframework/context/ConfigurableApplicationContext;)V
 → org.springframework.context.support.AbstractApplicationContext#refresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#onRefresh()V
 → org.springframework.boot.web.servlet.context.ServletWebServerApplicationContext#createWebServer()V
 → org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory#getWebServer([Lorg/springframework/boot/web/servlet/ServletContextInitializer;)Lorg/springframework/boot/web/server/WebServer;
 → org.apache.catalina.startup.Tomcat#getHost()Lorg/apache/catalina/Host;
 → org.apache.catalina.core.StandardHost#<init>()V
 → org.apache.catalina.core.StandardPipeline#setBasic(Lorg/apache/catalina/Valve;)V
 → org.apache.catalina.util.LifecycleBase#stop()V
 → org.apache.catalina.util.LifecycleBase#fireLifecycleEvent(Ljava/lang/String;Ljava/lang/Object;)V
 → org.apache.catalina.startup.Tomcat$FixContextListener#lifecycleEvent(Lorg/apache/catalina/LifecycleEvent;)V
 → org.apache.catalina.authenticator.NonLoginAuthenticator#<init>()V
 → org.apache.catalina.authenticator.AuthenticatorBase#<init>()V
```

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Tomcat's embedded-mode FixContextListener really does install a NonLoginAuthenticator during lifecycleEvent when a context lacks a login config, so reaching AuthenticatorBase#<init> during Spring Boot's embedded Tomcat startup is genuine framework wiring. The StandardPipeline#setBasic -> LifecycleBase#stop hop (13-15) looks like a CHA artifact, but it only shortcuts the path to fireLifecycleEvent, which is otherwise a real lifecycle mechanism, so the endpoint remains believable.


### CVE-2026-77310

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — The path follows a genuine Spring Data repository-populator flow: ContextRefreshedEvent -> AbstractRepositoryPopulatorFactoryBean -> Jackson2ResourceReader.readFrom -> ObjectReader.readTree, which is exactly how Spring Data initializes repositories from data.json. The only weak link is readTree deserializing into JsonNode rather than a typed InetAddress target, so reaching FromStringDeserializer._deserialize would require the JSON to contain a typed InetAddress value, but the wiring itself is real framework behavior.


### CVE-2026-89425

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** artifact — The Spring Boot startup chain through ResourceReaderRepositoryPopulator/Jackson2ResourceReader#readFrom is a real framework flow, but it calls ObjectReader#readTree(InputStream), which constructs a UTF8StreamJsonParser — the implementation the advisory explicitly confirms is correctly bounded. The hop to UTF8DataInputJsonParser (only created via createParser(DataInput)) is a CHA artifact: this call site never dispatches to the DataInput-backed parser, so the vulnerable code is not actually reached.


### CVE-2026-91776

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — This is a real Spring Data code path: ResourceReaderRepositoryPopulator/Jackson2ResourceReader genuinely uses Jackson's ObjectReader (with default typing enabled) to read repository initializer JSON, so the chain down to AsArrayTypeDeserializer._findDeserializer is a genuine framework flow, not a CHA artifact. Note the data source is classpath resources rather than attacker input, so exploitability is questionable, but the call sequence itself is real.


### CVE-2026-91777

```
org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V
 → org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#run([Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;
 → org.springframework.boot.SpringApplication#prepareContext(Lorg/springframework/boot/DefaultBootstrapContext;Lorg/springframework/context/ConfigurableApplicationContext;Lorg/springframework/core/env/ConfigurableEnvironment;Lorg/springframework/boot/SpringApplicationRunListeners;Lorg/springframework/boot/ApplicationArguments;Lorg/springframework/boot/Banner;)V
 → org.springframework.boot.DefaultBootstrapContext#close(Lorg/springframework/context/ConfigurableApplicationContext;)V
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

**Entry call (app → dependency):** `org.springframework.samples.petclinic.PetClinicApplication#main([Ljava/lang/String;)V` → `org.springframework.boot.SpringApplication#run(Ljava/lang/Class;[Ljava/lang/String;)Lorg/springframework/context/ConfigurableApplicationContext;`


**Path plausibility (z-ai/glm-5.3):** plausible — Spring Petclinic actually configures a RepositoryPopulator (data.json) via Jackson2ResourceReader, so the chain SpringApplication.run → ContextRefreshedEvent → AbstractRepositoryPopulatorFactoryBean → Jackson2ResourceReader → ObjectReader → MapDeserializer is a genuine framework wiring path. The only caveat is that the vulnerable forward-reference methods require @JsonIdentityInfo content in the populated JSON, but the call chain itself is real, not a CHA artifact.




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

- class javax.activation.CommandInfo$Beans$1 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParseException found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandObject found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap$1 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSourceDataContentHandler found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.FileDataSource found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandMap found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$5 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeTypeParameterList found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$3 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap$1 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.UnsupportedDataTypeException found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.CommandInfo$Beans found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.FileTypeMap found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimetypesFileTypeMap found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandlerDataSource found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataSource found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MailcapCommandMap found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$1 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.ActivationDataFlavor found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$2 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport$4 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.ObjectDataContentHandler found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.SecuritySupport found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandlerFactory found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.MimeType found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataContentHandler found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.URLDataSource found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen
- class javax.activation.DataHandler$1 found in both work/spring-petclinic/target/dependency/jakarta.activation-1.2.2.jar and work/spring-petclinic/target/dependency/jakarta.activation-api-1.2.2.jar; keeping the first one seen


