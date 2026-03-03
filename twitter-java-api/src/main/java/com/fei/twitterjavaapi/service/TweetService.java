package com.fei.twitterjavaapi.service;

import com.fei.twitterjavaapi.exception.AccessDeniedException;
import com.fei.twitterjavaapi.exception.BadRequestException;
import com.fei.twitterjavaapi.exception.ResourceNotFoundException;
import com.fei.twitterjavaapi.mapper.TweetMapper;
import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.tweet.TweetRequest;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.entity.Hashtag;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.model.enums.MediaType;
import com.fei.twitterjavaapi.model.event.UserRepliedEvent;
import com.fei.twitterjavaapi.repository.FollowRepository;
import com.fei.twitterjavaapi.repository.HashtagRepository;
import com.fei.twitterjavaapi.repository.LikeRepository;
import com.fei.twitterjavaapi.repository.TweetRepository;
import com.fei.twitterjavaapi.util.HashtagParser;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import org.springframework.transaction.support.TransactionTemplate;
import org.springframework.web.multipart.MultipartFile;

import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class TweetService {

    private final TweetRepository tweetRepository;
    private final LikeRepository likeRepository;
    private final FollowRepository followRepository;
    private final FileStorageService fileStorageService;
    private final HashtagRepository hashtagRepository;
    private final HashtagParser hashtagParser;
    private final TweetMapper tweetMapper;
    private final TransactionTemplate transactionTemplate;
    private final ApplicationEventPublisher eventPublisher;

    /**
     * Creates a new Tweet.
     * Strategy: Fail-Fast -> Upload (No-Lock) -> Save (Transactional)
     */
    public TweetResponse createTweet(User user, TweetRequest request, MultipartFile file) {
        log.info("User {} is creating a new tweet", user.getId());

        // 1. Validation
        String cleanContent = validateAndClean(request, file, user);
        // Do this BEFORE the expensive file upload.
        validateParentTweet(request.parentId());

        // 2. Resolve Media Type (Fail Fast if invalid type)
        MediaType mediaType = resolveMediaType(file);

        // 3. Upload File (Expensive Network Call - No DB Transaction)
        String mediaUrl = null;
        if (mediaType != MediaType.NONE) {
            // This happens without holding a DB connection
            mediaUrl = fileStorageService.uploadFile(file);
        }

        // Capture variables for lambda
        String finalMediaUrl = mediaUrl;

        try {
            // 4. Save to DB (Transactional)
            // transactionTemplate handles the transaction boundary explicitly
            return transactionTemplate
                    .execute(status -> saveTweetToDb(user, request, cleanContent, finalMediaUrl, mediaType));

        } catch (Exception e) {
            // 5. Rollback Compensation
            // If the DB save fails (SQL error, constraint violation), delete the file.
            if (finalMediaUrl != null) {
                log.warn("DB Transaction failed. Rolling back file upload: {}", finalMediaUrl);
                fileStorageService.deleteFile(finalMediaUrl);
            }
            throw e;
        }
    }

    // Internal Transactional Logic
    private TweetResponse saveTweetToDb(User user, TweetRequest request, String content, String mediaUrl,
            MediaType mediaType) {
        Tweet parent = null;

        if (request.parentId() != null) {
            parent = tweetRepository.findById(request.parentId())
                    .orElseThrow(() -> new ResourceNotFoundException("Tweet", "id", request.parentId()));

            // Update Reply Counter
            tweetRepository.incrementReplyCount(parent.getId());
        }

        Tweet tweet = Tweet.builder()
                .content(content)
                .user(user)
                .parent(parent)
                .mediaType(mediaType)
                .mediaUrl(mediaUrl)
                .hashtags(new HashSet<>())
                .build();

        // Increase/Add Hashtags
        processHashtagsForCreate(tweet, content);

        Tweet savedTweet = tweetRepository.save(tweet);
        log.info("Tweet created successfully with ID: {}", savedTweet.getId());

        if (parent != null) {
            eventPublisher.publishEvent(new UserRepliedEvent(user, parent, savedTweet));
        }

        return tweetMapper.toResponse(savedTweet, false, false, false);
    }

    @Transactional
    public void deleteTweet(User user, Long tweetId) {
        log.info("User {} requesting deletion of tweet {}", user.getId(), tweetId);

        Tweet tweet = tweetRepository.findById(tweetId)
                .orElseThrow(() -> new ResourceNotFoundException("Tweet", "id", tweetId));

        // Authorization
        if (!tweet.getUser().getId().equals(user.getId())) {
            log.warn("Unauthorized delete attempt by User {} on Tweet {}", user.getId(), tweetId);
            throw new AccessDeniedException("You can only delete your own tweets");
        }

        // Decrease/Remove Hashtags
        removeHashtagsForDelete(tweet);

        // Parent reply count cleanup
        if (tweet.getParent() != null) {
            tweetRepository.decrementReplyCount(tweet.getParent().getId());
        }

        // Harvest media URLs for clean up
        List<String> allMediaToDelete = tweetRepository.findAllMediaUrlsInThread(tweetId);

        // Delete from DB
        tweetRepository.delete(tweet);
        log.info("Tweet {} deleted from database", tweetId);

        // Schedule media cleanup (support rollback)
        if (!allMediaToDelete.isEmpty()) {
            TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
                @Override
                public void afterCommit() {
                    log.info("Transaction committed. Starting batch delete for {} files.", allMediaToDelete.size());
                    fileStorageService.deleteFiles(allMediaToDelete);
                }
            });
        }
    }

    @Transactional(readOnly = true)
    public TweetResponse getTweetById(User currentUser, Long tweetId) {
        log.info("Fetching single tweet details: {}", tweetId);

        Tweet tweet = tweetRepository.findById(tweetId)
                .orElseThrow(() -> new ResourceNotFoundException("Tweet", "id", tweetId));

        Set<Long> likedTweetIds = new HashSet<>();
        Set<Long> retweetedTweetIds = new HashSet<>();
        Set<Long> followedAuthorIds = new HashSet<>();

        if (currentUser != null) {
            Long currentUserId = currentUser.getId();

            // Gather IDs to check (Main Tweet + Potential Original Tweet)
            List<Long> tweetsToCheck = new ArrayList<>();
            tweetsToCheck.add(tweet.getId());

            List<Long> authorsToCheck = new ArrayList<>();
            authorsToCheck.add(tweet.getUser().getId());

            if (tweet.getRetweet() != null) {
                tweetsToCheck.add(tweet.getRetweet().getId());
                authorsToCheck.add(tweet.getRetweet().getUser().getId());
            }

            likedTweetIds = likeRepository.findLikedTweetIdsByUserId(currentUserId, tweetsToCheck);
            retweetedTweetIds = tweetRepository.findRetweetedTweetIdsByUserId(currentUserId, tweetsToCheck);
            followedAuthorIds = followRepository.findFollowedUserIds(currentUserId, authorsToCheck);
        }

        return tweetMapper.toResponse(tweet, likedTweetIds, retweetedTweetIds, followedAuthorIds);
    }

    @Transactional(readOnly = true)
    public PageResponse<TweetResponse> getReplies(User currentUser, Long tweetId, int page, int size) {
        log.info("Fetching replies for tweet {}. Page: {}", tweetId, page);

        if (!tweetRepository.existsById(tweetId)) {
            throw new ResourceNotFoundException("Tweet", "id", tweetId);
        }

        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.ASC, "createdAt"));
        Page<Tweet> repliesPage = tweetRepository.findAllByParentId(tweetId, pageable);
        return tweetMapper.toResponsePage(repliesPage, currentUser);
    }

    // Creation Hashtag
    private void processHashtagsForCreate(Tweet tweet, String content) {
        if (content == null)
            return;

        Set<String> tagTexts = hashtagParser.parseHashtags(content);
        if (tagTexts.isEmpty())
            return;

        List<Hashtag> existingHashtags = hashtagRepository.findByTextIn(new ArrayList<>(tagTexts));
        Map<String, Hashtag> existingMap = existingHashtags.stream()
                .collect(Collectors.toMap(Hashtag::getText, Function.identity()));

        for (String text : tagTexts) {
            Hashtag hashtag = existingMap.get(text);
            if (hashtag == null) {
                hashtag = Hashtag.builder().text(text).usageCount(0).build();
            }
            hashtag.setUsageCount(hashtag.getUsageCount() + 1);
            tweet.getHashtags().add(hashtagRepository.save(hashtag));
        }
    }

    // Deletion Hashtag
    private void removeHashtagsForDelete(Tweet tweet) {
        Set<Hashtag> tags = tweet.getHashtags();
        if (tags.isEmpty())
            return;

        for (Hashtag tag : new HashSet<>(tags)) {
            int newCount = Math.max(0, tag.getUsageCount() - 1);
            tag.setUsageCount(newCount);

            if (newCount == 0) {
                hashtagRepository.delete(tag);
            } else {
                hashtagRepository.save(tag);
            }
        }
        tweet.getHashtags().clear();
    }

    // Helper to normalize the content
    private String getCleanContent(String content) {
        if (content == null)
            return null;
        String trimmed = content.trim();
        return trimmed.isEmpty() ? null : trimmed;
    }

    private String validateAndClean(TweetRequest request, MultipartFile file, User user) {
        String cleanContent = getCleanContent(request.content());
        boolean hasContent = cleanContent != null;
        boolean hasFile = file != null && !file.isEmpty();

        if (!hasContent && !hasFile) {
            log.warn("Empty tweet attempt by User ID: {}", user.getId());
            throw new BadRequestException("A tweet must have either text content or an image/video.");
        }
        return cleanContent;
    }

    private void validateParentTweet(Long parentId) {
        if (parentId != null) {
            boolean parentExists = tweetRepository.existsById(parentId);
            if (!parentExists) {
                throw new ResourceNotFoundException("Tweet", "id", parentId);
            }
        }
    }

    private MediaType resolveMediaType(MultipartFile file) {
        if (file == null || file.isEmpty()) {
            return MediaType.NONE;
        }

        String contentType = file.getContentType();
        if (contentType == null) {
            throw new BadRequestException("Missing Content-Type");
        }

        if (contentType.startsWith("image/")) {
            return MediaType.IMAGE;
        } else if (contentType.startsWith("video/")) {
            return MediaType.VIDEO;
        }

        throw new BadRequestException("Only images and videos allowed");
    }
}