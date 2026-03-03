package com.fei.twitterjavaapi.model.dto.user;

import jakarta.validation.constraints.Size;

public record UpdateProfileRequest(
        @Size(max = 100, message = "Display name cannot exceed 100 characters")
        String displayName,

        @Size(max = 160, message = "Bio cannot exceed 160 characters")
        String bio
) {}