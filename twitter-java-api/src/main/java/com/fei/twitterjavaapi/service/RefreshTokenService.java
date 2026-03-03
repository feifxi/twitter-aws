package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.AccessDeniedException;
import com.fei.twitterjavaapi.model.entity.RefreshToken;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.RefreshTokenRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class RefreshTokenService {

    @Value("${jwt.refresh-expiration}")
    private Long refreshTokenDurationMs;

    private final RefreshTokenRepository refreshTokenRepository;

    @Transactional
    public RefreshToken createRefreshToken(User user) {
        log.info("Creating new refresh token for user ID: {}", user.getId());

        // Rotate Token - Delete old, create new (enforces 1 device policy)
        int deletedCount = refreshTokenRepository.deleteByUser(user);
        if (deletedCount > 0) {
            log.debug("Rotated Token: Deleted {} old refresh tokens for user {}", deletedCount, user.getId());
        }

        RefreshToken refreshToken = RefreshToken.builder()
                .user(user)
                .token(UUID.randomUUID().toString())
                .expiryDate(Instant.now().plusMillis(refreshTokenDurationMs))
                .build();

        RefreshToken savedToken = refreshTokenRepository.save(refreshToken);
        log.info("New refresh token generated successfully for user {}", user.getId());

        return savedToken;
    }

    public Optional<RefreshToken> findByToken(String token) {
        log.debug("Searching for refresh token in database");
        return refreshTokenRepository.findByToken(token);
    }

    @Transactional
    public RefreshToken verifyExpiration(RefreshToken token) {
        log.debug("Verifying expiration for token belonging to user {}", token.getUser().getId());

        if (token.getExpiryDate().compareTo(Instant.now()) < 0) {
            log.warn("Refresh token expired at {}. Deleting from DB. User ID: {}",
                    token.getExpiryDate(), token.getUser().getId());

            refreshTokenRepository.delete(token);
            throw new AccessDeniedException("Refresh token was expired. Please make a new signin request");
        }

        return token;
    }

    @Transactional
    public void deleteByUser(User user) {
        log.info("Revoking all refresh tokens for user ID: {}", user.getId());
        refreshTokenRepository.deleteByUser(user);
    }
}