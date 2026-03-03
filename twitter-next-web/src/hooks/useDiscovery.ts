'use client';

import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import type { PageResponse, HashtagResponse, UserResponse } from '@/types';

export function useTrendingHashtags(limit = 10) {
  return useQuery({
    queryKey: ['discovery', 'trending', limit],
    queryFn: async (): Promise<HashtagResponse[]> => {
      const { data } = await axiosInstance.get<HashtagResponse[]>(
        '/discovery/trending',
        { params: { limit } }
      );
      return data;
    },
    staleTime: 2 * 60 * 1000,
  });
}

export function useSuggestedUsers(pageSize = 5) {
  return useQuery({
    queryKey: ['discovery', 'users', pageSize],
    queryFn: async (): Promise<PageResponse<UserResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<UserResponse>>(
        '/discovery/users',
        { params: { page: 0, size: pageSize } }
      );
      return data;
    },
    staleTime: 2 * 60 * 1000,
  });
}

export function useSearchUsers(query: string, enabled = true) {
  return useQuery({
    queryKey: ['search', 'users', query],
    queryFn: async (): Promise<PageResponse<UserResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<UserResponse>>(
        '/search/users',
        { params: { q: query, size: 5 } }
      );
      return data;
    },
    enabled: query.length > 0 && enabled,
    staleTime: 60 * 1000,
  });
}

export function useSearchHashtags(query: string) {
  return useQuery({
    queryKey: ['search', 'hashtags', query],
    queryFn: async (): Promise<HashtagResponse[]> => {
      const { data } = await axiosInstance.get<HashtagResponse[]>(
        '/search/hashtags',
        { params: { q: query, limit: 10 } }
      );
      return data;
    },
    enabled: query.length > 0,
    staleTime: 60 * 1000,
  });
}

