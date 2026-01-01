///usr/bin/env jbang "$0" "$@" ; exit $?

package app;

//JAVA 25

//DEPS org.apache.camel:camel-bom:4.14.2@pom
//DEPS org.apache.camel:camel-grpc
//DEPS org.apache.camel:camel-core
//DEPS org.apache.camel:camel-main
//DEPS org.apache.camel:camel-stream
//DEPS org.slf4j:slf4j-api:2.0.17
//DEPS com.grpc:proto-idl:1.0.0-SNAPSHOT
//DEPS com.google.protobuf:protobuf-java:4.33.0

/**
 * Note below dependency are not being used
 * Below will skip the logs from printing to console
 * //DEPS org.slf4j:slf4j-nop:2.0.17
 * Below is not required since the api includes simplelogger
 * Adding below will display warning message multiple slf4j dependency added
 * //DEPS org.slf4j:slf4j-simple:2.0.17
 */

import org.apache.camel.*;
import org.apache.camel.builder.*;
import org.apache.camel.main.*;
import org.apache.camel.spi.*;
import static org.apache.camel.builder.PredicateBuilder.*;
import com.proto.app.OrderKey;
import com.proto.app.OrderStatus;
import java.util.Date;
import java.text.SimpleDateFormat;

import static java.lang.System.*;

public class GrpcCamelClient{

    public static void main(String ... args) throws Exception{
        setProperty("org.slf4j.simpleLogger.logFile", "System.out");
        setProperty("org.slf4j.simpleLogger.defaultLogLevel", "info");
        out.println("Starting camel route...");
        Main main = new Main();

        String outputFormat = "[OrderId: %s, StatusCode: %s, UserName: %s, UpdatedBy: %s, EventTime: %s]";
        SimpleDateFormat dateFormat = new SimpleDateFormat("dd-MM-yyyy HH:mm:ss");
        main.configure().addRoutesBuilder(new RouteBuilder(){
            public void configure() throws Exception{
              //creating a request for fetching the status stream
              OrderKey orderKey = com.proto.app.OrderKey.newBuilder()
                      .setUserName("demo1")
                      .setOrderId(0L)
                      .build();
              from("timer:test?repeatCount=1&period=2500")
                  .setBody(constant(orderKey))
                  .to("grpc://localhost:9090/com.proto.app.OrderService?method=getOrderStatus&synchronous=true")
                  .split(body())
                  .log(LoggingLevel.INFO,"Recieved response : ${body}")
                  .process(exchange -> {
                    OrderStatus status = exchange.getIn().getBody(OrderStatus.class);
                      if(status != null){
                          System.out.println("RESPONSE: "+ String.format(outputFormat, status.getOrderId(),status.getStatusCode(),status.getUserName(),
                                   status.getUpdatedBy(),dateFormat.format(new Date(status.getEventTime()))));
                      }
                  });
            }
        });
        main.run();
    }
}