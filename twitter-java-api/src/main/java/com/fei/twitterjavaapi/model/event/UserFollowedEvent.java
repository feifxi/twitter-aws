package com.fei.twitterjavaapi.model.event;

import com.fei.twitterjavaapi.model.entity.User;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserFollowedEvent {
    private final User actor;   // Who followed?
    private final User target;  // Who was followed?
}