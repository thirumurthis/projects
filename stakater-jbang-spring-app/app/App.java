///usr/bin/env jbang "$0" "$@" ; exit $?

package app;
//JAVA 25+

//RUNTIME_OPTIONS -Dspring.config.location=file:./config/application.yaml

//DEPS org.springframework.boot:spring-boot-dependencies:4.0.1@pom
//DEPS org.springframework.boot:spring-boot-starter-web

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;


@SpringBootApplication
@ComponentScan(basePackages = {".","app"})
@RestController
@RequestMapping("/api")

public class App{
    void main(String... args) {
        SpringApplication.run(App.class, args);
    }

    String formatStr = "{\"message\": \"%s\"}";

    @Value("${app.message}")
    private String msg;

    @GetMapping("/message")
    public String getMessage(){
        return String.format(formatStr, msg);
    }
}
