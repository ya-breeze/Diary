'use client';

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, AuthData } from '@/types';
import { authApi } from '@/lib/api';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  login: (credentials: AuthData) => Promise<void>;
  logout: () => Promise<void>;
  validateSession: () => Promise<boolean>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (credentials: AuthData) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.login(credentials);
          const user = await authApi.getUser();
          set({ user, isAuthenticated: true, isLoading: false });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      logout: async () => {
        try {
          await authApi.logout();
        } catch (error) {
          console.error('Logout failed', error);
        } finally {
          set({ user: null, isAuthenticated: false, error: null });
        }
      },

      validateSession: async () => {
        const { isAuthenticated } = get();
        if (!isAuthenticated) return false;

        try {
          const user = await authApi.getUser();
          set({ user, isAuthenticated: true });
          return true;
        } catch {
          set({ user: null, isAuthenticated: false });
          return false;
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'diary-auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
