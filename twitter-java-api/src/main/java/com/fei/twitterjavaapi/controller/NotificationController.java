package com.fei.twitterjavaapi.controller;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.notification.NotificationResponse;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.service.NotificationService;
import com.fei.twitterjavaapi.manager.SseManager;
import lombok.RequiredArgsConstructor;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

@RestController
@RequestMapping("/api/v1/notifications")
@RequiredArgsConstructor
public class NotificationController {

    private final NotificationService notificationService;
    private final SseManager sseManager;

    // Subscribe to Real-Time Stream (GET /stream)
    @GetMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public SseEmitter subscribe(@AuthenticationPrincipal User user) {
        return sseManager.subscribe(user.getId());
    }

    // Get Notification History (Pagination)
    @GetMapping
    public ResponseEntity<PageResponse<NotificationResponse>> getNotifications(
            @AuthenticationPrincipal User user,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size
    ) {
        return ResponseEntity.ok(notificationService.getUserNotifications(user, page, size));
    }

    // Get Unread Count (Red Badge)
    @GetMapping("/unread-count")
    public ResponseEntity<Long> getUnreadCount(@AuthenticationPrincipal User user) {
        return ResponseEntity.ok(notificationService.countUnread(user));
    }

    // Mark all as read
    @PostMapping("/mark-read")
    public ResponseEntity<Void> markRead(@AuthenticationPrincipal User user) {
        notificationService.markAllAsRead(user);
        return ResponseEntity.ok().build();
    }
}