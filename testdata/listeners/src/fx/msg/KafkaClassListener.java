package fx.msg;

import org.springframework.kafka.annotation.KafkaHandler;
import org.springframework.kafka.annotation.KafkaListener;

/** A class-level listener dispatching to @KafkaHandler methods by payload type. */
@KafkaListener("mixed")
public class KafkaClassListener {
    @KafkaHandler
    public void onString(String s) {}

    @KafkaHandler(isDefault = true)
    public void onOther(Object o) {}
}
