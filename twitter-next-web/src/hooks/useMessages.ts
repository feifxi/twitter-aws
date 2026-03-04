import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import { useAuthStore } from '@/store/useAuthStore';
import type { ConversationResponse, MessageResponse, PublicRoomMessageResponse, PageResponse } from '@/types';

export const conversationsQueryKey = ['messages', 'conversations'] as const;
export const conversationMessagesQueryKey = (conversationId: number) =>
  ['messages', 'conversation', conversationId] as const;
export const publicRoomMessagesQueryKey = (roomKey: string) =>
  ['messages', 'public-room', roomKey] as const;

export function useConversations(pageSize = 20) {
  const { accessToken } = useAuthStore();
  return useInfiniteQuery({
    queryKey: conversationsQueryKey,
    queryFn: async ({ pageParam }: { pageParam: string | null }): Promise<PageResponse<ConversationResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<ConversationResponse>>('/messages/conversations', {
        params: { cursor: pageParam ?? undefined, size: pageSize },
      });
      return data;
    },
    enabled: !!accessToken,
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => (lastPage.hasNext ? lastPage.nextCursor : undefined),
    staleTime: 30000,
  });
}

export function useConversationMessages(conversationId: number | null, pageSize = 30) {
  const { accessToken } = useAuthStore();
  return useInfiniteQuery({
    queryKey: conversationId ? conversationMessagesQueryKey(conversationId) : ['messages', 'conversation', 'none'],
    queryFn: async ({ pageParam }: { pageParam: string | null }): Promise<PageResponse<MessageResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<MessageResponse>>(
        `/messages/conversations/${conversationId}/messages`,
        { params: { cursor: pageParam ?? undefined, size: pageSize } },
      );
      return data;
    },
    enabled: !!conversationId && !!accessToken,
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => (lastPage.hasNext ? lastPage.nextCursor : undefined),
    staleTime: 10000,
  });
}

export function useSendMessageToConversation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { conversationId: number; content: string }): Promise<MessageResponse> => {
      const { data } = await axiosInstance.post<MessageResponse>(
        `/messages/conversations/${payload.conversationId}/messages`,
        { content: payload.content },
      );
      return data;
    },
    onSuccess: (_message, variables) => {
      queryClient.invalidateQueries({ queryKey: conversationsQueryKey });
      queryClient.invalidateQueries({ queryKey: conversationMessagesQueryKey(variables.conversationId) });
    },
  });
}

export function useSendMessageToUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { userId: number; content: string }): Promise<MessageResponse> => {
      const { data } = await axiosInstance.post<MessageResponse>(
        `/messages/users/${payload.userId}/messages`,
        { content: payload.content },
      );
      return data;
    },
    onSuccess: (message) => {
      queryClient.invalidateQueries({ queryKey: conversationsQueryKey });
      queryClient.invalidateQueries({ queryKey: conversationMessagesQueryKey(message.conversationId) });
    },
  });
}

export function usePublicRoomMessages(roomKey: string, pageSize = 30) {
  return useInfiniteQuery({
    queryKey: publicRoomMessagesQueryKey(roomKey),
    queryFn: async ({ pageParam }: { pageParam: string | null }): Promise<PageResponse<PublicRoomMessageResponse>> => {
      const { data } = await axiosInstance.get<PageResponse<PublicRoomMessageResponse>>(
        `/messages/public/${roomKey}/messages`,
        { params: { cursor: pageParam ?? undefined, size: pageSize } },
      );
      return data;
    },
    enabled: !!roomKey,
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => (lastPage.hasNext ? lastPage.nextCursor : undefined),
    staleTime: 5000,
  });
}

export function useSendPublicRoomMessage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { roomKey: string; content: string }): Promise<PublicRoomMessageResponse> => {
      const { data } = await axiosInstance.post<PublicRoomMessageResponse>(
        `/messages/public/${payload.roomKey}/messages`,
        { content: payload.content },
      );
      return data;
    },
    onSuccess: (_message, variables) => {
      queryClient.invalidateQueries({ queryKey: publicRoomMessagesQueryKey(variables.roomKey) });
    },
  });
}

