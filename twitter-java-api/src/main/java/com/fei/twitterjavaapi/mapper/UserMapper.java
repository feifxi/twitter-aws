package com.fei.twitterjavaapi.mapper;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.FollowRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.List;
import java.util.Set;

@Component
@RequiredArgsConstructor
public class UserMapper {

    private final FollowRepository followRepository;

    public UserResponse toResponse(User targetUser, User currentUser) {
        boolean isFollowing = false;
        if (currentUser != null && !currentUser.getId().equals(targetUser.getId())) {
            isFollowing = followRepository.isFollowing(currentUser.getId(), targetUser.getId());
        }
        return UserResponse.fromEntity(targetUser, isFollowing);
    }

    public PageResponse<UserResponse> toResponsePage(Page<User> userPage, User currentUser) {
        if (userPage.isEmpty()) {
            return PageResponse.from(Page.empty());
        }

        List<Long> targetUserIds = userPage.getContent().stream().map(User::getId).toList();

        // Batch fetch who the current user follows from this specific list
        Set<Long> followedIds = (currentUser == null) ? Collections.emptySet()
                : followRepository.findFollowedUserIds(currentUser.getId(), targetUserIds);

        // Map the Spring Page
        Page<UserResponse> mappedPage = userPage.map(user ->
                UserResponse.fromEntity(user, followedIds.contains(user.getId()))
        );

        return PageResponse.from(mappedPage);
    }
}
