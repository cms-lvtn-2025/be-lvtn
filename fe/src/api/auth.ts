import apiClient from './client';
import type { LoginResponse, Setup2FAResponse, VerifyResponse, ApiResponse, User } from '../types';

export const authApi = {
  login: async (email: string, password: string) => {
    const response = await apiClient.post<LoginResponse>('/auth/login', { email, password });
    return response.data;
  },

  setup2FA: async (tempToken: string) => {
    const response = await apiClient.post<Setup2FAResponse>(
      '/auth/setup-2fa',
      {},
      {
        headers: { Authorization: `Bearer ${tempToken}` },
      }
    );
    return response.data;
  },

  verify2FA: async (tempToken: string, token: string) => {
    const response = await apiClient.post<VerifyResponse>(
      '/auth/verify-2fa',
      { token },
      {
        headers: { Authorization: `Bearer ${tempToken}` },
      }
    );
    return response.data;
  },

  requestOTPEmail: async (tempToken: string) => {
    const response = await apiClient.post<ApiResponse<{ email: string; expiresIn: string }>>(
      '/auth/request-otp-email',
      {},
      {
        headers: { Authorization: `Bearer ${tempToken}` },
      }
    );
    return response.data;
  },

  verifyOTPEmail: async (tempToken: string, otp: string) => {
    const response = await apiClient.post<VerifyResponse>(
      '/auth/verify-otp-email',
      { otp },
      {
        headers: { Authorization: `Bearer ${tempToken}` },
      }
    );
    return response.data;
  },

  getMe: async () => {
    const response = await apiClient.get<ApiResponse<User>>('/auth/me');
    return response.data;
  },

  logout: async () => {
    const response = await apiClient.post<ApiResponse<void>>('/auth/logout');
    return response.data;
  },

  register: async (email: string, password: string, role: 'admin' | 'user') => {
    const response = await apiClient.post<ApiResponse<User>>('/auth/register', { email, password, role });
    return response.data;
  },
};
