package com.fei.twitterjavaapi.model.dto.user;

import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.enums.Role;

public record UserResponse(
        Long id,
        String username,
        String email,
        String displayName,
        String avatarUrl,
        String bio,
        Role role,
        int followersCount,
        int followingCount,
        boolean followedByMe
) {

    // FEED MAPPER: Use when viewing a list of tweets or a profile
    public static UserResponse fromEntity(User user, boolean isFollowing) {
        if (user == null) return null;

        return new UserResponse(
                user.getId(),
                user.getHandle(),
                user.getEmail(),
                user.getDisplayName(),
                user.getAvatarUrl(),
                user.getBio(),
                user.getRole(),
                user.getFollowersCount(),
                user.getFollowingCount(),
                isFollowing
        );
    }

    // Overloaded method for when we don't need/have follow context (like login response)
    public static UserResponse fromEntity(User user) {
        return fromEntity(user, false);
    }
}