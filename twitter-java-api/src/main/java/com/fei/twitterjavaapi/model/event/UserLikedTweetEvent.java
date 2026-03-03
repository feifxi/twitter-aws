package com.fei.twitterjavaapi.model.event;

import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserLikedTweetEvent {
    private final User actor;  // Who liked it?
    private final Tweet tweet; // Which tweet?
}
