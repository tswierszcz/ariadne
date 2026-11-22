package fx.web;

import fx.trace.Ran;

public class BoomException extends RuntimeException {
    public BoomException() { Ran.stack(); }
}
