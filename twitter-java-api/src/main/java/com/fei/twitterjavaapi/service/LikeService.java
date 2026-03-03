package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.ResourceNotFoundException;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.TweetLike;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.event.UserLikedTweetEvent;
import com.fei.twitterjavaapi.repository.LikeRepository;
import com.fei.twitterjavaapi.repository.TweetRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
@Slf4j
public class LikeService {

    private final LikeRepository likeRepository;
    private final TweetRepository tweetRepository;
    private final ApplicationEventPublisher eventPublisher;

    @Transactional
    public void likeTweet(User user, Long tweetId) {
        log.info("User {} is liking tweet {}", user.getId(), tweetId);

        // Check if already liked
        if (likeRepository.existsByUserIdAndTweetId(user.getId(), tweetId)) {
            log.warn("User {} attempted to like tweet {} but already liked it", user.getId(), tweetId);
            return;
        }

        // Load Tweet
        Tweet tweet = tweetRepository.findById(tweetId)
                .orElseThrow(() -> {
                    log.warn("Like attempt failed: Tweet {} not found", tweetId);
                    return new ResourceNotFoundException("Tweet", "id", tweetId);
                });

        // Save Like
        TweetLike like = TweetLike.builder()
                .id(new TweetLike.TweetLikeId(user.getId(), tweetId))
                .user(user)
                .tweet(tweet)
                .build();

        likeRepository.save(like);

        // Increment Counter
        tweetRepository.incrementLikeCount(tweetId);
        log.info("Like count incremented for tweet {}", tweetId);

        // Send Notification Event
        eventPublisher.publishEvent(new UserLikedTweetEvent(user, tweet));
    }

    @Transactional
    public void unlikeTweet(User user, Long tweetId) {
        log.info("User {} is unliking tweet {}", user.getId(), tweetId);

        // Check if Like Exists
        if (!likeRepository.existsByUserIdAndTweetId(user.getId(), tweetId)) {
            log.warn("User {} attempted to unlike tweet {} but no record found", user.getId(), tweetId);
            return;
        }

        // Delete Like
        likeRepository.deleteByUserIdAndTweetId(user.getId(), tweetId);

        // Decrement Counter
        tweetRepository.decrementLikeCount(tweetId);
        log.info("Like count decremented for tweet {}", tweetId);
    }
}