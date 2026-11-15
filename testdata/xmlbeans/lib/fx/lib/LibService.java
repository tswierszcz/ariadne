package fx.lib;

import fx.trace.Ran;

/**
 * A library class (deps/fx-lib.jar) declared as a bean in the
 * application's XML — like JPetStore's SqlSessionFactoryBean.
 */
public class LibService {
    public LibService() { Ran.stack(); }
    public void setEndpoint(String endpoint) { Ran.stack(); }
    public void open() { Ran.stack(); }
    public void close() { Ran.stack(); }
}
