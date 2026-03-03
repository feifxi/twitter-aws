package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.Hashtag;
import com.fei.twitterjavaapi.model.projection.TrendingHashtagProjection;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface HashtagRepository extends JpaRepository<Hashtag, Long> {
     // Batch Fetch: Resolves N+1 problem
     List<Hashtag> findByTextIn(List<String> texts);

     /**
      * Autocomplete Search.
      * Finds tags starting with the prefix (Case Insensitive).
      * Example: Input "jav" -> Returns "java", "javascript", "javafx"
      */
     @Query("SELECT h FROM Hashtag h WHERE LOWER(h.text) LIKE LOWER(CONCAT(:prefix, '%')) ORDER BY h.usageCount DESC")
     List<Hashtag> searchHashtagsByPrefix(@Param("prefix") String prefix, Pageable pageable);

     @Query(value = """
        SELECT
            h.text AS text,
            COUNT(th.tweet_id) AS count
        FROM hashtags h
        JOIN tweet_hashtags th ON h.id = th.hashtag_id
        JOIN tweets t ON th.tweet_id = t.id
        WHERE t.created_at >= NOW() - INTERVAL '24 hours'
        GROUP BY h.id, h.text
        ORDER BY count DESC
        LIMIT :limit
     """, nativeQuery = true)
     List<TrendingHashtagProjection> findTrendingHashtags(@Param("limit") int limit);

     // The Fallback Query (All-Time Top)
     @Query(value = """
         SELECT
             h.text AS text,
             h.usage_count AS count
         FROM hashtags h
         ORDER BY h.usage_count DESC
         LIMIT :limit
     """, nativeQuery = true)
     List<TrendingHashtagProjection> findAllTimeTopHashtags(@Param("limit") int limit);
}