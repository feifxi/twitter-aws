package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.UnauthorizedException;
import com.fei.twitterjavaapi.mapper.TweetMapper;
import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.TweetRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
@Slf4j
public class FeedService {

    private final TweetRepository tweetRepository;
    private final TweetMapper tweetMapper;

    @Transactional(readOnly = true)
    public PageResponse<TweetResponse> getForYouFeed(User currentUser, int page, int size) {
        log.info("Loading 'For You' feed for user: {}", currentUser != null ? currentUser.getId() : "Guest");
        Pageable pageable = PageRequest.of(page, size);

        Page<Tweet> tweets = tweetRepository.findForYouFeed(pageable);
        return tweetMapper.toResponsePage(tweets, currentUser);
    }

    @Transactional(readOnly = true)
    public PageResponse<TweetResponse> getFollowingTimeline(User currentUser, int page, int size) {
        // If not logged in, they can't have a following feed
        if (currentUser == null) {
            throw new UnauthorizedException("Login to see following feed");
        }
        log.info("Loading 'Following' timeline for user: {}", currentUser.getId());
        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "createdAt"));

        Page<Tweet> tweets = tweetRepository.findFollowingTimeline(currentUser.getId(), pageable);
        return tweetMapper.toResponsePage(tweets, currentUser);
    }

    @Transactional(readOnly = true)
    public PageResponse<TweetResponse> getUserTweets(User currentUser, Long userId, int page, int size) {
        log.info("Fetching profile feed for user {}. Page: {}", userId, page);
        Pageable pageable =  PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "createdAt"));

        Page<Tweet> tweetsPage = tweetRepository.findAllByUserIdAndParentIdIsNull(userId, pageable);
        return tweetMapper.toResponsePage(tweetsPage, currentUser);
    }
}
