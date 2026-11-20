package fx;

import com.fasterxml.jackson.annotation.JsonIdentityInfo;
import com.fasterxml.jackson.annotation.JsonIdentityReference;
import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;
import com.fasterxml.jackson.annotation.JsonUnwrapped;
import com.fasterxml.jackson.annotation.ObjectIdGenerators;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.dataformat.yaml.YAMLMapper;
import java.net.InetSocketAddress;
import java.nio.file.Path;
import java.util.Date;
import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * LIBRARY_MODE LB4 oracle for jackson-dataformat-yaml 2.12.3: a user of
 * YAMLMapper (the library's API, with the ObjectMapper methods it
 * inherits) reads YAML into maps and POJOs. Each scenario is independent;
 * an exception is printed and the next runs. Run by
 * scripts/library-lb4-jvm-oracle.sh.
 */
public class YamlClient {
    public enum Color { RED, GREEN }

    public static class Settings {
        public Color color;
        public Date when;
        public Path path;
        public UUID id;
        public InetSocketAddress address;
    }

    public static class Name {
        public String first;
        public String last;
    }

    public static class Person {
        @JsonUnwrapped
        public Name name;
        public int age;
    }

    @JsonTypeInfo(use = JsonTypeInfo.Id.NAME, property = "type")
    @JsonSubTypes({@JsonSubTypes.Type(value = Circle.class, name = "circle")})
    public static class Shape {
    }

    public static class Circle extends Shape {
        public int r;
    }

    @JsonIdentityInfo(generator = ObjectIdGenerators.PropertyGenerator.class, property = "id")
    public static class Node {
        public String id;
        @JsonIdentityReference(alwaysAsId = true)
        public List<Node> links;
    }

    public static class Graph {
        public List<Node> refs;
        public List<Node> nodes;
    }

    public static class Point {
        public int x;
    }

    public static class Holder {
        @JsonTypeInfo(use = JsonTypeInfo.Id.CLASS, property = "@class")
        public Object value;
    }

    static void run(String name, Runnable r) {
        try {
            r.run();
            System.out.println(name + ": ok");
        } catch (RuntimeException e) {
            System.out.println(name + ": " + e.getClass().getSimpleName());
        }
    }

    interface Body {
        Object call() throws Exception;
    }

    static Runnable body(Body b) {
        return () -> {
            try {
                b.call();
            } catch (Exception e) {
                throw new RuntimeException(e.getClass().getSimpleName() + ": " + e.getMessage(), e);
            }
        };
    }

    public static void main(String[] args) {
        YAMLMapper m = new YAMLMapper();
        run("map", body(() -> m.readValue("a: [1, [2, 3]]\nb: {c: d}\n", Map.class)));
        run("pojo", body(() -> m.readValue(
                "color: GREEN\nwhen: 2020-01-02T03:04:05.000+00:00\npath: /tmp/x\nid: 123e4567-e89b-12d3-a456-426614174000\naddress: localhost:8080\n",
                Settings.class)));
        run("unwrapped", body(() -> m.readValue("first: A\nlast: B\nage: 3\n", Person.class)));
        run("polymorphic", body(() -> m.readValue("type: circle\nr: 2\n", Shape.class)));
        run("forward-reference", body(() -> m.readValue("refs: [b]\nnodes: [{id: a, links: [b]}, {id: b, links: []}]\n", Graph.class)));
        YAMLMapper unwrap = new YAMLMapper();
        unwrap.enable(DeserializationFeature.UNWRAP_SINGLE_VALUE_ARRAYS);
        run("unwrap-single-value-array", body(() -> unwrap.readValue("[{x: 1}]\n", Point.class)));
        run("generic-class-id", body(() -> m.readValue("value: {'@class': 'java.util.ArrayList<java.lang.String>'}\n", Holder.class)));
        YAMLMapper typed = new YAMLMapper();
        typed.activateDefaultTyping(com.fasterxml.jackson.databind.jsontype.BasicPolymorphicTypeValidator.builder()
                .allowIfSubTypeIsArray().allowIfBaseType(Object.class).build());
        run("default-typing-basic-validator", body(() -> typed.readValue("['java.util.ArrayList', ['a']]\n", Object.class)));
        YAMLMapper limited = new YAMLMapper();
        limited.activateDefaultTyping(new com.fasterxml.jackson.databind.jsontype.DefaultBaseTypeLimitingValidator());
        run("default-typing-limiting-validator", body(() -> limited.readValue("['java.util.ArrayList', ['a']]\n", Object.class)));
        run("type-mismatch", body(() -> m.readValue("x: [oops]\n", Point.class)));
    }
}
