'use client';

import {
  useInfiniteQuery,
  useQueryClient,
  type InfiniteData,
} from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import type { PageResponse, TweetResponse } from '@/types';

const PAGE_SIZE = 20;

export type FeedKind = 'global' | 'following';

export function feedQueryKey(kind: FeedKind) {
  return ['feeds', kind] as const;
}

async function fetchFeed(
  kind: FeedKind,
  page: number
): Promise<PageResponse<TweetResponse>> {
  const path = kind === 'global' ? '/feeds/global' : '/feeds/following';
  const { data } = await axiosInstance.get<PageResponse<TweetResponse>>(path, {
    params: { page, size: PAGE_SIZE },
  });
  return data;
}

export function useGlobalFeed(enabled = true) {
  return useInfiniteQuery({
    queryKey: feedQueryKey('global'),
    queryFn: ({ pageParam }) => fetchFeed('global', pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage.last ? undefined : lastPage.page + 1,
    enabled,
  });
}

export function useFollowingFeed(enabled = true) {
  return useInfiniteQuery({
    queryKey: feedQueryKey('following'),
    queryFn: ({ pageParam }) => fetchFeed('following', pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage.last ? undefined : lastPage.page + 1,
    enabled,
  });
}

/** Helper to update a single tweet in infinite feed cache (for optimistic like/retweet). */
export function useUpdateFeedTweet() {
  const queryClient = useQueryClient();

  function updateTweetInFeed(
    kind: FeedKind,
    tweetId: number,
    updater: (t: TweetResponse) => TweetResponse
  ) {
    queryClient.setQueryData<InfiniteData<PageResponse<TweetResponse>>>(
      feedQueryKey(kind),
      (old) => {
        if (!old) return old;
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            content: page.content.map((t) => {
              const target = t.retweetedTweet?.id === tweetId ? t.retweetedTweet : t;
              if (target.id === tweetId) return updater(t);
              if (t.retweetedTweet?.id === tweetId)
                return { ...t, retweetedTweet: updater(t.retweetedTweet!) };
              return t;
            }),
          })),
        };
      }
    );
  }

  return { updateTweetInFeed };
}
