package fx;

import io.jsonwebtoken.jackson.io.JacksonDeserializer;
import io.jsonwebtoken.jackson.io.JacksonSerializer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

/**
 * LIBRARY_MODE LB4 oracle for jjwt-jackson 0.11.2: what jjwt does with
 * the library — serialize a claims map and deserialize JWT header/body
 * JSON (also with custom claim types, the library's public constructor
 * for that) — called directly on JacksonSerializer/JacksonDeserializer.
 * Run by scripts/library-lb4-jvm-oracle.sh.
 */
public class JwtJsonClient {
    public static class Profile {
        public String name;
        public Path home;
        public UUID id;
        public java.net.InetSocketAddress address;
        public java.util.Locale locale;
    }

    public static class Name {
        public String first;
        public String last;
    }

    public static class Person {
        @com.fasterxml.jackson.annotation.JsonUnwrapped
        public Name name;
        public int age;
    }

    @com.fasterxml.jackson.annotation.JsonTypeInfo(use = com.fasterxml.jackson.annotation.JsonTypeInfo.Id.NAME, property = "type")
    @com.fasterxml.jackson.annotation.JsonSubTypes({@com.fasterxml.jackson.annotation.JsonSubTypes.Type(value = Circle.class, name = "circle")})
    public static class Shape {
    }

    public static class Circle extends Shape {
        public int r;
    }

    @com.fasterxml.jackson.annotation.JsonIdentityInfo(generator = com.fasterxml.jackson.annotation.ObjectIdGenerators.PropertyGenerator.class, property = "id")
    public static class Node {
        public String id;
        @com.fasterxml.jackson.annotation.JsonIdentityReference(alwaysAsId = true)
        public java.util.List<Node> links;
    }

    public static class Graph {
        public java.util.List<Node> refs;
        public java.util.List<Node> nodes;
    }

    static void run(String name, Runnable r) {
        try {
            r.run();
            System.out.println(name + ": ok");
        } catch (RuntimeException e) {
            System.out.println(name + ": " + e.getClass().getSimpleName());
        }
    }

    public static void main(String[] args) {
        run("serialize", () -> {
            Map<String, Object> claims = new HashMap<>();
            claims.put("sub", "joe");
            claims.put("roles", new String[] {"a", "b"});
            new JacksonSerializer<Map<String, Object>>().serialize(claims);
        });
        run("deserialize", () -> new JacksonDeserializer<Map<String, ?>>()
                .deserialize("{\"sub\":\"joe\",\"aud\":[\"a\",[\"b\"]],\"exp\":1}".getBytes(StandardCharsets.UTF_8)));
        run("claim-types", () -> {
            Map<String, Class> types = new HashMap<>();
            types.put("profile", Profile.class);
            new JacksonDeserializer<Map<String, ?>>(types)
                    .deserialize("{\"profile\":{\"name\":\"n\",\"home\":\"/tmp\",\"id\":\"123e4567-e89b-12d3-a456-426614174000\"}}".getBytes(StandardCharsets.UTF_8));
        });
        run("claim-types-more", () -> {
            Map<String, Class> types = new HashMap<>();
            types.put("person", Person.class);
            types.put("shape", Shape.class);
            types.put("graph", Graph.class);
            types.put("doc", org.w3c.dom.Document.class);
            types.put("profile", Profile.class);
            new JacksonDeserializer<Map<String, ?>>(types).deserialize(("{\"person\":{\"first\":\"A\",\"last\":\"B\",\"age\":3},"
                    + "\"shape\":{\"type\":\"circle\",\"r\":2},"
                    + "\"graph\":{\"refs\":[\"b\"],\"nodes\":[{\"id\":\"a\",\"links\":[\"b\"]},{\"id\":\"b\",\"links\":[]}]},"
                    + "\"doc\":\"<a/>\","
                    + "\"profile\":{\"address\":\"localhost:8080\",\"locale\":\"pl_PL\"}}").getBytes(StandardCharsets.UTF_8));
        });
        run("malformed", () -> new JacksonDeserializer<Map<String, ?>>()
                .deserialize("{\"sub\": tru}".getBytes(StandardCharsets.UTF_8)));
    }
}
