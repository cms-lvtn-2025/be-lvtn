import apiClient from './client';
import type { ApiResponse, MinioConfig } from '../types';

export const minioApi = {
  getAll: async () => {
    const response = await apiClient.get<ApiResponse<MinioConfig[]>>('/minio');
    return response.data;
  },

  create: async (data: Partial<MinioConfig>) => {
    const response = await apiClient.post<ApiResponse<MinioConfig>>('/minio', data);
    return response.data;
  },

  update: async (id: string, data: Partial<MinioConfig>) => {
    const response = await apiClient.put<ApiResponse<MinioConfig>>(`/minio/${id}`, data);
    return response.data;
  },

  delete: async (id: string) => {
    const response = await apiClient.delete<ApiResponse<void>>(`/minio/${id}`);
    return response.data;
  },

  testConnection: async (id: string) => {
    const response = await apiClient.post<ApiResponse<{ connected: boolean; buckets: string[] }>>(`/minio/${id}/test`);
    return response.data;
  },
};
