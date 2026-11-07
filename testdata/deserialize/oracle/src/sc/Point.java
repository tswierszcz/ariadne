package sc;

import java.io.Serializable;

/** A record: serialization runs its canonical constructor (not modeled). */
public record Point(int x) implements Serializable {
    public Point {
        Rec.ran("sc.Point#<init>");
    }
}
