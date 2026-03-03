package com.fei.twitterjavaapi.controller;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.hashtag.TrendingHashtagDTO;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.service.SearchService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/v1/search")
@RequiredArgsConstructor
public class SearchController {

    private final SearchService searchService;

    @GetMapping("/tweets")
    public ResponseEntity<PageResponse<TweetResponse>> searchTweets(
            @AuthenticationPrincipal User user,
            @RequestParam(name = "q") String query,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {

        PageResponse<TweetResponse> results = searchService.searchTweets(user, query, page, size);
        return ResponseEntity.ok(results);
    }

    @GetMapping("/users")
    public ResponseEntity<PageResponse<UserResponse>> searchUsers(
            @AuthenticationPrincipal User user,
            @RequestParam(name = "q") String query,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {

        PageResponse<UserResponse> results = searchService.searchUsers(user, query, page, size);
        return ResponseEntity.ok(results);
    }

    @GetMapping("/hashtags")
    public ResponseEntity<List<TrendingHashtagDTO>> searchHashtags(
            @RequestParam("q") String query,
            @RequestParam(defaultValue = "5") int limit
    ) {
        return ResponseEntity.ok(searchService.searchHashtags(query, limit));
    }
}