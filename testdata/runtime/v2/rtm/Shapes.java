package rtm;

interface Shape {
    long area(long n);
}

final class Square implements Shape {
    public long area(long n) {
        long s = 0;
        for (long i = 0; i < n; i++) {
            s += i * i % 11; // v2: another build of the same class
        }
        return s;
    }
}

/** A CHA target that never runs. */
final class Circle implements Shape {
    public long area(long n) {
        return -n;
    }
}
