package com.fei.twitterjavaapi.controller;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.service.FeedService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/feeds")
@RequiredArgsConstructor
public class FeedController {

    private final FeedService feedService;

    // Global / For you (Public)
    @GetMapping("/global")
    public ResponseEntity<PageResponse<TweetResponse>> getGlobalFeed(
            @AuthenticationPrincipal User user,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        return ResponseEntity.ok(feedService.getForYouFeed(user, page, size));
    }

    // Following Timeline
    @GetMapping("/following")
    public ResponseEntity<PageResponse<TweetResponse>> getFollowingFeed(
            @AuthenticationPrincipal User user,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        return ResponseEntity.ok(feedService.getFollowingTimeline(user, page, size));
    }

    @GetMapping("/user/{userId}")
    public ResponseEntity<PageResponse<TweetResponse>> getUserProfileFeed(
            @AuthenticationPrincipal User user,
            @PathVariable Long userId,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        PageResponse<TweetResponse> tweetPage = feedService.getUserTweets(user, userId, page, size);
        return ResponseEntity.ok(tweetPage);
    }
}