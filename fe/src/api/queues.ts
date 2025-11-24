import apiClient from './client';
import type { ApiResponse, Queue, Job } from '../types';

export const queuesApi = {
  getAll: async () => {
    const response = await apiClient.get<ApiResponse<Queue[]>>('/queues');
    return response.data;
  },

  getDetails: async (name: string) => {
    const response = await apiClient.get<ApiResponse<Queue>>(`/queues/${name}`);
    return response.data;
  },

  getJobs: async (name: string, params?: { status?: string; limit?: number; offset?: number }) => {
    const response = await apiClient.get<ApiResponse<Job[]>>(`/queues/${name}/jobs`, { params });
    return response.data;
  },

  addJob: async (name: string, jobData: { name?: string; data?: any; opts?: any }) => {
    const response = await apiClient.post<ApiResponse<{ jobId: string; queueName: string }>>(
      `/queues/${name}/jobs`,
      jobData
    );
    return response.data;
  },

  getJobStatus: async (name: string, jobId: string) => {
    const response = await apiClient.get<ApiResponse<Job>>(`/queues/${name}/jobs/${jobId}`);
    return response.data;
  },

  removeJob: async (name: string, jobId: string) => {
    const response = await apiClient.delete<ApiResponse<void>>(`/queues/${name}/jobs/${jobId}`);
    return response.data;
  },

  retryJob: async (name: string, jobId: string) => {
    const response = await apiClient.post<ApiResponse<{ jobId: string; state: string }>>(
      `/queues/${name}/jobs/${jobId}/retry`
    );
    return response.data;
  },
};
