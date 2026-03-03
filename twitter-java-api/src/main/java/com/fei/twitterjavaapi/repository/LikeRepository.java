package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.TweetLike;
import com.fei.twitterjavaapi.model.entity.TweetLike.TweetLikeId;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Set;

@Repository
public interface LikeRepository extends JpaRepository<TweetLike, TweetLikeId> {

    // 1. Check Existence (Using the entity relationships)
    @Query("SELECT COUNT(tl) > 0 FROM TweetLike tl WHERE tl.user.id = :userId AND tl.tweet.id = :tweetId")
    boolean existsByUserIdAndTweetId(@Param("userId") Long userId, @Param("tweetId") Long tweetId);

    // 2. Delete (Native is faster and cleaner for composite keys)
    @Modifying
    @Query(value = "DELETE FROM tweet_likes WHERE user_id = :userId AND tweet_id = :tweetId", nativeQuery = true)
    void deleteByUserIdAndTweetId(@Param("userId") Long userId, @Param("tweetId") Long tweetId);

    // 3. Optimized Fetch for "Did I like these tweets?"
    @Query("SELECT tl.tweet.id FROM TweetLike tl WHERE tl.user.id = :userId AND tl.tweet.id IN :tweetIds")
    Set<Long> findLikedTweetIdsByUserId(@Param("userId") Long userId, @Param("tweetIds") List<Long> tweetIds);
}