package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.notification.NotificationResponse;
import com.fei.twitterjavaapi.model.entity.Notification;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.NotificationRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
public class NotificationService {

    private final NotificationRepository notificationRepository;

    @Transactional(readOnly = true)
    public PageResponse<NotificationResponse> getUserNotifications(User user, int page, int size) {
        Pageable pageable = PageRequest.of(page, size);

        // Uses the optimized "JOIN FETCH" query to avoid N+1
        Page<Notification> pageNotifications = notificationRepository.findByRecipientId(user.getId(), pageable);

        return PageResponse.from(
                pageNotifications.map(NotificationResponse::fromEntity)
        );
    }

    @Transactional(readOnly = true)
    public long countUnread(User user) {
        return notificationRepository.countByRecipientIdAndIsReadFalse(user.getId());
    }

    @Transactional
    public void markAllAsRead(User user) {
        notificationRepository.markAllAsRead(user.getId());
    }
}