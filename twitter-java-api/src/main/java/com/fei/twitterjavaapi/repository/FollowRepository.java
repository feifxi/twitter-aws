package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.Follow;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Set;

@Repository
public interface FollowRepository extends JpaRepository<Follow, Follow.FollowKey> {

    // 1. Check if A follows B (Fast Boolean check)
    @Query(value = """
        SELECT CASE WHEN COUNT(*) > 0 THEN true ELSE false END 
        FROM follows 
        WHERE follower_id = :followerId AND following_id = :followingId
    """, nativeQuery = true)
    boolean isFollowing(@Param("followerId") Long followerId, @Param("followingId") Long followingId);

    // 1. Follow (Insert the row)
    // use native INSERT to avoid creating entity object in Java just to save it.
    @Modifying
    @Query(value = "INSERT INTO follows (follower_id, following_id, created_at) VALUES (:followerId, :followingId, NOW())", nativeQuery = true)
    void followUser(@Param("followerId") Long followerId, @Param("followingId") Long followingId);

    // 2. Unfollow (Delete the row)
    @Modifying
    @Query(value = "DELETE FROM follows WHERE follower_id = :followerId AND following_id = :followingId", nativeQuery = true)
    void unfollowUser(@Param("followerId") Long followerId, @Param("followingId") Long followingId);

    // Batch Fetch: Returns a Set of IDs of the users I follow from a specific list
    @Query("SELECT f.id.followingId FROM Follow f WHERE f.id.followerId = :followerId AND f.id.followingId IN :targetIds")
    Set<Long> findFollowedUserIds(@Param("followerId") Long followerId, @Param("targetIds") List<Long> targetIds);
}