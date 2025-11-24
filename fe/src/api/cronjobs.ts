import apiClient from './client';
import type { ApiResponse, CronJob } from '../types';

export const cronjobsApi = {
  getAll: async (params?: { enabled?: boolean }) => {
    const response = await apiClient.get<ApiResponse<CronJob[]>>('/cronjobs', { params });
    return response.data;
  },

  create: async (data: Partial<CronJob>) => {
    const response = await apiClient.post<ApiResponse<CronJob>>('/cronjobs', data);
    return response.data;
  },

  update: async (id: string, data: Partial<CronJob>) => {
    const response = await apiClient.put<ApiResponse<CronJob>>(`/cronjobs/${id}`, data);
    return response.data;
  },

  toggle: async (id: string) => {
    const response = await apiClient.patch<ApiResponse<{ id: string; enabled: boolean }>>(`/cronjobs/${id}/toggle`);
    return response.data;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete<ApiResponse<void>>(`/cronjobs/${id}`);
    return response.data;
  },
};
