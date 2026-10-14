package fixture;

/**
 * A CDI-observer-method stand-in: {@link #onEvent}'s first parameter
 * carries {@link Param}, a parameter-level annotation, while its second
 * parameter carries none — exercising both "has one of these" and "has
 * none" in the same RuntimeVisibleParameterAnnotations attribute.
 */
public class Observer {
    public void onEvent(@Param String event, String unused) {
    }
}
