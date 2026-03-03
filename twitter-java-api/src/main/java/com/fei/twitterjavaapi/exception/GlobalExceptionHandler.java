package com.fei.twitterjavaapi.exception;

import com.fei.twitterjavaapi.model.dto.common.ErrorResponse;
import jakarta.servlet.http.HttpServletRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.multipart.MaxUploadSizeExceededException;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.web.servlet.resource.NoResourceFoundException;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

@RestControllerAdvice
@Slf4j
public class GlobalExceptionHandler {

    // 1. Business Logic Exceptions (e.g. ResourceNotFound, AccessDenied)
    @ExceptionHandler(AppException.class)
    public ResponseEntity<ErrorResponse> handleAppException(AppException ex, HttpServletRequest request) {
        log.warn("AppException occurred: [{}] {} at {}", ex.getStatus(), ex.getMessage(), request.getRequestURI());

        return buildResponse(ex.getStatus(), ex.getMessage(), request.getRequestURI(), null);
    }

    // 2. Spring ResponseStatusException
    @ExceptionHandler(ResponseStatusException.class)
    public ResponseEntity<ErrorResponse> handleResponseStatusException(ResponseStatusException ex, HttpServletRequest request) {
        log.warn("ResponseStatusException occurred: [{}] {} at {}", ex.getStatusCode(), ex.getReason(), request.getRequestURI());

        return buildResponse(
                HttpStatus.valueOf(ex.getStatusCode().value()),
                ex.getReason(),
                request.getRequestURI(),
                null
        );
    }

    // 3. Validation Errors (DTO @Valid failed)
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<ErrorResponse> handleValidationErrors(MethodArgumentNotValidException ex, HttpServletRequest request) {
        // Collect errors
        List<ErrorResponse.ValidationError> validationErrors = ex.getBindingResult().getFieldErrors()
                .stream()
                .map(error -> new ErrorResponse.ValidationError(
                        error.getField(),
                        error.getDefaultMessage()
                ))
                .collect(Collectors.toList());

        log.warn("Validation failed at {}: {}", request.getRequestURI(), validationErrors);

        return buildResponse(
                HttpStatus.BAD_REQUEST,
                "Validation Failed",
                request.getRequestURI(),
                validationErrors
        );
    }

    // 4. File Size Limit Exceeded
    @ExceptionHandler(MaxUploadSizeExceededException.class)
    public ResponseEntity<ErrorResponse> handleMaxSizeException(MaxUploadSizeExceededException ex, HttpServletRequest request) {
        log.warn("File upload exceeded limit at {}: {}", request.getRequestURI(), ex.getMessage());

        return buildResponse(
                HttpStatus.CONTENT_TOO_LARGE,
                "File is too large! Maximum size allowed is 5MB.",
                request.getRequestURI(),
                null
        );
    }

    // 5. Handle Not Found Endpoint
    @ExceptionHandler(NoResourceFoundException.class)
    public ResponseEntity<ErrorResponse> handleNoResourceFound(NoResourceFoundException ex, HttpServletRequest request) {
        log.warn("404 Not Found: {} {}", request.getMethod(), request.getRequestURI());

        return buildResponse(
                HttpStatus.NOT_FOUND,
                "The requested resource was not found.",
                request.getRequestURI(),
                null
        );
    }

    // 6. Global Fallback (500 Internal Server Error)
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGlobalException(Exception ex, HttpServletRequest request) {
        // ERROR level because this is a server crash/bug (NullPointer, SQL down, etc.)
        log.error("Unexpected Internal Server Error at {}: ", request.getRequestURI(), ex);

        return buildResponse(
                HttpStatus.INTERNAL_SERVER_ERROR,
                "An unexpected error occurred.",
                request.getRequestURI(),
                null
        );
    }

    // Helper Method
    private ResponseEntity<ErrorResponse> buildResponse(HttpStatus status, String message, String path, List<ErrorResponse.ValidationError> errors) {
        ErrorResponse response = new ErrorResponse(
                LocalDateTime.now(),
                status.value(),
                status.getReasonPhrase(),
                message,
                path,
                errors
        );
        return new ResponseEntity<>(response, status);
    }
}