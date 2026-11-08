package rt;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

/** Classes created by a factory the analysis cannot resolve (the class name is built at run time). */
@Retention(RetentionPolicy.RUNTIME)
@interface Scanned {
}

/** Annotated: within the unknown-allocation fallback (what frameworks scan for). */
@Scanned
final class Dynamic implements Shape {
    public double area() {
        Ran.here();
        return 7;
    }
}

/** Named nowhere the analysis can see: outside the fallback, covered only by the active gap. */
final class Unnamed implements Shape {
    public double area() {
        Ran.here();
        return 8;
    }
}

final class UnknownFactory {
    static Shape create(String simpleName) throws ReflectiveOperationException {
        String cn = new StringBuilder("rt").append('.').append(simpleName).toString();
        return (Shape) Class.forName(cn).getDeclaredConstructor().newInstance();
    }
}
