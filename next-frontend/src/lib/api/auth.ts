import { apiClient } from './client';
import type { AuthData, AuthResponse, Family, User } from '@/types';

export const authApi = {
  login: (credentials: AuthData) =>
    apiClient<AuthResponse>('/v1/authorize', {
      method: 'POST',
      body: credentials,
    }),

  logout: () =>
    apiClient<void>('/v1/logout', {
      method: 'POST',
    }),

  getUser: () => apiClient<User>('/v1/user'),
  getFamily: () => apiClient<Family>('/v1/family'),
};
