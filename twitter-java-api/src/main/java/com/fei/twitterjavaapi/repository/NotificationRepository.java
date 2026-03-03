package com.fei.twitterjavaapi.repository;

import com.fei.twitterjavaapi.model.entity.Notification;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

@Repository
public interface NotificationRepository extends JpaRepository<Notification, Long> {

    // Batch Fetching: Loads Actor & Tweet in 1 query to prevent N+1
    @Query("""
        SELECT n FROM Notification n
        JOIN FETCH n.actor
        LEFT JOIN FETCH n.tweet
        WHERE n.recipient.id = :userId
        ORDER BY n.createdAt DESC
    """)
    Page<Notification> findByRecipientId(@Param("userId") Long userId, Pageable pageable);

    long countByRecipientIdAndIsReadFalse(Long recipientId);

    @Modifying
    @Query("UPDATE Notification n SET n.isRead = true WHERE n.recipient.id = :userId")
    void markAllAsRead(@Param("userId") Long userId);
}