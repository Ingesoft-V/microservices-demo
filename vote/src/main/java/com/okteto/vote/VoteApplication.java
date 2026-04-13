package com.okteto.vote;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class VoteApplication {

	// This is the main entry point of the Spring Boot application. It starts the application context and the embedded server.
	public static void main(String[] args) {
		SpringApplication.run(VoteApplication.class, args);
	}

}
