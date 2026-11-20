package ar;

/** Like Jetty's ChannelEndPoint: returns one of its final task fields. */
public class Conn implements Endpoint {
    private final Runnable fill = new Runnable() {
        @Override
        public void run() {
            Work.fill();
        }
    };
    private final Runnable write = () -> Work.write();
    // Returned only to code that never runs it (Main#main).
    private final Runnable close = new CloseTask();
    private boolean flip;

    @Override
    public Runnable onSelected() {
        flip = !flip;
        return flip ? fill : write;
    }

    Runnable closer() {
        return close;
    }
}
