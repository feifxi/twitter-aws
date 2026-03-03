package com.fei.twitterjavaapi.security;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fei.twitterjavaapi.model.dto.common.ErrorResponse;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.MediaType;
import org.springframework.security.access.AccessDeniedException;
import org.springframework.security.web.access.AccessDeniedHandler;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.LocalDateTime;

@Component
@RequiredArgsConstructor
public class CustomAccessDeniedHandler implements AccessDeniedHandler {

    private final ObjectMapper objectMapper;

    @Override
    public void handle(HttpServletRequest request, HttpServletResponse response, AccessDeniedException accessDeniedException) throws IOException, ServletException {
        // Set Response Type
        response.setStatus(HttpServletResponse.SC_FORBIDDEN); // 403
        response.setContentType(MediaType.APPLICATION_JSON_VALUE);

        // Build Standard Error JSON
        ErrorResponse errorResponse = new ErrorResponse(
                LocalDateTime.now(),
                403,
                "Forbidden",
                "You do not have permission to access this resource.",
                request.getRequestURI(),
                null
        );

        // Write JSON to response
        objectMapper.writeValue(response.getOutputStream(), errorResponse);
    }
}