package com.fei.twitterjavaapi.model.dto.common;

import com.fasterxml.jackson.annotation.JsonInclude;

import java.time.LocalDateTime;
import java.util.List;

public record ErrorResponse(
        LocalDateTime timestamp,
        int status,
        String error,
        String message,
        String path,

        // Only show this list if there are validation errors
        @JsonInclude(JsonInclude.Include.NON_EMPTY)
        List<ValidationError> errors
) {
        // Nested Record for details
        public record ValidationError(String field, String message) {}
}