package com.fei.twitterjavaapi.controller;

import com.fei.twitterjavaapi.model.dto.auth.AuthResponse;
import com.fei.twitterjavaapi.model.dto.auth.GoogleAuthRequest;
import com.fei.twitterjavaapi.model.dto.common.ApiResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.service.AuthService;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.env.Environment;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseCookie;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.server.ResponseStatusException;

import static com.fei.twitterjavaapi.constant.Constant.REFRESH_TOKEN_COOKIE_NAME;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
public class AuthController {

    @Value("${jwt.refresh-expiration}")
    private long refreshExpirationMs;

    private final Environment environment;
    private final AuthService authService;

    // Login (Google -> Access + Refresh)
    @PostMapping("/google")
    public ResponseEntity<AuthResponse> googleLogin(
            @Valid @RequestBody GoogleAuthRequest authRequest,
            HttpServletResponse response
    ) {
        AuthResponse authData = authService.loginWithGoogle(authRequest.token());
        setRefreshTokenCookie(response, authData.refreshToken(), refreshExpirationMs);
        return ResponseEntity.ok(authData);
    }

    // Refresh (Old Refresh -> New Access + New Refresh)
    @PostMapping("/refresh")
    public ResponseEntity<AuthResponse> refreshToken(
            @CookieValue(name = REFRESH_TOKEN_COOKIE_NAME, required = false) String refreshToken,
            HttpServletResponse response
    ) {
        if (refreshToken == null || refreshToken.isBlank()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Refresh token missing");
        }

        AuthResponse authData = authService.refreshToken(refreshToken);
        setRefreshTokenCookie(response, authData.refreshToken(), refreshExpirationMs);
        return ResponseEntity.ok(authData);
    }

    // Logout (Revoke Refresh Token)
    @PostMapping("/logout")
    public ResponseEntity<ApiResponse> logout(@AuthenticationPrincipal User user, HttpServletResponse response) {
        // Clear the cookie regardless of whether user is null (UI cleanup)
        clearRefreshTokenCookie(response);

        if (user != null) {
            authService.logout(user);
        }

        return ResponseEntity.ok(new ApiResponse(true, "Logged out successfully"));
    }

    // Get current logged-in user
    @GetMapping("/me")
    public ResponseEntity<UserResponse> getCurrentUser(@AuthenticationPrincipal User currentUser) {
        return ResponseEntity.ok(authService.getCurrentUser(currentUser));
    }

    private void setRefreshTokenCookie(HttpServletResponse response, String token, long durationMs) {
        boolean isProd = isProduction();

        ResponseCookie cookie = ResponseCookie.from(REFRESH_TOKEN_COOKIE_NAME, token)
                .httpOnly(true)
                .secure(isProd) // HTTPS only in prod
                .path("/api/v1/auth/refresh")
                .maxAge(durationMs / 1000)
                .sameSite(isProd ? "None" : "Lax") // "None" + Secure allows cross-site in prod if needed
                .build();

        response.addHeader(HttpHeaders.SET_COOKIE, cookie.toString());
    }

    private void clearRefreshTokenCookie(HttpServletResponse response) {
        ResponseCookie cookie = ResponseCookie.from(REFRESH_TOKEN_COOKIE_NAME, "")
                .httpOnly(true)
                .secure(isProduction())
                .path("/api/v1/auth/refresh")
                .maxAge(0) // Immediately expires the cookie
                .sameSite(isProduction() ? "None" : "Lax")
                .build();

        response.addHeader(HttpHeaders.SET_COOKIE, cookie.toString());
    }

    private boolean isProduction() {
        String[] activeProfiles = environment.getActiveProfiles();
        // If "dev" profile is active, we are NOT in production
        return activeProfiles.length == 0 || !activeProfiles[0].equals("dev");
    }
}