package lib.api;

import org.apache.commons.text.StringSubstitutor;

/** (d) A package-private Listener nothing creates: CHA still offers it at Pipeline#run. */
class ExpandingListener implements Listener {
    public void onEvent(String text) {
        System.out.println(StringSubstitutor.createInterpolator().replace(text));
    }
}
