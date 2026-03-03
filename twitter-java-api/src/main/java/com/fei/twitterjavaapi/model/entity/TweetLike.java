package com.fei.twitterjavaapi.model.entity;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;

import java.io.Serializable;
import java.time.LocalDateTime;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@Entity
@Table(name = "tweet_likes")
public class TweetLike {

    @EmbeddedId
    private TweetLikeId id;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("userId") // Automatically maps this User's ID to id.userId
    @JoinColumn(name = "user_id")
    private User user;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("tweetId") // Automatically maps this Tweet's ID to id.tweetId
    @JoinColumn(name = "tweet_id")
    private Tweet tweet;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    // The Composite Key Class
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Embeddable
    public static class TweetLikeId implements Serializable {

        @Column(name = "user_id", nullable = false)
        private Long userId;

        @Column(name = "tweet_id", nullable = false)
        private Long tweetId;
    }
}