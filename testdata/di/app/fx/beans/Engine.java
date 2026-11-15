package fx.beans;

import fx.trace.Ran;

/** Created by a @Bean method; start/stop are its initMethod/destroyMethod. */
public class Engine {
    public void start() { Ran.stack(); }
    public void stop() { Ran.stack(); }
}
