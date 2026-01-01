
package com.spring.grpc.client.standalone;

import com.google.gson.Gson;
import com.proto.app.OrderServiceGrpc;
import com.proto.app.SimRequest;
import com.proto.app.SimResponse;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ForkJoinPool;
import java.util.concurrent.TimeUnit;

/**
 * This class is an example with ManagedChannelBuilder client
 * can be executed directly with main method of java without spring
 * The reset config is added manually to the json string and client
 */

public class ChannelBuilderClient {
    private static final Logger logger = LoggerFactory.getLogger(ChannelBuilderClient.class);
    public static void main(String ... args){

        String config = """
                {
                  "methodConfig": [
                    {"name": [{
                          "service": "com.proto.app.OrderService",
                          "method": "specialCaseSimulator"
                        }],
                      "retryPolicy": {
                        "maxAttempts": 4,
                        "initialBackoff": "0.1s",
                        "maxBackoff": "1s",
                        "backoffMultiplier": 2,
                        "retryableStatusCodes": ["UNAVAILABLE","DEADLINE_EXCEEDED"]
                      }
                    }
                  ]
                }
                """;

        Gson gson = new Gson();
        Map<String,?> serviceConfig = gson.fromJson(config, Map.class);

        logger.info("print retry config: {}",serviceConfig);
        // Build the channel with retry policy
        ManagedChannel channel = ManagedChannelBuilder.forAddress("localhost", 9090)
                .usePlaintext()
                .disableServiceConfigLookUp()
                .defaultServiceConfig(serviceConfig)
                .enableRetry()
                .keepAliveTime(30,TimeUnit.SECONDS)
                .keepAliveTimeout(10, TimeUnit.SECONDS)
                .keepAliveWithoutCalls(true)
                .build();

        OrderServiceGrpc.OrderServiceBlockingStub stub =
                OrderServiceGrpc.newBlockingStub(channel);

        logger.info("simulate server network based retry");

        try(ForkJoinPool executor = new ForkJoinPool()) {

            for (int i = 0; i < 5; i++) {
                executor.execute(() -> {
                    try {Thread.sleep(4_000);
                        }catch (InterruptedException e) {throw new RuntimeException(e);}
                    Map<String, String> reqMap = new HashMap<>();
                    reqMap.put("simType", "serverException");
                    SimRequest request = SimRequest.newBuilder().putAllSimulatorRequest(reqMap).build();
                    SimResponse response = stub.specialCaseSimulator(request);
                    logger.info("Server Response :- {}", response);
                    });
            }
            executor.shutdown();
            try {
                channel.shutdown().awaitTermination(60, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }
        }
    }
}
