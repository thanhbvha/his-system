import { create } from 'zustand';
import apiClient from '@/lib/apiClient';

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
  patient: { id: string; full_name: string; dob: string; gender: string };
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
  queue_entry_id?: string;
  chief_complaint?: string;
}

interface VisitState {
  worklist: Visit[];
  selectedVisit: Visit | null;
  isLoading: boolean;

  fetchWorklist: () => Promise<void>;
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

  fetchWorklist: async () => {
    set({ isLoading: true });
    try {
      // The API should automatically filter by the currently logged in doctor
      // For now, we fetch today's active visits
      const res = await apiClient.get('/visits?status=IN_PROGRESS,WAITING');
      set({ worklist: res.data.data.items || [] });
    } catch (error) {
      console.error("Failed to fetch worklist:", error);
    } finally {
      set({ isLoading: false });
    }
  },

  fetchVisitDetail: async (visitId: string) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/visits/${visitId}`);
      set({ selectedVisit: res.data.data });
    } catch (error) {
      console.error("Failed to fetch visit detail:", error);
    } finally {
      set({ isLoading: false });
    }
  },

  createVisit: async (payload: CreateVisitPayload) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.post('/visits', payload);
      const newVisit = res.data.data;
      // Refresh worklist after creating
      await get().fetchWorklist();
      return newVisit;
    } catch (error) {
      console.error("Failed to create visit:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  recordVitals: async (visitId: string, vitals: Partial<VisitVital>) => {
    set({ isLoading: true });
    try {
      await apiClient.post(`/visits/${visitId}/vitals`, vitals);
      await get().fetchVisitDetail(visitId); // Refresh details
    } catch (error) {
      console.error("Failed to record vitals:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  createOrder: async (visitId: string, order: { order_type: string; details: string }) => {
    set({ isLoading: true });
    try {
      await apiClient.post(`/visits/${visitId}/orders`, order);
      await get().fetchVisitDetail(visitId); // Refresh details
    } catch (error) {
      console.error("Failed to create order:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  closeVisit: async (visitId: string) => {
    set({ isLoading: true });
    try {
      await apiClient.put(`/visits/${visitId}/status`, { status: "COMPLETED" });
      set({ selectedVisit: null });
      await get().fetchWorklist();
    } catch (error) {
      console.error("Failed to close visit:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  }
}));
