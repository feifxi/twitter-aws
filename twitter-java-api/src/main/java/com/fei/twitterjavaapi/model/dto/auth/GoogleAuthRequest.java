package com.fei.twitterjavaapi.model.dto.auth;

import jakarta.validation.constraints.NotBlank;

public record GoogleAuthRequest(
        @NotBlank
        String token
) {}
