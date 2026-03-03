package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.BadRequestException;
import com.fei.twitterjavaapi.exception.ResourceNotFoundException;
import com.fei.twitterjavaapi.exception.UnauthorizedException;
import com.fei.twitterjavaapi.mapper.UserMapper;
import com.fei.twitterjavaapi.model.dto.auth.AuthResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.RefreshToken;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.enums.Role;
import com.fei.twitterjavaapi.repository.UserRepository;
import com.google.api.client.googleapis.auth.oauth2.GoogleIdToken;
import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier;
import com.google.api.client.http.javanet.NetHttpTransport;
import com.google.api.client.json.gson.GsonFactory;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Collections;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class AuthService {

    @Value("${spring.security.oauth2.client.registration.google.client-id}")
    private String googleClientId;

    private final UserRepository userRepository;
    private final JwtService jwtService;
    private final RefreshTokenService refreshTokenService;
    private final UserMapper userMapper;

    public AuthResponse loginWithGoogle(String googleIdToken) {
        log.info("Attempting Google login/registration");

        GoogleIdToken.Payload payload = verifyGoogleToken(googleIdToken);
        if (payload == null) {
            log.warn("Google Token verification failed. Potential invalid request or expired Token.");
            throw new BadRequestException("Invalid Google Token.");
        }

        String email = payload.getEmail();
        log.debug("Google Token verified for email: {}", email);

        // Find existing or Register new
        User user = userRepository.findByEmail(email).orElseGet(() -> {
            log.info("First time login detected. Registering new user with email: {}", email);

            // Generate a safe unique username (e.g., "john_a1b2")
            String baseName = email.split("@")[0];
            String safeUsername = baseName + "_" + UUID.randomUUID().toString().substring(0, 4);

            User newUser = new User();
            newUser.setEmail(email);
            newUser.setUsername(safeUsername);
            newUser.setDisplayName((String) payload.get("name"));
            newUser.setAvatarUrl((String) payload.get("picture"));
            newUser.setRole(Role.USER);
            newUser.setProvider("GOOGLE");
            return userRepository.save(newUser);
        });

        String accessToken = jwtService.generateToken(user);
        RefreshToken refreshToken = refreshTokenService.createRefreshToken(user);

        log.info("User logged in successfully: [ID: {}, Email: {}]", user.getId(), email);

        return new AuthResponse(
                accessToken,
                refreshToken.getToken(),
                userMapper.toResponse(user, user)
        );
    }

    public AuthResponse refreshToken(String requestRefreshToken) {
        log.debug("Attempting to rotate access token using refresh token");

        return refreshTokenService.findByToken(requestRefreshToken)
                .map(token -> {
                    log.debug("Valid refresh token found for user ID: {}", token.getUser().getId());
                    return refreshTokenService.verifyExpiration(token);
                })
                .map(RefreshToken::getUser)
                .map(user -> {
                    String newAccessToken = jwtService.generateToken(user);
                    RefreshToken newRefreshToken = refreshTokenService.createRefreshToken(user);

                    log.info("Access token successfully refreshed for user ID: {}", user.getId());
                    return new AuthResponse(
                            newAccessToken,
                            newRefreshToken.getToken(),
                            userMapper.toResponse(user, user)
                    );
                })
                .orElseThrow(() -> {
                    log.warn("Failed refresh token attempt. Token not found.");
                    return new UnauthorizedException("Refresh token not found or revoked.");
                });
    }

    public void logout(User user) {
        log.info("Logging out user ID: {}. Revoking all refresh tokens.", user.getId());
        refreshTokenService.deleteByUser(user);
    }

    public UserResponse getCurrentUser(User user) {
        // Since the 'user' comes from the SecurityContext (JWT Filter),
        // it might be a "detached" entity or missing latest updates.
        // It's safer to fetch the latest version from DB.
        return userRepository.findById(user.getId())
                .map(existingUser -> userMapper.toResponse(existingUser, existingUser)) // Self-view
                .orElseThrow(() -> new ResourceNotFoundException("User", "id", user.getId()));
    }

    private GoogleIdToken.Payload verifyGoogleToken(String idTokenString) {
        try {
            GoogleIdTokenVerifier verifier = new GoogleIdTokenVerifier.Builder(
                    new NetHttpTransport(), new GsonFactory())
                    .setAudience(Collections.singletonList(googleClientId))
                    .build();

            GoogleIdToken idToken = verifier.verify(idTokenString);
            return (idToken != null) ? idToken.getPayload() : null;
        } catch (Exception e) {
            log.error("Internal error during Google Token verification: {}", e.getMessage());
            return null;
        }
    }
}