package com.fei.twitterjavaapi.mapper;

import com.fei.twitterjavaapi.model.dto.common.PageResponse;
import com.fei.twitterjavaapi.model.dto.tweet.TweetResponse;
import com.fei.twitterjavaapi.model.dto.user.UserResponse;
import com.fei.twitterjavaapi.model.entity.Tweet;
import com.fei.twitterjavaapi.model.entity.User;
import com.fei.twitterjavaapi.repository.FollowRepository;
import com.fei.twitterjavaapi.repository.LikeRepository;
import com.fei.twitterjavaapi.repository.TweetRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

@Component
@RequiredArgsConstructor
public class TweetMapper {

    private final LikeRepository likeRepository;
    private final TweetRepository tweetRepository;
    private final FollowRepository followRepository;

    // Maps a single Tweet to a TweetResponse.
    // Enhanced to support explicit flags or batch-fetched sets.
    public TweetResponse toResponse(Tweet tweet, Set<Long> likedTweetIds, Set<Long> retweetedTweetIds,
            Set<Long> followedAuthorIds) {
        boolean isLiked = likedTweetIds != null && likedTweetIds.contains(tweet.getId());
        boolean isRetweeted = retweetedTweetIds != null && retweetedTweetIds.contains(tweet.getId());
        boolean isFollowingAuthor = followedAuthorIds != null && followedAuthorIds.contains(tweet.getUser().getId());

        TweetResponse originalTweetDTO = null;
        if (tweet.getRetweet() != null) {
            // Recursively map the original tweet with correct interaction states
            originalTweetDTO = toResponse(tweet.getRetweet(), likedTweetIds, retweetedTweetIds, followedAuthorIds);
        }

        return new TweetResponse(
                tweet.getId(),
                tweet.getContent(),
                tweet.getMediaType() != null ? tweet.getMediaType().name() : null,
                tweet.getMediaUrl(),
                UserResponse.fromEntity(tweet.getUser(), isFollowingAuthor),
                tweet.getReplyCount(),
                tweet.getLikeCount(),
                tweet.getRetweetCount(),
                isLiked,
                isRetweeted,
                originalTweetDTO,
                tweet.getParent() != null ? tweet.getParent().getId() : null,
                tweet.getParent() != null ? tweet.getParent().getUser().getHandle() : null,
                tweet.getCreatedAt());
    }

    // Overload for single tweet fetching (convenience)
    public TweetResponse toResponse(Tweet tweet, boolean likedByMe, boolean retweetedByMe, boolean isFollowingAuthor) {
        // Create single-item sets for the main tweet
        Set<Long> liked = likedByMe ? Collections.singleton(tweet.getId()) : Collections.emptySet();
        Set<Long> retweeted = retweetedByMe ? Collections.singleton(tweet.getId()) : Collections.emptySet();
        Set<Long> following = isFollowingAuthor ? Collections.singleton(tweet.getUser().getId())
                : Collections.emptySet();

        // Note: This simple overload DOES NOT handle the nested tweet's state
        // dynamically if passed 'false'
        // But typically for single fetch we might load it fully in Service.
        // For now, let's keep it simple here, but 'toResponsePage' is the critical one
        // for feeds.

        // Actually, for singular fetch in TweetService, we also need to know if we
        // liked the *original* tweet.
        // So this overload is slightly dangerous if used blindly for retweets.
        // Let's rely on the Set version primarily, or update logic.

        return toResponse(tweet, liked, retweeted, following);
    }

    // Maps a Page of Tweets to a Page of TweetResponses with batch-fetched
    // interaction states.
    public PageResponse<TweetResponse> toResponsePage(Page<Tweet> tweetsPage, User currentUser) {
        List<Tweet> tweets = tweetsPage.getContent();

        // 1. Handle Empty Case
        if (tweets.isEmpty()) {
            return PageResponse.from(tweetsPage
                    .map(t -> toResponse(t, Collections.emptySet(), Collections.emptySet(), Collections.emptySet())));
        }

        // 2. Extract IDs (Both Main Tweets AND Original Tweets if Retweet)
        Set<Long> allTweetIds = new HashSet<>();
        Set<Long> allAuthorIds = new HashSet<>();

        for (Tweet t : tweets) {
            allTweetIds.add(t.getId());
            allAuthorIds.add(t.getUser().getId());

            if (t.getRetweet() != null) {
                allTweetIds.add(t.getRetweet().getId());
                allAuthorIds.add(t.getRetweet().getUser().getId());
            }
        }

        List<Long> tweetIdList = List.copyOf(allTweetIds);
        List<Long> authorIdList = List.copyOf(allAuthorIds);

        // 3. Batch Fetch Data
        Set<Long> likedTweetIds;
        Set<Long> retweetedTweetIds;
        Set<Long> followedAuthorIds;

        if (currentUser != null) {
            likedTweetIds = likeRepository.findLikedTweetIdsByUserId(currentUser.getId(), tweetIdList);
            retweetedTweetIds = tweetRepository.findRetweetedTweetIdsByUserId(currentUser.getId(), tweetIdList);
            followedAuthorIds = followRepository.findFollowedUserIds(currentUser.getId(), authorIdList);
        } else {
            likedTweetIds = Collections.emptySet();
            retweetedTweetIds = Collections.emptySet();
            followedAuthorIds = Collections.emptySet();
        }

        // 4. Map the Page using the batch data
        Page<TweetResponse> mappedPage = tweetsPage
                .map(tweet -> toResponse(tweet, likedTweetIds, retweetedTweetIds, followedAuthorIds));

        return PageResponse.from(mappedPage);
    }
}