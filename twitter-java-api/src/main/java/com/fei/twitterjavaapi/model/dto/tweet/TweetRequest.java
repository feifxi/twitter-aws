package com.fei.twitterjavaapi.model.dto.tweet;

import jakarta.validation.constraints.Size;

public record TweetRequest(
        @Size(max = 280, message = "Tweet content must be under 280 characters")
        String content,
        Long parentId
) {}