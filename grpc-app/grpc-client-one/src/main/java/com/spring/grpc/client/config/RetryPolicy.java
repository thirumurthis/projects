package com.spring.grpc.client.config;


import java.util.List;

public record RetryPolicy(double maxAttempts, String initialBackoff,
      String maxBackoff, double  backoffMultiplier, List<String> retryableStatusCodes) {
}
