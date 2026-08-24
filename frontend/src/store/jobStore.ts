import { create } from 'zustand';
import type { Job, ProgressResponse } from '../types/models';
import { jobService } from '../api/jobs';

interface JobState {
  jobs: Record<string, Job>;
  totalJobs: number;
  currentPage: number;
  pageSize: number;
  activeJobId: string | null;
  progress: Record<string, ProgressResponse>;
  isLoading: boolean;
  error: string | null;

  setActiveJob: (id: string | null) => void;
  fetchJobs: (page?: number) => Promise<void>;
  fetchProgress: (id: string) => Promise<void>;
  addJob: (id: string, initialProgress: ProgressResponse) => void;
}

export const useJobStore = create<JobState>((set, get) => ({
  jobs: {},
  totalJobs: 0,
  currentPage: 1,
  pageSize: 9,
  activeJobId: null,
  progress: {},
  isLoading: false,
  error: null,

  setActiveJob: (id) => set({ activeJobId: id }),

  fetchJobs: async (page) => {
    set({ isLoading: true });
    try {
      const state = get();
      const targetPage = page ?? state.currentPage;
      const limit = state.pageSize;
      const offset = (targetPage - 1) * limit;

      const response = await jobService.listJobs(limit, offset);
      
      const jobsMap: Record<string, Job> = {};
      if (response.data) {
        response.data.forEach((job) => {
          jobsMap[job.id] = job;
        });
      }
      
      set({ 
        jobs: jobsMap, 
        totalJobs: response.total_count || 0,
        currentPage: targetPage,
        isLoading: false, 
        error: null 
      });
    } catch (err: any) {
      console.error('Failed to fetch jobs', err);
      set({ isLoading: false, error: err.message });
    }
  },

  fetchProgress: async (id: string) => {
    try {
      const data = await jobService.getJobProgress(id);
      set((state) => ({
        progress: {
          ...state.progress,
          [id]: data,
        }
      }));
    } catch (err: any) {
      console.error('Failed to fetch job progress', err);
    }
  },

  addJob: (id: string, initialProgress: ProgressResponse) => {
    set((state) => ({
      activeJobId: id,
      progress: {
        ...state.progress,
        [id]: initialProgress,
      }
    }));
  }
}));
