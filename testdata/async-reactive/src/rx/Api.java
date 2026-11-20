package rx;

import java.util.concurrent.SubmissionPublisher;

public class Api {
    /** Never called: Later is subscribed nowhere that runs. */
    static void later(SubmissionPublisher<String> pub) {
        pub.subscribe(new Later());
    }
}
