package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.User;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);

    @Query("""
            SELECT u FROM User u
            WHERE LOWER(u.username) LIKE LOWER(CONCAT('%', :query, '%'))
               OR LOWER(u.displayName) LIKE LOWER(CONCAT('%', :query, '%'))
            ORDER BY u.followersCount DESC
            """)
    Page<User> searchUsers(@Param("query") String query, Pageable pageable);

    // FOLLOWER COUNTS (The person being followed)
    @Modifying
    @Query("UPDATE User u SET u.followersCount = u.followersCount + 1 WHERE u.id = :userId")
    void incrementFollowersCount(@Param("userId") Long userId);

    @Modifying
    @Query("UPDATE User u SET u.followersCount = u.followersCount - 1 WHERE u.id = :userId")
    void decrementFollowersCount(@Param("userId") Long userId);

    // FOLLOWING COUNTS (The person doing the action)
    @Modifying
    @Query("UPDATE User u SET u.followingCount = u.followingCount + 1 WHERE u.id = :userId")
    void incrementFollowingCount(@Param("userId") Long userId);

    @Modifying
    @Query("UPDATE User u SET u.followingCount = u.followingCount - 1 WHERE u.id = :userId")
    void decrementFollowingCount(@Param("userId") Long userId);

    // Fetch people who follow the target user
    @Query("SELECT u FROM User u JOIN Follow f ON u.id = f.id.followerId WHERE f.id.followingId = :targetUserId")
    Page<User> findFollowersByUserId(@Param("targetUserId") Long targetUserId, Pageable pageable);

    // Fetch people the target user is following
    @Query("SELECT u FROM User u JOIN Follow f ON u.id = f.id.followingId WHERE f.id.followerId = :targetUserId")
    Page<User> findFollowingByUserId(@Param("targetUserId") Long targetUserId, Pageable pageable);

    // List ALL users (except self), but sort by:
    // 1. Not followed by me (Status = 0)
    // 2. Followed by me (Status = 1)
    // 3. Then by popularity
    @Query(value = """
                SELECT u FROM User u
                LEFT JOIN Follow f ON f.following.id = u.id AND f.follower.id = :currentUserId
                WHERE u.id != :currentUserId
                ORDER BY (CASE WHEN f.id IS NULL THEN 0 ELSE 1 END) ASC, u.followersCount DESC
            """, countQuery = """
                SELECT count(u) FROM User u
                WHERE u.id != :currentUserId
            """)
    Page<User> findSuggestedUsers(@Param("currentUserId") Long currentUserId, Pageable pageable);

    // Guest: Top users globally
    @Query(value = "SELECT u FROM User u ORDER BY u.followersCount DESC", countQuery = "SELECT count(u) FROM User u")
    Page<User> findTopUsersGlobally(Pageable pageable);

    @Query("SELECT f.following.id FROM Follow f WHERE f.follower.id = :followerId AND f.following.id IN :targetIds")
    List<Long> findFollowedUserIds(@Param("followerId") Long followerId, @Param("targetIds") List<Long> targetIds);
}