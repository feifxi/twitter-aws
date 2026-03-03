package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.RefreshToken;
import com.fei.twitterjavaapi.model.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface RefreshTokenRepository extends JpaRepository<RefreshToken, Long> {
    Optional<RefreshToken> findByToken(String token);
    int deleteByUser(User user);
}
