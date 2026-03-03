package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.BadRequestException;
import com.fei.twitterjavaapi.exception.ResourceNotFoundException;
import com.fei.twitterjavaapi.mapper.UserMapper;
import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.user.UpdateProfileRequest;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.event.UserFollowedEvent;
import com.fei.twitterjavaapi.repository.FollowRepository;
import com.fei.twitterjavaapi.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionTemplate;
import org.springframework.web.multipart.MultipartFile;

@Service
@RequiredArgsConstructor
@Slf4j
public class UserService {

    private final UserRepository userRepository;
    private final FollowRepository followRepository;
    private final FileStorageService fileStorageService;
    private final UserMapper userMapper;
    private final TransactionTemplate transactionTemplate;
    private final ApplicationEventPublisher eventPublisher;

    public UserResponse updateProfile(Long currentUserId, UpdateProfileRequest request, MultipartFile avatarFile) {
        log.info("User {} is updating profile", currentUserId);

        User user = userRepository.findById(currentUserId)
                .orElseThrow(() -> new ResourceNotFoundException("User", "id", currentUserId));

        String oldAvatarUrl = user.getAvatarUrl();
        String newAvatarUrl = null;

        // Upload File
        if (avatarFile != null && !avatarFile.isEmpty()) {
            newAvatarUrl = fileStorageService.uploadFile(avatarFile);
        }

        final String finalNewAvatarUrl = newAvatarUrl;
        User updatedUser;

        try {
            // Database : Fast Transaction
            updatedUser = transactionTemplate.execute(status ->
                    performDbUpdate(user, request, finalNewAvatarUrl)
            );

        } catch (Exception e) {
            // Compensation: DB failed, delete the orphaned new file
            if (finalNewAvatarUrl != null) {
                log.warn("DB Transaction failed. Rolling back new avatar upload: {}", finalNewAvatarUrl);
                fileStorageService.deleteFile(finalNewAvatarUrl);
            }
            throw e;
        }

        // Post-Commit Cleanup: DB succeeded, safe to delete the old file
        if (finalNewAvatarUrl != null && oldAvatarUrl != null && oldAvatarUrl.contains("blob.core.windows.net")) {
            log.info("Profile updated successfully. Cleaning up old avatar: {}", oldAvatarUrl);
            // This happens outside the DB transaction
            fileStorageService.deleteFile(oldAvatarUrl);
        }

        // Map and Return
        return userMapper.toResponse(updatedUser, updatedUser);
    }

    /**
     * Strictly handles the Database updates.
     * Executes entirely within a transaction.
     */
    private User performDbUpdate(User user, UpdateProfileRequest request, String newAvatarUrl) {
        // Update Display Name
        if (request.displayName() != null && !request.displayName().isBlank()) {
            user.setDisplayName(request.displayName().trim());
        }

        // Update Bio
        if (request.bio() != null) {
            String trimmedBio = request.bio().trim();
            user.setBio(trimmedBio.isEmpty() ? null : trimmedBio);
        }

        // Update Avatar URL
        if (newAvatarUrl != null) {
            user.setAvatarUrl(newAvatarUrl);
        }

        return userRepository.save(user);
    }

    @Transactional
    public void followUser(User currentUser, Long targetUserId) {
        if (currentUser.getId().equals(targetUserId)) {
            throw new BadRequestException("You cannot follow yourself");
        }

        // Prevent duplicate follow records
        if (followRepository.isFollowing(currentUser.getId(), targetUserId)) {
            log.warn("User {} is already following user {}. Ignoring request.", currentUser.getId(), targetUserId);
            return;
        }

        // Fetch Target User
        User targetUser = userRepository.findById(targetUserId)
                .orElseThrow(() -> new ResourceNotFoundException("User", "id", targetUserId));

        log.info("User {} is following User {}", currentUser.getId(), targetUserId);
        followRepository.followUser(currentUser.getId(), targetUserId);

        // Update Counters
        userRepository.incrementFollowingCount(currentUser.getId());
        userRepository.incrementFollowersCount(targetUserId);

        // Send Notification Event
        eventPublisher.publishEvent(new UserFollowedEvent(currentUser, targetUser));
    }

    @Transactional
    public void unfollowUser(User currentUser, Long targetUserId) {
        // Can't unfollow if not following
        if (!followRepository.isFollowing(currentUser.getId(), targetUserId)) {
            log.info("User {} is not following user {}. Ignoring unfollow request.", currentUser.getId(), targetUserId);
            return;
        }

        log.info("User {} is unfollowing User {}", currentUser.getId(), targetUserId);
        followRepository.unfollowUser(currentUser.getId(), targetUserId);

        // Update Counters
        userRepository.decrementFollowingCount(currentUser.getId());
        userRepository.decrementFollowersCount(targetUserId);
    }

    @Transactional(readOnly = true)
    public UserResponse getUserProfile(Long targetUserId, User currentUser) {
        User targetUser = userRepository.findById(targetUserId)
                .orElseThrow(() -> new ResourceNotFoundException("User", "id", targetUserId));
        return userMapper.toResponse(targetUser, currentUser);
    }

    @Transactional(readOnly = true)
    public PageResponse<UserResponse> getFollowers(Long targetUserId, User currentUser, int page, int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<User> followers = userRepository.findFollowersByUserId(targetUserId, pageable);
        return userMapper.toResponsePage(followers, currentUser);
    }

    @Transactional(readOnly = true)
    public PageResponse<UserResponse> getFollowing(Long targetUserId, User currentUser, int page, int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<User> following = userRepository.findFollowingByUserId(targetUserId, pageable);
        return userMapper.toResponsePage(following, currentUser);
    }
}