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
@Table(name = "follows")
public class Follow {

    @EmbeddedId
    private FollowKey id;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("followerId") // Maps to FollowKey.followerId
    @JoinColumn(name = "follower_id")
    private User follower;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("followingId") // Maps to FollowKey.followingId
    @JoinColumn(name = "following_id")
    private User following;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    // The Composite Key Class
    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    @Embeddable
    public static class FollowKey implements Serializable {

        @Column(name = "follower_id", nullable = false)
        private Long followerId;

        @Column(name = "following_id", nullable = false)
        private Long followingId;
    }
}


