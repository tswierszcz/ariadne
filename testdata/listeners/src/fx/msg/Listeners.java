package fx.msg;

import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.jms.annotation.JmsListener;
import org.springframework.kafka.annotation.KafkaListener;

/** Method-level listeners, one with two (repeated) annotations. */
public class Listeners {
    @JmsListener("orders")
    public void onJms(String m) {}

    @JmsListener("a")
    @JmsListener("b")
    public void onTwoQueues(String m) {}

    @KafkaListener("events")
    public void onKafka(String m) {}

    @KafkaListener("x")
    @KafkaListener("y")
    public void onTwoTopics(String m) {}

    @RabbitListener("jobs")
    public void onRabbit(String m) {}

    /** No listener annotation: no root. */
    public void helper(String m) {}
}
