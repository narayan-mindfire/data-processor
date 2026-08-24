import { apiClient } from './client';
import type { Job, JobConfig, ProgressResponse, ExportResponse, PaginatedJobsResponse } from '../types/models';

export const jobService = {
  listJobs: async (limit = 10, offset = 0): Promise<PaginatedJobsResponse> => {
    const { data } = await apiClient.get(`/pipelines?limit=${limit}&offset=${offset}`);
    return data;
  },

  createJob: async (config: JobConfig): Promise<Job> => {
    const { data } = await apiClient.post('/pipelines', config);
    return data;
  },

  getJobProgress: async (id: string): Promise<ProgressResponse> => {
    const { data } = await apiClient.get(`/pipelines/${id}/progress`);
    return data;
  },

  cancelJob: async (id: string): Promise<{ message: string }> => {
    const { data } = await apiClient.patch(`/pipelines/${id}/cancel`);
    return data;
  },

  deleteJob: async (id: string): Promise<{ message: string }> => {
    const { data } = await apiClient.delete(`/pipelines/${id}`);
    return data;
  },

  getExportURLs: async (id: string, format: 'json' | 'csv'): Promise<ExportResponse> => {
    const { data } = await apiClient.get(`/pipelines/${id}/export/${format}`);
    return data;
  },

  getJobResults: async (id: string): Promise<any[]> => {
    const { data } = await apiClient.get(`/pipelines/${id}/results`);
    return data || [];
  },
};
