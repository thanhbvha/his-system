import { create } from 'zustand';
import apiClient from '@/lib/apiClient';

export interface Patient {
  id: string;
  full_name: string;
  dob: string; // YYYY-MM-DD
  gender: string;
  phone_masked?: string;
  phone?: string;
  cccd?: string;
  email?: string;
  address?: string;
  address_detail?: string;
  patient_code?: string;
}

interface PatientStore {
  searchResults: Patient[];
  selectedPatient: Patient | null;
  isLoading: boolean;
  searchPatients: (query: string) => Promise<void>;
  selectPatient: (patient: Patient | null) => void;
  createPatient: (data: Partial<Patient>) => Promise<Patient>;
  getPatientDetail: (id: string) => Promise<void>;
  getPatientHistory: (id: string) => Promise<any[]>;
}

export const usePatientStore = create<PatientStore>((set) => ({
  searchResults: [],
  selectedPatient: null,
  isLoading: false,

  searchPatients: async (query: string) => {
    if (!query) {
      set({ searchResults: [] });
      return;
    }
    set({ isLoading: true });
    try {
      let url = `/patients?q=${encodeURIComponent(query)}`;
      // Auto-detect CCCD (12 digits) or Phone (10 digits)
      if (/^\d{12}$/.test(query)) {
        url = `/patients?cccd=${encodeURIComponent(query)}`;
      } else if (/^\d{10,11}$/.test(query)) {
        url = `/patients?phone=${encodeURIComponent(query)}`;
      }

      const res = await apiClient.get(url);
      set({ searchResults: res.data.data.items || [] });
    } catch (error) {
      console.error("Failed to search patients:", error);
      set({ searchResults: [] });
    } finally {
      set({ isLoading: false });
    }
  },

  selectPatient: (patient) => {
    set({ selectedPatient: patient });
  },

  createPatient: async (data) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.post('/patients', data);
      return res.data.data;
    } catch (error) {
      console.error("Failed to create patient:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  getPatientDetail: async (id: string) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/patients/${id}`);
      set({ selectedPatient: res.data.data });
    } catch (error) {
      console.error("Failed to fetch patient detail:", error);
    } finally {
      set({ isLoading: false });
    }
  },

  getPatientHistory: async (id: string) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/patients/${id}/history`);
      return res.data.data || [];
    } catch (error) {
      console.error("Failed to fetch patient history:", error);
      return [];
    } finally {
      set({ isLoading: false });
    }
  }
}));
