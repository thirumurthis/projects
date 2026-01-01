package com.spring.grpc;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@ConfigurationPropertiesScan
@ComponentScan(basePackages ={"com.spring.grpc"})
public class GrpcClientOneApp {

	public static void main(String[] args) {
		SpringApplication.run(GrpcClientOneApp.class, args);
	}

}
