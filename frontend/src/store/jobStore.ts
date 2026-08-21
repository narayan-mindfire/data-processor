import { create } from 'zustand';
import type { Job, ProgressResponse } from '../types/models';
import { jobService } from '../api/jobs';

interface JobState {
  jobs: Record<string, Job>;
  activeJobId: string | null;
  progress: Record<string, ProgressResponse>;
  isLoading: boolean;
  error: string | null;

  setActiveJob: (id: string | null) => void;
  fetchJobs: () => Promise<void>;
  fetchProgress: (id: string) => Promise<void>;
  addJob: (id: string, initialProgress: ProgressResponse) => void;
}

export const useJobStore = create<JobState>((set, get) => ({
  jobs: {},
  activeJobId: null,
  progress: {},
  isLoading: false,
  error: null,

  setActiveJob: (id) => set({ activeJobId: id }),

  fetchJobs: async () => {
    set({ isLoading: true });
    try {
      const data = await jobService.listJobs();
      const jobsMap: Record<string, Job> = {};
      data.forEach((job) => {
        jobsMap[job.id] = job;
      });
      set({ jobs: jobsMap, isLoading: false, error: null });
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
