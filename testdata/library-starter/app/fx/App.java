package fx;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

/**
 * LIBRARY_MODE starter oracle: the smallest application a user of
 * mybatis-spring-boot-starter 2.1.2 writes — Spring Boot 2.2.5, an
 * embedded HSQLDB, one @Mapper, no application.properties, no YAML, no
 * SpEL of its own. Starts the context, runs one query, closes the
 * context. Run by scripts/library-starter-oracle.sh.
 */
@SpringBootApplication
public class App {
    public static void main(String[] args) {
        try (ConfigurableApplicationContext ctx = SpringApplication.run(App.class, args)) {
            System.out.println("users=" + ctx.getBean(UserMapper.class).users());
        }
    }
}
