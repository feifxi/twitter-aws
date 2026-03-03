package com.fei.twitterjavaapi.manager;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
@Slf4j
public class SseManager {

    // Map: UserID -> Active Emitter
    private final Map<Long, SseEmitter> emitters = new ConcurrentHashMap<>();

    public SseEmitter subscribe(Long userId) {
        // 30 Minute Timeout (Standard for SSE)
        SseEmitter emitter = new SseEmitter(1800000L);

        emitters.put(userId, emitter);

        // Cleanup hooks
        emitter.onCompletion(() -> emitters.remove(userId));
        emitter.onTimeout(() -> emitters.remove(userId));
        emitter.onError((e) -> emitters.remove(userId));

        return emitter;
    }

    public void sendNotification(Long userId, Object payload) {
        SseEmitter emitter = emitters.get(userId);
        if (emitter != null) {
            try {
                emitter.send(SseEmitter.event()
                        .name("notification") // Event Name
                        .data(payload));      // JSON Data
            } catch (IOException e) {
                emitters.remove(userId); // Connection is dead
            }
        }
    }
}