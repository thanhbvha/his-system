import { create } from "zustand";
import apiClient from "@/lib/apiClient";

export interface VisitVital {
  id: string;
  bp_systolic?: number;
  bp_diastolic?: number;
  heart_rate?: number;
  temperature?: number;
  spo2?: number;
  weight_kg?: number;
  height_cm?: number;
  recorded_at: string;
}

export interface VisitOrder {
  id: string;
  order_type: 'LAB' | 'RADIOLOGY' | 'PROCEDURE';
  details: string;
  status: string;
}

export interface Visit {
  id: string;
  queue_entry_id?: string;
  patient: { id: string; full_name: string; patient_code?: string; dob: string; gender: string };
  doctor: { id: string; full_name: string };
  status: string;
  chief_complaint?: string;
  started_at?: string;
  vitals: VisitVital[];
  orders: VisitOrder[];
}

export interface CreateVisitPayload {
  patient_id: string;
  doctor_id: string;
  queue_entry_id: string;
  chief_complaint?: string;
}

interface VisitState {
  worklist: Visit[];
  selectedVisit: Visit | null;
  isLoading: boolean;
  error: string | null;

  fetchWorklist: (doctorId?: string, date?: string) => Promise<void>;
  fetchVisitDetail: (visitId: string) => Promise<void>;
  createVisit: (payload: CreateVisitPayload) => Promise<Visit>;
  recordVitals: (visitId: string, vitals: Partial<VisitVital>) => Promise<void>;
  createOrder: (visitId: string, order: { order_type: string; details: string }) => Promise<void>;
  closeVisit: (visitId: string) => Promise<void>;
}

export const useVisitStore = create<VisitState>((set, get) => ({
  worklist: [],
  selectedVisit: null,
  isLoading: false,
  error: null,

  fetchWorklist: async (doctorId, date) => {
    set({ isLoading: true, error: null });
    try {
      const params = new URLSearchParams();
      if (doctorId) params.append('doctor_id', doctorId);
      if (date) params.append('date', date);
      
      const res = await apiClient.get(`/visits?${params.toString()}`);
      set({ worklist: res.data.data, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  fetchVisitDetail: async (visitId) => {
    const currentVisit = get().selectedVisit;
    if (currentVisit && currentVisit.id !== visitId) {
      set({ selectedVisit: null, isLoading: true, error: null });
    } else {
      set({ isLoading: true, error: null });
    }
    try {
      const res = await apiClient.get(`/visits/${visitId}`);
      set({ selectedVisit: res.data.data, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  createVisit: async (payload) => {
    set({ isLoading: true, error: null });
    try {
      const res = await apiClient.post(`/visits`, payload);
      set({ isLoading: false });
      return res.data.data;
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },

  recordVitals: async (visitId, vitals) => {
    try {
      const res = await apiClient.post(`/visits/${visitId}/vitals`, vitals);
      // Optimistic update
      const current = get().selectedVisit;
      if (current && current.id === visitId) {
        set({
          selectedVisit: {
            ...current,
            vitals: [res.data.data, ...(current.vitals || [])],
          }
        });
      }
    } catch (error: any) {
      set({ error: error.message });
      throw error;
    }
  },

  createOrder: async (visitId, order) => {
    try {
      const res = await apiClient.post(`/visits/${visitId}/orders`, order);
      const current = get().selectedVisit;
      if (current && current.id === visitId) {
        set({
          selectedVisit: {
            ...current,
            orders: [res.data.data, ...(current.orders || [])],
          }
        });
      }
    } catch (error: any) {
      set({ error: error.message });
      throw error;
    }
  },

  closeVisit: async (visitId) => {
    set({ isLoading: true, error: null });
    try {
      await apiClient.post(`/visits/${visitId}/close`);
      const current = get().selectedVisit;
      
      // Also complete the queue entry if exists
      if (current && current.queue_entry_id) {
        try {
          await apiClient.post(`/queue/complete/${current.queue_entry_id}`);
        } catch (queueErr) {
          console.warn("Failed to complete queue entry:", queueErr);
        }
      }

      if (current && current.id === visitId) {
        set({
          selectedVisit: { ...current, status: 'COMPLETED' },
          isLoading: false
        });
      } else {
        set({ isLoading: false });
      }
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },
}));
