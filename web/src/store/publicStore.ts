import { create } from 'zustand';
import apiClient from '@/lib/apiClient';

export interface Service {
  id: string;
  name: string;
  price: number;
  duration_minutes: number;
}

export interface Doctor {
  id: string;
  full_name: string;
  specialty: string;
  service_id: string;
}

export interface ClinicInfo {
  name: string;
  address: string;
  phone: string;
}

interface PublicState {
  services: Service[];
  doctors: Doctor[];
  clinicInfo: ClinicInfo | null;
  isLoading: boolean;
  fetchServices: () => Promise<void>;
  fetchDoctors: (serviceId?: string) => Promise<void>;
  fetchClinicInfo: () => Promise<void>;
}

export const usePublicStore = create<PublicState>((set) => ({
  services: [],
  doctors: [],
  clinicInfo: null,
  isLoading: false,

  fetchServices: async () => {
    set({ isLoading: true });
    try {
      const response = await apiClient.get('/public/services');
      set({ services: response.data.data });
    } catch (error) {
      console.error('Failed to fetch services', error);
    } finally {
      set({ isLoading: false });
    }
  },

  fetchDoctors: async (serviceId?: string) => {
    set({ isLoading: true });
    try {
      const url = serviceId ? `/public/doctors?service_id=${serviceId}` : '/public/doctors';
      const response = await apiClient.get(url);
      set({ doctors: response.data.data });
    } catch (error) {
      console.error('Failed to fetch doctors', error);
    } finally {
      set({ isLoading: false });
    }
  },

  fetchClinicInfo: async () => {
    set({ isLoading: true });
    try {
      const response = await apiClient.get('/public/clinic-info');
      set({ clinicInfo: response.data.data });
    } catch (error) {
      console.error('Failed to fetch clinic info', error);
    } finally {
      set({ isLoading: false });
    }
  },
}));
