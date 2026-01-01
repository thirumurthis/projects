package com.spring.grpc.client;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.proto.app.OrderServiceGrpc;
import com.spring.grpc.client.config.GrpcServerConfig;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.grpc.client.GrpcChannelFactory;

import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Configuration
public class OrderClientConfig {

    private static final Logger log = LoggerFactory.getLogger(OrderClientConfig.class.getName());

    private final GrpcServerConfig retryClient;

    ObjectMapper objectMapper = new ObjectMapper();
    @Value("${spring.grpc.client.channels.local.address}")
    private String targetServerAddress;

    public OrderClientConfig(GrpcServerConfig retryClient) {
        this.retryClient = retryClient;
    }

    @Bean
    OrderServiceGrpc.OrderServiceBlockingStub stub(GrpcChannelFactory channelFactory) {

        log.info(">>>>> target server address {}",targetServerAddress);
        ManagedChannelBuilder<?> channelBuilder = ManagedChannelBuilder
                .forTarget(targetServerAddress)
                .keepAliveTime(10, TimeUnit.SECONDS);;

        if(retryClient.negotiationType().equalsIgnoreCase("PLAINTEXT")){
            channelBuilder.usePlaintext();
        }
        if(retryClient.enabled()){
            Map<String, Object> config = this.buildServiceConfig();
            log.info("configuration: {}",config.toString());
            channelBuilder.defaultServiceConfig(config);
            channelBuilder.enableRetry();
            channelBuilder.maxRetryAttempts(5);
        }

        ManagedChannel channel = channelBuilder.build();

        return OrderServiceGrpc.newBlockingStub(channel);
    }

    public Map<String, Object> buildServiceConfig() {

        Map<String, Object> svcConfig = Map.of("loadBalancingConfig",
                List.of(Map.of("weighted_round_robin", Map.of()),
                        Map.of("round_robin", Map.of()),
                        Map.of("pick_first", Map.of("shuffleAddressList", true))),
                "methodConfig", List.of(
                        Map.of("name", retryClient.name().stream().toList()),
                                //"waitForReady", true,
                                Map.of("retryPolicy",Map.of(
                                        "maxAttempts",retryClient.retryPolicy().maxAttempts(),
                                        "initialBackoff", retryClient.retryPolicy().initialBackoff(),
                                        "backoffMultiplier",retryClient.retryPolicy().backoffMultiplier(),
                                        "maxBackoff", retryClient.retryPolicy().maxBackoff(),
                                        "retryableStatusCodes", retryClient.retryPolicy().retryableStatusCodes()
                                                                )
                                )
                        )
                );
        log.info("config map {}",svcConfig);
        try {
            String jsonString = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(svcConfig);
            log.info("converted config to string : {}",jsonString);
        } catch (JsonProcessingException e) {
            log.warn("exception occurred during json conversion",e);
        }

        return svcConfig;
    }
}
