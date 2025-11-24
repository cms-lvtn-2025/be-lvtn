import apiClient from './client';
import type { ApiResponse, Service } from '../types';

export const servicesApi = {
  getAll: async (params?: { enabled?: boolean; healthy?: boolean; protocol?: string }) => {
    const response = await apiClient.get<ApiResponse<Service[]>>('/services', { params });
    return response.data;
  },

  getById: async (id: string) => {
    const response = await apiClient.get<ApiResponse<Service>>(`/services/${id}`);
    return response.data;
  },

  create: async (data: Partial<Service>) => {
    const response = await apiClient.post<ApiResponse<Service>>('/services', data);
    return response.data;
  },

  update: async (id: string, data: Partial<Service>) => {
    const response = await apiClient.put<ApiResponse<Service>>(`/services/${id}`, data);
    return response.data;
  },

  toggle: async (id: string) => {
    const response = await apiClient.patch<ApiResponse<{ id: string; enabled: boolean }>>(`/services/${id}/toggle`);
    return response.data;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete<ApiResponse<void>>(`/services/${id}`);
    return response.data;
  },

  healthCheck: async (id: string) => {
    const response = await apiClient.post<ApiResponse<{ id: string; healthy: boolean; lastHealthCheck: string }>>(
      `/services/${id}/health-check`
    );
    return response.data;
  },
};
