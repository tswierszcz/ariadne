package al;

public class Main {
    public static void main(String[] args) throws Exception {
        Bus bus = new Bus();
        bus.register(new A());
        bus.register(e -> Work.lam(e));
        new U(); // created, never registered
        Thread t = new Thread(() -> bus.publish("go"));
        t.start();
        t.join();

        IndexedBus ib = new IndexedBus();
        ib.register(new B());
        ib.publish("indexed");

        new SilentBus().register(new C());

        Model m = new Model();
        m.addListener(new View());
        m.setX(1);
        new QuietModel().addListener(new Unseen());
    }
}
