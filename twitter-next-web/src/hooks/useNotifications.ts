'use client';

import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/store/useAuthStore';
import { axiosInstance } from '@/api/axiosInstance';
import type { PageResponse, NotificationResponse } from '@/types';

export const notificationQueryKey = ['notifications'] as const;
export const unreadCountQueryKey = ['notifications', 'unread'] as const;

export function useNotifications(pageSize = 20) {
    return useInfiniteQuery({
        queryKey: notificationQueryKey,
        queryFn: async ({ pageParam }: { pageParam: number }): Promise<PageResponse<NotificationResponse>> => {
            const { data } = await axiosInstance.get<PageResponse<NotificationResponse>>(
                '/notifications',
                { params: { page: pageParam, size: pageSize } }
            );
            return data;
        },
        initialPageParam: 0,
        getNextPageParam: (lastPage) =>
            lastPage.last ? undefined : lastPage.page + 1,
        staleTime: 300000, // 5 minutes
    });
}

export function useUnreadCount() {
    const { accessToken } = useAuthStore();
    return useQuery({
        queryKey: unreadCountQueryKey,
        queryFn: async (): Promise<number> => {
            const { data } = await axiosInstance.get<number>('/notifications/unread-count');
            return data;
        },
        refetchInterval: 300000, // Poll every 5 minutes as fallback, rely on SSE
        enabled: !!accessToken,
    });
}

export function useMarkAllRead() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: () => axiosInstance.post('/notifications/mark-read'),
        onSuccess: () => {
            // Optimistically update unread count to 0
            queryClient.setQueryData(unreadCountQueryKey, 0);

            // Optimistically mark all loaded notifications as read
            queryClient.setQueryData<PageResponse<NotificationResponse>>(
                notificationQueryKey,
                (old) => {
                    if (!old) return old;
                    return {
                        ...old,
                        content: old.content.map((n) => ({ ...n, isRead: true })),
                    };
                }
            );
            // Also need to handle infinite query structure if it's cached as such
            queryClient.setQueriesData<{ pages: PageResponse<NotificationResponse>[] }>(
                { queryKey: notificationQueryKey },
                (old) => {
                    if (!old) return old;
                    return {
                        ...old,
                        pages: old.pages.map(page => ({
                            ...page,
                            content: page.content.map(n => ({ ...n, isRead: true }))
                        }))
                    };
                }
            );

            // Invalidate gently or not at all if we trust the optimistic update
            // queryClient.invalidateQueries({ queryKey: unreadCountQueryKey });
        },
    });
}
