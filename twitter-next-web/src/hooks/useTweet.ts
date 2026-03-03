'use client';

import {
  useQuery,
  useInfiniteQuery,
  useQueryClient,
  useMutation,
  type QueryKey,
} from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import type { PageResponse, TweetResponse } from '@/types';
import type { TweetRequestInput } from '@/lib/validation';
import { toast } from 'sonner';

export const tweetQueryKey = (id: number) => ['tweets', id] as const;

export function useTweet(tweetId: number | null) {
  return useQuery({
    queryKey: tweetQueryKey(tweetId!),
    queryFn: async (): Promise<TweetResponse> => {
      const { data } = await axiosInstance.get<TweetResponse>(`/tweets/${tweetId}`);
      return data;
    },
    enabled: tweetId != null && tweetId > 0,
  });
}

export function useRepliesInfinite(tweetId: number | null, pageSize = 20) {
  return useInfiniteQuery({
    queryKey: ['tweets', tweetId, 'replies'],
    queryFn: async ({
      pageParam,
    }: {
      pageParam: number;
    }): Promise<PageResponse<TweetResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<TweetResponse>>(
        `/tweets/${tweetId}/replies`,
        { params: { page: pageParam, size: pageSize } }
      );
      return data;
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage.last ? undefined : lastPage.page + 1,
    enabled: tweetId != null && tweetId > 0,
  });
}

export function useInvalidateTweet() {
  const queryClient = useQueryClient();
  return (tweetId: number) =>
    queryClient.invalidateQueries({ queryKey: tweetQueryKey(tweetId) });
}

import { useAuthStore } from '@/store/useAuthStore';
import { feedQueryKey } from './useFeed';
import type { InfiniteData } from '@tanstack/react-query';

export function useCreateTweet() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();

  return useMutation({
    mutationFn: async ({ content, media, parentId }: TweetRequestInput & { media?: File }) => {
      const formData = new FormData();
      if (content) {
        formData.append('content', content);
      }
      if (parentId) {
        formData.append('parentId', parentId.toString());
      }
      if (media) {
        formData.append('media', media);
      }
      const { data } = await axiosInstance.post<TweetResponse>('/tweets', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return data;
    },
    onMutate: async () => {
      // Don't cancel queries, we're not doing optimistic updates anymore
    },
    onError: (_err, _newTweet) => {
      // Invalidate replies query
      if (_newTweet.parentId) {
        queryClient.invalidateQueries({ queryKey: ['tweets', _newTweet.parentId, 'replies'] });
      }
    },
    onSuccess: (realTweet, variables) => {
      // Helper to instantly inject the real tweet at the top of a feed
      const prependToFeed = (key: QueryKey) => {
        queryClient.setQueryData<InfiniteData<PageResponse<TweetResponse>>>(key, (old) => {
          if (!old) return old;

          const newPages = old.pages.map(page => ({ ...page }));

          if (newPages.length > 0) {
            newPages[0] = {
              ...newPages[0],
              content: [realTweet, ...newPages[0].content]
            };
          }

          return { ...old, pages: newPages };
        });
      };

      if (!variables.parentId) {
        prependToFeed(feedQueryKey('global'));
        prependToFeed(feedQueryKey('following'));
        // Update user's profile feed if viewing their own profile
        if (user) {
          prependToFeed(['user-feed', user.id]);
        }
      } else {
        // Also update replies if it was a reply
        prependToFeed(['tweets', variables.parentId, 'replies']);
        // Ensure parent tweet reply count is visually incremented
        queryClient.setQueryData<TweetResponse>(tweetQueryKey(variables.parentId), (old) => {
          if (!old) return old;
          return { ...old, replyCount: old.replyCount + 1 };
        });
      }
    }
  });
}

export function useDeleteTweet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (tweetId: number) => {
      await axiosInstance.delete(`/tweets/${tweetId}`);
    },
    onMutate: async (tweetId) => {
      // Cancel any outgoing refetches so they don't overwrite our optimistic update
      await queryClient.cancelQueries({ queryKey: ['feeds'] });
      await queryClient.cancelQueries({ queryKey: ['tweets', tweetId] });

      // Helper to remove tweet from a feed
      const removeTweetFromFeed = (feedKey: QueryKey) => {
        const previousData = queryClient.getQueryData<InfiniteData<PageResponse<TweetResponse>>>(feedKey);

        if (previousData) {
          queryClient.setQueryData<InfiniteData<PageResponse<TweetResponse>>>(feedKey, (old) => {
            if (!old) return old;
            return {
              ...old,
              pages: old.pages.map((page) => ({
                ...page,
                content: page.content.filter((t) => t.id !== tweetId),
              })),
            };
          });
        }
        return previousData;
      };

      const previousGlobal = removeTweetFromFeed(feedQueryKey('global'));
      const previousFollowing = removeTweetFromFeed(feedQueryKey('following'));

      return { previousGlobal, previousFollowing };
    },
    onError: (_err, _tweetId, context) => {
      if (context?.previousGlobal) {
        queryClient.setQueryData(feedQueryKey('global'), context.previousGlobal);
      }
      if (context?.previousFollowing) {
        queryClient.setQueryData(feedQueryKey('following'), context.previousFollowing);
      }
      toast.error('Failed to delete tweet');
    },
    onSuccess: () => {
      toast.success('Tweet deleted');
      // Invalidate to ensure consistency, though optimistic update handles immediate UI
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      queryClient.invalidateQueries({ queryKey: ['profile'] });
    },
  });
}
