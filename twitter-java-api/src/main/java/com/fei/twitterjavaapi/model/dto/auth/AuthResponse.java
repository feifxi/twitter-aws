package com.fei.twitterjavaapi.model.dto.auth;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;

public record AuthResponse(
        String accessToken,
        @JsonIgnore
        String refreshToken,
        UserResponse user
) {}