import { create } from 'zustand';
import apiClient from '@/lib/apiClient';

export interface Doctor {
  id: string;
  full_name: string;
  specialty?: string;
  avatar_url?: string;
}

export interface Service {
  id: string;
  name: string;
  price?: number;
}

interface PublicStore {
  doctors: Doctor[];
  services: Service[];
  isLoading: boolean;
  fetchDoctors: (serviceId?: string) => Promise<void>;
  fetchServices: () => Promise<void>;
}

export const usePublicStore = create<PublicStore>((set) => ({
  doctors: [],
  services: [],
  isLoading: false,

  fetchDoctors: async (serviceId?: string) => {
    set({ isLoading: true });
    try {
      const url = serviceId ? `/public/doctors?service_id=${serviceId}` : `/public/doctors`;
      const res = await apiClient.get(url);
      set({ doctors: res.data.data || [] });
    } catch (error) {
      console.error("Failed to fetch doctors:", error);
      set({ doctors: [] });
    } finally {
      set({ isLoading: false });
    }
  },

  fetchServices: async () => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/public/services`);
      set({ services: res.data.data || [] });
    } catch (error) {
      console.error("Failed to fetch services:", error);
      set({ services: [] });
    } finally {
      set({ isLoading: false });
    }
  }
}));
