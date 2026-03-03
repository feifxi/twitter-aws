package com.fei.twitterjavaapi.model.event;

import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserRepliedEvent {
    private final User actor;
    private final Tweet parentTweet; // The tweet being replied TO
    private final Tweet replyTweet;  // The new reply itself
}