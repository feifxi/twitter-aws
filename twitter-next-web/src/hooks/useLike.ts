import { useMutation, useQueryClient } from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import type { TweetResponse } from '@/types';
import { tweetQueryKey } from './useTweet';
import { useUpdateTweetCache } from './useOptimisticTweetUpdate';

function toggleLikeInTweet(t: TweetResponse, tweetId: number, liked: boolean): TweetResponse {
  const delta = liked ? 1 : -1;
  const update = (x: TweetResponse) =>
    x.id === tweetId
      ? { ...x, isLiked: liked, likeCount: Math.max(0, x.likeCount + delta) }
      : x;
  if (t.id === tweetId) return update(t) as TweetResponse;
  if (t.retweetedTweet?.id === tweetId)
    return { ...t, retweetedTweet: update(t.retweetedTweet) as TweetResponse };
  return t;
}

export function useLikeTweet() {
  const queryClient = useQueryClient();
  const updateTweetCache = useUpdateTweetCache();

  return useMutation({
    mutationFn: async (tweetId: number) => {
      await axiosInstance.post(`/tweets/${tweetId}/like`);
    },
    onMutate: async (tweetId) => {
      await queryClient.cancelQueries({ queryKey: ['feeds'] });
      await queryClient.cancelQueries({ queryKey: ['tweets', tweetId] });
      await queryClient.cancelQueries({ queryKey: ['tweets'] }); // Cancel replies too if needed
      await queryClient.cancelQueries({ queryKey: ['search'] });

      updateTweetCache(tweetId, (t) => toggleLikeInTweet(t, tweetId, true));
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    onError: (_err, tweetId) => {
      // Rollback is complex with global helper. 
      // Simplest strategy: Invalidate everything on error to restore truth.
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      queryClient.invalidateQueries({ queryKey: ['tweets'] });
      queryClient.invalidateQueries({ queryKey: ['search'] });
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    onSettled: (_data, _err, tweetId) => {
      // Only invalidate specific tweet to respect scroll position
      queryClient.invalidateQueries({ queryKey: tweetQueryKey(tweetId) });
      // Do NOT invalidate feeds or search to avoid scroll jump
    },
  });
}

export function useUnlikeTweet() {
  const queryClient = useQueryClient();
  const updateTweetCache = useUpdateTweetCache();

  return useMutation({
    mutationFn: async (tweetId: number) => {
      await axiosInstance.delete(`/tweets/${tweetId}/like`);
    },
    onMutate: async (tweetId) => {
      await queryClient.cancelQueries({ queryKey: ['feeds'] });
      await queryClient.cancelQueries({ queryKey: ['tweets', tweetId] });
      await queryClient.cancelQueries({ queryKey: ['tweets'] });
      await queryClient.cancelQueries({ queryKey: ['search'] });

      updateTweetCache(tweetId, (t) => toggleLikeInTweet(t, tweetId, false));
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    onError: (_err, tweetId) => {
      queryClient.invalidateQueries({ queryKey: ['feeds'] });
      queryClient.invalidateQueries({ queryKey: ['tweets'] });
      queryClient.invalidateQueries({ queryKey: ['search'] });
    },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    onSettled: (_data, _err, tweetId) => {
      queryClient.invalidateQueries({ queryKey: tweetQueryKey(tweetId) });
    },
  });
}
