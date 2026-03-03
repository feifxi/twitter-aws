package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.model.entity.User;
import io.jsonwebtoken.*;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.security.Key;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

@Service
@Slf4j
public class JwtService {

    @Value("${jwt.secret}")
    private String secretKey;

    @Value("${jwt.access-expiration}")
    private long jwtExpiration;

    public String generateToken(User user) {
        Map<String, Object> extraClaims = new HashMap<>();
        extraClaims.put("id", user.getId());
        extraClaims.put("role", user.getRole());

        return Jwts.builder()
                .setClaims(extraClaims)
                .setSubject(user.getEmail())
                .setIssuedAt(new Date(System.currentTimeMillis()))
                .setExpiration(new Date(System.currentTimeMillis() + jwtExpiration))
                .signWith(getSignInKey(), SignatureAlgorithm.HS256)
                .compact();
    }

    // Used by Auth Filter - Returns null if Token is invalid/expired
    public String extractEmail(String token) {
        Claims claims = extractAllClaims(token);
        if (claims == null) return null; // Invalid Token
        return claims.getSubject();
    }

    public boolean isTokenValid(String token) {
        return extractAllClaims(token) != null;
    }

    private Claims extractAllClaims(String token) {
        try {
            return Jwts.parserBuilder()
                    .setSigningKey(getSignInKey())
                    .build()
                    .parseClaimsJws(token)
                    .getBody();
        } catch (ExpiredJwtException e) {
            // WARN: It is normal for tokens to expire, not a server error.
            log.warn("JWT Token is expired: {}", e.getMessage());
        } catch (MalformedJwtException e) {
            log.error("Invalid JWT Token structure: {}", e.getMessage());
        } catch (UnsupportedJwtException | IllegalArgumentException | io.jsonwebtoken.security.SignatureException e) {
            log.error("JWT Validation Error: {}", e.getMessage());
        }
        return null; // Return null so the Filter treats user as "Anonymous"
    }

    private Key getSignInKey() {
        byte[] keyBytes = Decoders.BASE64.decode(secretKey);
        return Keys.hmacShaKeyFor(keyBytes);
    }
}