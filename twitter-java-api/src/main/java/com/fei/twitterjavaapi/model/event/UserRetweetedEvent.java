package com.fei.twitterjavaapi.model.event;

import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserRetweetedEvent {
    private final User actor;
    private final Tweet targetTweet; // The tweet being retweeted
}