import apiClient from './client';
import type { ApiResponse, Workflow } from '../types';

export const workflowsApi = {
  getAll: async () => {
    const response = await apiClient.get<ApiResponse<Workflow[]>>('/workflows');
    return response.data;
  },

  getById: async (id: string) => {
    const response = await apiClient.get<ApiResponse<Workflow>>(`/workflows/${id}`);
    return response.data;
  },

  create: async (data: Partial<Workflow>) => {
    const response = await apiClient.post<ApiResponse<Workflow>>('/workflows', data);
    return response.data;
  },

  update: async (id: string, data: Partial<Workflow>) => {
    const response = await apiClient.put<ApiResponse<Workflow>>(`/workflows/${id}`, data);
    return response.data;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete<ApiResponse<void>>(`/workflows/${id}`);
    return response.data;
  },

  execute: async (id: string, data?: any) => {
    const response = await apiClient.post<ApiResponse<{ jobId: string; workflowId: string }>>(
      `/workflows/${id}/execute`,
      { data }
    );
    return response.data;
  },

  getAvailableServices: async () => {
    const response = await apiClient.get<ApiResponse<any>>('/workflows/services');
    return response.data;
  },
};
