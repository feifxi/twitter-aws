package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.mapper.TweetMapper;
import com.fei.twitterjavaapi.mapper.UserMapper;
import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.hashtag.TrendingHashtagDTO;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.Hashtag;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.HashtagRepository;
import com.fei.twitterjavaapi.repository.TweetRepository;
import com.fei.twitterjavaapi.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class SearchService {

    private final TweetRepository tweetRepository;
    private final UserRepository userRepository;
    private final HashtagRepository hashtagRepository;
    private final TweetMapper tweetMapper;
    private final UserMapper userMapper;

    @Transactional(readOnly = true)
    public PageResponse<TweetResponse> searchTweets(User currentUser, String rawQuery, int page, int size) {
        log.info("User {} searching for: {}", currentUser != null ? currentUser.getId() : "Guest", rawQuery);

        if (rawQuery == null || rawQuery.isBlank()) {
            return PageResponse.from(Page.empty());
        }

        String trimmedQuery = rawQuery.trim();
        PageRequest pageRequest = PageRequest.of(page, size);
        Page<Tweet> tweetPage;

        // STRATEGY 1: HASHTAG SEARCH (Exact Match)
        if (trimmedQuery.startsWith("#")) {
            // Remove # and any non-alphanumeric chars (keep underscores)
            String cleanHashtag = trimmedQuery.substring(1).replaceAll("[^a-zA-Z0-9_]", "");

            if (cleanHashtag.isEmpty()) return PageResponse.from(Page.empty());

            // Use the Optimized @EntityGraph method from TweetRepository
            tweetPage = tweetRepository.findTweetsByHashtag(cleanHashtag, pageRequest);
        }
        // STRATEGY 2: FULL-TEXT SEARCH (Fuzzy Match)
        else {
            String sanitizedQuery = prepareTsQuery(trimmedQuery);
            if (sanitizedQuery.isEmpty()) return PageResponse.from(Page.empty());

            // Use the Native PostgreSQL FTS method
            tweetPage = tweetRepository.searchTweets(sanitizedQuery, pageRequest);
        }

        return tweetMapper.toResponsePage(tweetPage, currentUser);
    }

    @Transactional(readOnly = true)
    public PageResponse<UserResponse> searchUsers(User currentUser, String rawQuery, int page, int size) {
        log.info("User {} searching for people: {}", currentUser != null ? currentUser.getId() : "Guest", rawQuery);

        if (rawQuery == null || rawQuery.trim().isEmpty()) {
            return PageResponse.from(Page.empty());
        }

        // Clean the query (Just trim it for the LIKE statement)
        String cleanQuery = rawQuery.trim();

        // Execute DB Search
        PageRequest pageRequest = PageRequest.of(page, size);
        Page<User> userPage = userRepository.searchUsers(cleanQuery, pageRequest);

        return userMapper.toResponsePage(userPage, currentUser);
    }

    /**
     * Autocomplete for the "Compose Tweet" box.
     * Query: "java" -> Returns top 5 tags starting with "java"
     */
    @Transactional(readOnly = true)
    public List<TrendingHashtagDTO> searchHashtags(String query, int limit) {
        if (query == null || query.isBlank()) {
            return Collections.emptyList();
        }

        // Remove the '#' if the frontend sent it (e.g., "#jav" -> "jav")
        String cleanPrefix = query.replace("#", "").trim();

        Pageable pageable = PageRequest.of(0, limit);
        List<Hashtag> hashtags = hashtagRepository.searchHashtagsByPrefix(cleanPrefix, pageable);

        return hashtags.stream()
                .map(h -> new TrendingHashtagDTO(h.getText(), h.getUsageCount()))
                .collect(Collectors.toList());
    }

    /**
     * Converts a raw user search like "java spring" into PostgreSQL format "java & spring".
     * Strips dangerous characters to prevent SQL syntax errors.
     */
    private String prepareTsQuery(String query) {
        if (query == null || query.isBlank()) return "";

        // Remove characters that aren't alphanumeric or spaces
        String clean = query.replaceAll("[^a-zA-Z0-9\\s]", "").trim();
        if (clean.isEmpty()) return "";

        // Join words with the '&' operator so PostgreSQL knows to search for ALL words
        return String.join(" & ", clean.split("\\s+"));
    }
}