import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@/api/axiosInstance';
import type { UserResponse } from '@/types';

export const currentUserQueryKey = ['auth', 'me'] as const;

export function useCurrentUser() {
  return useQuery({
    queryKey: currentUserQueryKey,
    queryFn: async (): Promise<UserResponse> => {
      const { data } = await axiosInstance.get<UserResponse>('/auth/me');
      return data;
    },
    staleTime: 5 * 60 * 1000,
  });
}
