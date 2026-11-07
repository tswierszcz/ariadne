package cg;

/** Drives every site with every receiver the fixture can create; prints site TAB method that ran. */
public final class Oracle {
    static void at(String site, Runnable r) {
        Trace.site = site;
        r.run();
    }

    public static void main(String[] args) {
        for (Task t : new Task[] {new TaskA(), new TaskB(), new TaskC()}) at("iface", () -> Sites.iface(t));
        for (Over o : new Over[] {new Over(), new OverSub()}) {
            at("overloadInt", () -> Sites.overloadInt(o));
            at("overloadString", () -> Sites.overloadString(o));
        }
        for (Greeter g : new Greeter[] {new PlainGreeter(), new LoudGreeter()}) at("defaultMethod", () -> Sites.defaultMethod(g));
        at("bridge", () -> Sites.bridge(new StrFn()));
        for (Animal a : new Animal[] {new Animal(), new Dog()}) at("covariant", () -> Sites.covariant(a));
        at("inherited", () -> Sites.inherited(new Sub()));
        at("superCall", () -> Sites.superCall(new Child()));
        for (String h : Trace.hits) System.out.println(h);
    }
}
