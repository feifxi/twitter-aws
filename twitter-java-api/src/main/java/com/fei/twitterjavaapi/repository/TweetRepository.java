package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.Tweet;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.Set;

@Repository
public interface TweetRepository extends JpaRepository<Tweet, Long> {

    // ========================================================================
    // 1. CORE FEEDS (READS) - OPTIMIZED WITH @EntityGraph
    // ========================================================================

    /**
     * Retrieves the "For You" feed using a Gravity Decay Algorithm.
     *
     * <p>
     * <strong>The Algorithm:</strong> Hacker News / Reddit "Hot" Ranking
     * </p>
     * <p>
     * Formula:
     * 
     * <pre>
     * Score = (Votes - 1) / (AgeInHours + 2) ^ Gravity
     * </pre>
     * </p>
     *
     * <ul>
     * <li><strong>Votes:</strong> (Likes * 2) + (Retweets * 3) + (Replies * 1). We
     * weight interactions differently.</li>
     * <li><strong>AgeInHours:</strong> Time since creation. We add +2 to prevent
     * dividing by zero for new tweets.</li>
     * <li><strong>Gravity (1.8):</strong> How fast the score decays.
     * <ul>
     * <li>Higher Gravity (e.g., 2.0) = News sites (New stuff replaces old stuff
     * very fast).</li>
     * <li>Lower Gravity (e.g., 1.5) = Pinterest (Old viral content stays visible
     * longer).</li>
     * <li><strong>1.8</strong> is the sweet spot for a social feed in early
     * stages.</li>
     * </ul>
     * </li>
     * </ul>
     *
     * @param pageable Pagination info (page, size)
     * @return A page of tweets sorted by their calculated "Hot" score.
     */
    // NOTE: @EntityGraph does NOT work on native queries.
    @Query(value = """
                SELECT * FROM tweets t
                WHERE t.parent_id IS NULL
                ORDER BY
                    (t.like_count * 2 + t.retweet_count * 3 + t.reply_count + 1) /
                    POWER((EXTRACT(EPOCH FROM NOW() - t.created_at) / 3600) + 2, 1.8)
                    DESC,
                    t.created_at DESC
            """, countQuery = "SELECT count(*) FROM tweets WHERE parent_id IS NULL", nativeQuery = true)
    Page<Tweet> findForYouFeed(Pageable pageable);

    // Following Timeline (People you follow)
    @EntityGraph(attributePaths = { "user", "retweet", "retweet.user" })
    @Query("""
            SELECT t FROM Tweet t
            WHERE t.user.id IN (SELECT f.following.id FROM Follow f WHERE f.follower.id = :userId)
            AND t.parent IS NULL
            """)
    Page<Tweet> findFollowingTimeline(@Param("userId") Long userId, Pageable pageable);

    // Main Profile Feed (User's tweets + retweets)
    @EntityGraph(attributePaths = { "user", "retweet", "retweet.user" })
    Page<Tweet> findAllByUserIdAndParentIdIsNull(Long userId, Pageable pageable);

    // Reply Thread (Flat strategy)
    @EntityGraph(attributePaths = { "user", "retweet", "retweet.user" })
    Page<Tweet> findAllByParentId(Long parentId, Pageable pageable);

    // ========================================================================
    // 2. ATOMIC COUNTERS (WRITES)
    // Uses direct SQL updates for performance (Avoids loading entity -> modifying
    // -> saving)
    // ========================================================================

    // REPLY COUNTERS
    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.replyCount = t.replyCount + 1 WHERE t.id = :tweetId")
    void incrementReplyCount(@Param("tweetId") Long tweetId);

    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.replyCount = t.replyCount - 1 WHERE t.id = :tweetId")
    void decrementReplyCount(@Param("tweetId") Long tweetId);

    // LIKE COUNTERS
    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.likeCount = t.likeCount + 1 WHERE t.id = :tweetId")
    void incrementLikeCount(@Param("tweetId") Long tweetId);

    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.likeCount = t.likeCount - 1 WHERE t.id = :tweetId")
    void decrementLikeCount(@Param("tweetId") Long tweetId);

    // RETWEET COUNTERS
    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.retweetCount = t.retweetCount + 1 WHERE t.id = :tweetId")
    void incrementRetweetCount(@Param("tweetId") Long tweetId);

    @Modifying
    @Transactional
    @Query("UPDATE Tweet t SET t.retweetCount = t.retweetCount - 1 WHERE t.id = :tweetId")
    void decrementRetweetCount(@Param("tweetId") Long tweetId);

    // ========================================================================
    // 3. RETWEET LOGIC
    // ========================================================================

    // Check if user has already retweeted a specific tweet (Fast)
    boolean existsByUserIdAndRetweetId(Long userId, Long retweetId);

    // Find specific retweet entity (Used for "Un-retweet" to delete the row)
    // We don't need the graph here because we just need the Tweet ID to delete it.
    Optional<Tweet> findByUserIdAndRetweetId(Long userId, Long retweetId);

    // Batch Fetch: Finds which of the given tweetIds were retweeted by the user
    // Returns: A Set of IDs that should have the "Green Retweet Button" active
    @Query("SELECT t.retweet.id FROM Tweet t WHERE t.user.id = :userId AND t.retweet.id IN :tweetIds")
    Set<Long> findRetweetedTweetIdsByUserId(@Param("userId") Long userId, @Param("tweetIds") List<Long> tweetIds);

    // ========================================================================
    // 4. SEARCHING
    // ========================================================================

    // Find by Hashtag
    @EntityGraph(attributePaths = { "user", "retweet", "retweet.user" })
    @Query("""
            SELECT t FROM Tweet t
            JOIN t.hashtags h
            WHERE LOWER(h.text) = LOWER(:hashtag)
            ORDER BY t.createdAt DESC
            """)
    Page<Tweet> findTweetsByHashtag(@Param("hashtag") String hashtag, Pageable pageable);

    // Using PostgreSQL Full-Text Search (FTS)
    @Query(value = """
            SELECT * FROM tweets
            WHERE search_vector @@ to_tsquery('english', :query)
            ORDER BY ts_rank(search_vector, to_tsquery('english', :query)) DESC, created_at DESC
            """, countQuery = "SELECT count(*) FROM tweets WHERE search_vector @@ to_tsquery('english', :query)", nativeQuery = true)
    Page<Tweet> searchTweets(@Param("query") String query, Pageable pageable);

    // ========================================================================
    // 5. UTILITIES & COMPLEX QUERIES
    // ========================================================================

    // Recursive Cleanup: Finds the tweet and ALL descendants' media URLs
    // Uses CTE (Common Table Expression) for tree traversal
    @Query(value = """
                WITH RECURSIVE tweet_tree AS (
                    SELECT id, media_url, parent_id
                    FROM tweets
                    WHERE id = :tweetId
                    UNION ALL
                    SELECT t.id, t.media_url, t.parent_id
                    FROM tweets t
                    INNER JOIN tweet_tree tt ON t.parent_id = tt.id
                )
                SELECT media_url FROM tweet_tree WHERE media_url IS NOT NULL
            """, nativeQuery = true)
    List<String> findAllMediaUrlsInThread(@Param("tweetId") Long tweetId);
}