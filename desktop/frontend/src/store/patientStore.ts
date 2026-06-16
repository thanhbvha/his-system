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
      const res = await apiClient.get(`/patients?q=${encodeURIComponent(query)}`);
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
  }
}));
