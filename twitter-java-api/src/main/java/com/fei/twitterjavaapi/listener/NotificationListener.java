package com.fei.twitterjavaapi.listener;

import com.fei.twitterjavaapi.model.dto.notification.NotificationResponse;
import com.fei.twitterjavaapi.model.entity.Notification;
import com.fei.twitterjavaapi.model.enums.NotificationType;
import com.fei.twitterjavaapi.model.event.UserFollowedEvent;
import com.fei.twitterjavaapi.model.event.UserLikedTweetEvent;
import com.fei.twitterjavaapi.model.event.UserRepliedEvent;
import com.fei.twitterjavaapi.model.event.UserRetweetedEvent;
import com.fei.twitterjavaapi.repository.NotificationRepository;
import com.fei.twitterjavaapi.manager.SseManager;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.event.TransactionPhase;
import org.springframework.transaction.event.TransactionalEventListener;

@Component
@RequiredArgsConstructor
@Slf4j
public class NotificationListener {

    private final NotificationRepository notificationRepository;
    private final SseManager sseManager;

    @Async // Run in background thread
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void handleLikeEvent(UserLikedTweetEvent event) {
        log.info("Async: Processing LIKE event for Tweet ID: {}", event.getTweet().getId());

        // Don't notify self-likes
        if (event.getActor().getId().equals(event.getTweet().getUser().getId())) return;

        // Persist to DB
        Notification notification = Notification.builder()
                .actor(event.getActor())
                .recipient(event.getTweet().getUser())
                .tweet(event.getTweet())
                .type(NotificationType.LIKE)
                .isRead(false)
                .build();

        Notification saved = notificationRepository.save(notification);

        // Push Real-Time to Frontend
        sseManager.sendNotification(
                saved.getRecipient().getId(),
                NotificationResponse.fromEntity(saved)
        );
    }

    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void handleFollowEvent(UserFollowedEvent event) {
        log.info("Async: Processing FOLLOW event. Actor: {} -> Target: {}",
                event.getActor().getId(), event.getTarget().getId());

        Notification notification = Notification.builder()
                .actor(event.getActor())
                .recipient(event.getTarget())
                .type(NotificationType.FOLLOW)
                .tweet(null) // Follows are not linked to a specific tweet
                .isRead(false)
                .build();

        Notification saved = notificationRepository.save(notification);

        sseManager.sendNotification(
                saved.getRecipient().getId(),
                NotificationResponse.fromEntity(saved)
        );
    }

    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void handleReplyEvent(UserRepliedEvent event) {
        log.info("Async: Processing REPLY event. Actor: {} -> Parent Tweet: {}",
                event.getActor().getId(), event.getParentTweet().getId());

        // Don't notify if I reply to my own tweet
        if (event.getActor().getId().equals(event.getParentTweet().getUser().getId())) return;

        Notification notification = Notification.builder()
                .actor(event.getActor())
                .recipient(event.getParentTweet().getUser()) // The owner of the PARENT tweet
                .tweet(event.getReplyTweet())                // Link to the NEW reply
                .type(NotificationType.REPLY)
                .isRead(false)
                .build();

        Notification saved = notificationRepository.save(notification);

        sseManager.sendNotification(
                saved.getRecipient().getId(),
                NotificationResponse.fromEntity(saved)
        );
    }

    @Async
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    public void handleRetweetEvent(UserRetweetedEvent event) {
        log.info("Async: Processing RETWEET event. Actor: {} -> Target Tweet: {}",
                event.getActor().getId(), event.getTargetTweet().getId());

        // Don't notify if I retweet myself
        if (event.getActor().getId().equals(event.getTargetTweet().getUser().getId())) return;

        Notification notification = Notification.builder()
                .actor(event.getActor())
                .recipient(event.getTargetTweet().getUser()) // The owner of the ORIGINAL tweet
                .tweet(event.getTargetTweet())               // Link to the ORIGINAL tweet
                .type(NotificationType.RETWEET)
                .isRead(false)
                .build();

        Notification saved = notificationRepository.save(notification);

        sseManager.sendNotification(
                saved.getRecipient().getId(),
                NotificationResponse.fromEntity(saved)
        );
    }
}