package com.fei.twitterjavaapi.util;

import org.springframework.stereotype.Component;
import java.util.HashSet;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Component
public class HashtagParser {

    // Regex: Matches #text but stops at spaces or punctuation
    // Supports: #java, #Spring_Boot, #react19
    private static final Pattern HASHTAG_PATTERN = Pattern.compile("#(\\w+)");

    public Set<String> parseHashtags(String content) {
        if (content == null || content.isBlank()) {
            return new HashSet<>();
        }

        Set<String> tags = new HashSet<>();
        Matcher matcher = HASHTAG_PATTERN.matcher(content);

        while (matcher.find()) {
            // Group 1 captures the text without the '#'
            // We store it as lowercase to handle #Java and #java as the same tag
            tags.add(matcher.group(1).toLowerCase());
        }
        return tags;
    }
}