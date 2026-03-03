package com.fei.twitterjavaapi;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

@SpringBootApplication
@EnableAsync
public class TwitterJavaApiApplication {

    public static void main(String[] args) {
        SpringApplication.run(TwitterJavaApiApplication.class, args);
    }

}
