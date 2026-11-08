package cg;

/** One call site per method: the oracle drives each with every receiver the fixture can create. */
public final class Sites {
    static void iface(Task t) { t.run(); }

    static void overloadInt(Over o) { o.f(1); }

    static void overloadString(Over o) { o.f("x"); }

    static void defaultMethod(Greeter g) { g.hello(); }

    @SuppressWarnings({"unchecked", "rawtypes"})
    static void bridge(Fn f) { f.apply("x"); }

    static void covariant(Animal a) { a.self(); }

    static void inherited(Sub s) { s.hook(); }

    static void superCall(Child c) { c.callSuper(); }
}
