package com.fei.twitterjavaapi.model.dto.tweet;

import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import java.time.LocalDateTime;

public record TweetResponse(
        Long id,
        String content, // Will be NULL for a Retweet
        String mediaType,
        String mediaUrl,
        UserResponse user,
        int replyCount,
        int likeCount,
        int retweetCount,
        boolean likedByMe,
        boolean retweetedByMe,
        TweetResponse originalTweet, // Null if not a retweet
        Long replyToTweetId,
        String replyToUserHandle,
        LocalDateTime createdAt) {
    public boolean isRetweet() {
        return originalTweet != null && (content == null || content.isBlank());
    }
}