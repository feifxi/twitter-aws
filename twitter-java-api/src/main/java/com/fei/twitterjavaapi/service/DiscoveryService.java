package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.hashtag.TrendingHashtagDTO;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.projection.TrendingHashtagProjection;
import com.fei.twitterjavaapi.repository.HashtagRepository;
import com.fei.twitterjavaapi.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class DiscoveryService {

    private final HashtagRepository hashtagRepository;
    private final UserRepository userRepository;

    /**
     * Calculates the top trending hashtags over the last 24 hours.
     * Caches the result for performance.
     */
    @Transactional(readOnly = true)
    public List<TrendingHashtagDTO> getTrendingHashtags(int limit) {
        log.info("Calculating trending hashtags limit: {}", limit);

        // 1. Try to get tags from the last 24 hours
        List<TrendingHashtagProjection> projections = hashtagRepository.findTrendingHashtags(limit);

        // 2. FALLBACK: If nothing happened today, get the all-time most popular
        if (projections.isEmpty()) {
            projections = hashtagRepository.findAllTimeTopHashtags(limit);
        }

        return projections.stream()
                .map(proj -> new TrendingHashtagDTO(proj.getText(), proj.getCount()))
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public PageResponse<UserResponse> getSuggestedUsers(User currentUser, int page, int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<User> usersPage;

        if (currentUser == null) {
            // Guest? Show global top users
            usersPage = userRepository.findTopUsersGlobally(pageable);
            return PageResponse.from(
                    usersPage.map(user -> UserResponse.fromEntity(user, false)));
        } else {
            // Logged in? Show personalized suggestions (sorted by not followed first)
            usersPage = userRepository.findSuggestedUsers(currentUser.getId(), pageable);

            // Determine which of these users are followed by me
            List<Long> userIds = usersPage.getContent().stream().map(User::getId).collect(Collectors.toList());
            List<Long> followedIds = userRepository.findFollowedUserIds(currentUser.getId(), userIds);

            return PageResponse.from(
                    usersPage.map(user -> UserResponse.fromEntity(user, followedIds.contains(user.getId()))));
        }
    }
}