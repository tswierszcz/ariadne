package al;

public class B implements Listener {
    @Override
    public void onEvent(String e) {
        Work.hit("B");
    }
}
