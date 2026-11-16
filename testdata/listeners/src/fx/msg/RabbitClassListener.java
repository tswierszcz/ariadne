package fx.msg;

import org.springframework.amqp.rabbit.annotation.RabbitHandler;
import org.springframework.amqp.rabbit.annotation.RabbitListener;

@RabbitListener("mixed")
public class RabbitClassListener {
    @RabbitHandler
    public void onString(String s) {}
}
