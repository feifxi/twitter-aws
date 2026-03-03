package com.fei.twitterjavaapi.controller;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.hashtag.TrendingHashtagDTO;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.service.DiscoveryService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/v1/discovery")
@RequiredArgsConstructor
public class DiscoveryController {

    private final DiscoveryService discoveryService;

    @GetMapping("/trending")
    public ResponseEntity<List<TrendingHashtagDTO>> getTrendingHashtags(
            @RequestParam(defaultValue = "10") int limit
    ) {
        // Capping the limit to prevent malicious requests asking for 100,000 tags
        int safeLimit = Math.min(limit, 50);

        return ResponseEntity.ok(discoveryService.getTrendingHashtags(safeLimit));
    }

    @GetMapping("/users")
    public ResponseEntity<PageResponse<UserResponse>> getSuggestedUsers(
            @AuthenticationPrincipal User user,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        return ResponseEntity.ok(discoveryService.getSuggestedUsers(user, page, size));
    }
}