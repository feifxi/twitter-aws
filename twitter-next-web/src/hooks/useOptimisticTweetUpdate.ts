import { useQueryClient, type InfiniteData } from '@tanstack/react-query';
import type { TweetResponse, PageResponse } from '@/types';
import { feedQueryKey } from './useFeed';

export function useUpdateTweetCache() {
    const queryClient = useQueryClient();

    // Updater now returns TweetResponse | null (null means delete)
    return (tweetId: number, updater: (t: TweetResponse) => TweetResponse | null) => {
        // Helper to update a tweet within a list (array of tweets)
        const updateTweetList = (list: TweetResponse[]) => {
            return list.map((t) => {
                // Direct match
                if (t.id === tweetId) return updater(t);

                // Retweet match (update the original tweet inside the retweet)
                // If I'm strictly updating the INNER tweet, I probably don't want to delete the OUTER wrapper 
                // UNLESS the logic specifically handles wrapper deletion (which is the goal here).
                // But specifically for 'toggleRetweetInTweet', if I un-retweet, I might want to delete the wrapper.

                // If the updater returns null for the inner tweet, what happens?
                // The wrapper should probably remain but the inner tweet is gone? No, that breaks invariants.
                // Actually, 'tweetId' passed to this function is usually the ID of the tweet being interacted with.
                // If I interact with a Retweet Wrapper (ID 100) of Original (ID 50).
                // I usually pass 50 to the mutation.

                if (t.retweetedTweet?.id === tweetId) {
                    // Try updating the wrapper itself first. 
                    // This allows logic like "delete this wrapper if it's my own retweet" to work.
                    const wrapperUpdate = updater(t);

                    // If the updater returns null, it means the wrapper should be deleted.
                    if (wrapperUpdate === null) return null;

                    // If the updater returned a NEW object, it means it handled the wrapper logic explicitly.
                    // We check for referential equality assumed in standard immutable update patterns.
                    if (wrapperUpdate !== t) {
                        return wrapperUpdate;
                    }

                    // Fallback: The updater didn't change the wrapper (returned 't'), 
                    // so we assume it only knows how to update the inner tweet.
                    const updatedOriginal = updater(t.retweetedTweet!);
                    if (updatedOriginal === null) {
                        // If the inner tweet is explicitly deleted, the wrapper is invalid.
                        return null;
                    }
                    return { ...t, retweetedTweet: updatedOriginal };
                }
                return t;
            }).filter((t): t is TweetResponse => t !== null); // Filter out nulls
        };

        // Helper for InfiniteQuery data (pages of content)
        const updateInfiniteData = (old: InfiniteData<PageResponse<TweetResponse>> | undefined) => {
            if (!old) return old;
            return {
                ...old,
                pages: old.pages.map((page) => ({
                    ...page,
                    content: updateTweetList(page.content),
                })),
            };
        };

        // Helper for PageResponse data (single page)
        const updatePageResponse = (old: PageResponse<TweetResponse> | undefined) => {
            if (!old) return old;
            return {
                ...old,
                content: updateTweetList(old.content),
            };
        };

        // 1. Global Feed (Infinite)
        queryClient.setQueryData(feedQueryKey('global'), updateInfiniteData);

        // 2. Following Feed (Infinite)
        queryClient.setQueryData(feedQueryKey('following'), updateInfiniteData);

        // 3. User Feeds (Infinite) - e.g., ['feeds', 'user', userId]
        queryClient.setQueriesData<InfiniteData<PageResponse<TweetResponse>>>(
            { queryKey: ['feeds', 'user'] },
            updateInfiniteData
        );

        // 4. Search Results (PageResponse) - e.g., ['search', 'tweets', q]
        queryClient.setQueriesData<PageResponse<TweetResponse>>(
            { queryKey: ['search', 'tweets'] },
            updatePageResponse
        );

        // 5. Tweet Details & Replies
        queryClient.setQueriesData<unknown>(
            { queryKey: ['tweets'] },
            (old: unknown) => {
                if (!old) return old;

                // Case A: Infinite Data (Replies)
                if ((old as InfiniteData<PageResponse<TweetResponse>>).pages) {
                    return updateInfiniteData(old as InfiniteData<PageResponse<TweetResponse>>);
                }

                // Case B: Single Tweet
                const t = old as TweetResponse;
                if (t.id === tweetId) {
                    const res = updater(t);
                    if (res === null) return undefined; // Remove from cache (makes it undefined/null)
                    return res;
                }
                if (t.retweetedTweet?.id === tweetId) {
                    const updatedOriginal = updater(t.retweetedTweet!);
                    if (updatedOriginal === null) return undefined; // Cascading delete
                    return { ...t, retweetedTweet: updatedOriginal };
                }

                return old;
            }
        );
    };
}
