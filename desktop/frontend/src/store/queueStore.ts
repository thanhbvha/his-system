import { create } from "zustand";
import apiClient from "@/lib/apiClient";

export interface QueuePatient {
  id: string;
  full_name: string;
  patient_code: string;
  phone?: string;
}

export interface QueueEntry {
  id: string;
  patient: QueuePatient;
  visit_id?: string;
  service_type: string;
  queue_number: string;
  status: 'WAITING' | 'CALLED' | 'IN_PROGRESS' | 'DONE' | 'SKIPPED';
  created_at: string;
  called_at?: string;
}

export interface QueueStats {
  waiting_count: number;
  called_count: number;
  avg_wait_minutes: number;
}

export interface WSEvent {
  type: string;
  payload: any;
  timestamp?: string;
}

interface CheckInPayload {
  patient_id: string;
  service_type: string;
  appointment_id?: string;
}

interface QueueState {
  entries: QueueEntry[];
  stats: QueueStats | null;
  isLoading: boolean;
  
  fetchQueue: () => Promise<void>;
  fetchStats: () => Promise<void>;
  checkIn: (payload: CheckInPayload) => Promise<QueueEntry>;
  callNext: (id: string) => Promise<void>;
  skip: (id: string) => Promise<void>;
  complete: (id: string) => Promise<void>;
  
  applyWSEvent: (event: WSEvent) => void;
}

export const useQueueStore = create<QueueState>((set, get) => ({
  entries: [],
  stats: null,
  isLoading: false,

  fetchQueue: async () => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get("/queue");
      set({ entries: res.data.data || [], isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      console.error("fetchQueue error:", error);
    }
  },

  fetchStats: async () => {
    try {
      const res = await apiClient.get("/queue/stats");
      set({ stats: res.data.data });
    } catch (error) {
      console.error("fetchStats error:", error);
    }
  },

  checkIn: async (payload) => {
    const res = await apiClient.post("/queue/checkin", payload);
    return res.data.data;
  },

  callNext: async (id) => {
    await apiClient.post(`/queue/call/${id}`);
  },

  skip: async (id) => {
    await apiClient.post(`/queue/skip/${id}`);
  },

  complete: async (id) => {
    await apiClient.post(`/queue/complete/${id}`);
  },

  applyWSEvent: (event: WSEvent) => {
    const { entries } = get();
    console.log("WS EVENT RECEIVED:", event);
    
    switch (event.type) {
      case "queue.sync":
        // Server sends full state on connect
        set({ entries: event.payload.entries || [] });
        break;
      
      case "queue.checked_in":
        set(state => ({ entries: [...state.entries, event.payload] }));
        break;

      case "queue.called":
      case "queue.skipped":
      case "queue.completed":
        const updatedEntry = event.payload;
        set(state => ({
          entries: state.entries.map(e => e.id === updatedEntry.id ? updatedEntry : e)
        }));
        break;

      case "queue.stats_updated":
        set({ stats: event.payload });
        break;
    }
  }
}));
