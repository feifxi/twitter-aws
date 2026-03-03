import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { UserResponse } from '@/types';

interface AuthState {
  user: UserResponse | null;
  accessToken: string | null;
  isInitialized: boolean;
  setAuth: (accessToken: string, user: UserResponse) => void;
  setAccessToken: (accessToken: string) => void;
  logout: () => void;
  setInitialized: (val: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      isInitialized: false,
      setAuth: (accessToken, user) => set({ accessToken, user, isInitialized: true }),
      setAccessToken: (accessToken) => set({ accessToken }),
      logout: () => set({ accessToken: null, user: null }),
      setInitialized: (val: boolean) => set({ isInitialized: val })
    }),
    {
      name: 'auth-storage', // unique name
      partialize: (state) => ({ accessToken: state.accessToken, user: state.user }), // Don't persist isInitialized
    }
  )
);

