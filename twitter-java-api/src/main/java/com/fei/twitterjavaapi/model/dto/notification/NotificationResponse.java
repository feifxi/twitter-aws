package com.fei.twitterjavaapi.model.dto.notification;

import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.Notification;
import com.fei.twitterjavaapi.model.enums.NotificationType;

import java.time.LocalDateTime;

public record NotificationResponse(
        Long id,
        NotificationType type,
        UserResponse actor,
        Long tweetId,
        String tweetContent,    // Text Preview ("Nice code!")
        String tweetMediaUrl,
        boolean isRead,
        LocalDateTime createdAt
) {
    public static NotificationResponse fromEntity(Notification n) {
        Long tId = null;
        String tContent = null;
        String tMedia = null;

        if (n.getTweet() != null) {
            tId = n.getTweet().getId();

            // For LIKE/RETWEET: Shows the original tweet text
            // For REPLY: Shows the reply text
            tContent = n.getTweet().getContent();
            tMedia = n.getTweet().getMediaUrl();
        }

        return new NotificationResponse(
                n.getId(),
                n.getType(),
                UserResponse.fromEntity(n.getActor(), false), // Actor info
                tId,
                tContent,
                tMedia,
                n.isRead(),
                n.getCreatedAt()
        );
    }
}