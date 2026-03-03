package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.ResourceNotFoundException;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.enums.MediaType;
import com.fei.twitterjavaapi.model.event.UserRetweetedEvent;
import com.fei.twitterjavaapi.repository.TweetRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.HashSet;
import java.util.Optional;

@Service
@RequiredArgsConstructor
@Slf4j
public class RetweetService {

    private final TweetRepository tweetRepository;
    private final ApplicationEventPublisher eventPublisher;

    @Transactional
    public void retweet(User user, Long tweetId) {
        log.info("User {} is retweeting tweet {}", user.getId(), tweetId);

        Tweet targetTweet = tweetRepository.findById(tweetId)
                .orElseThrow(() -> new ResourceNotFoundException("Tweet", "id", tweetId));

        // Flatten Logic: Always retweet the original post
        if (targetTweet.getRetweet() != null) {
            targetTweet = targetTweet.getRetweet();
        }

        // If already retweeted, just return
        if (tweetRepository.existsByUserIdAndRetweetId(user.getId(), targetTweet.getId())) {
            log.warn("User {} already retweeted tweet {}. Skipping.", user.getId(), targetTweet.getId());
            return;
        }

        Tweet retweet = Tweet.builder()
                .user(user)
                .retweet(targetTweet)
                .content(null)
                .mediaType(MediaType.NONE)
                .hashtags(new HashSet<>())
                .build();

        tweetRepository.save(retweet);
        tweetRepository.incrementRetweetCount(targetTweet.getId());
        log.info("Retweet created for User {} on Tweet {}", user.getId(), targetTweet.getId());

        eventPublisher.publishEvent(new UserRetweetedEvent(user, targetTweet));
    }

    @Transactional
    public void unretweet(User user, Long tweetId) {
        log.info("User {} is undoing retweet for tweet {}", user.getId(), tweetId);

        Tweet targetTweet = tweetRepository.findById(tweetId)
                .orElseThrow(() -> new ResourceNotFoundException("Tweet", "id", tweetId));

        // Flatten Logic: Find the original if target is a retweet
        if (targetTweet.getRetweet() != null) {
            targetTweet = targetTweet.getRetweet();
        }

        // Idempotency: If the retweet doesn't exist, just return
        Optional<Tweet> existingRetweet = tweetRepository.findByUserIdAndRetweetId(user.getId(), targetTweet.getId());

        if (existingRetweet.isEmpty()) {
            log.warn("No retweet record found for User {} and Tweet {}. Skipping.", user.getId(), targetTweet.getId());
            return;
        }

        tweetRepository.delete(existingRetweet.get());
        tweetRepository.decrementRetweetCount(targetTweet.getId());
        log.info("Retweet removed for User {} on Tweet {}", user.getId(), targetTweet.getId());
    }
}