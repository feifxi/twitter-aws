package com.fei.twitterjavaapi.model.dto.common;

public record ApiResponse(
        Boolean success,
        String message
) {}