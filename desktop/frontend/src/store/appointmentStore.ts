import { create } from 'zustand';
import apiClient from '@/lib/apiClient';

export interface Appointment {
  id: string;
  patient_id: string;
  doctor_id: string;
  service_id: string;
  slot_id: string;
  status: 'PENDING' | 'CONFIRMED' | 'CHECKED_IN' | 'COMPLETED' | 'CANCELLED';
  note?: string;
  scheduled_at: string;
  // Included relations for frontend display
  patient?: { full_name: string; phone_masked: string; patient_code: string };
  doctor?: { full_name: string };
  service?: { name: string };
}

export interface AppointmentSlot {
  id: string;
  doctor_id: string;
  date: string; // YYYY-MM-DD
  start_time: string; // HH:mm
  end_time: string; // HH:mm
  is_booked: boolean;
}

interface AppointmentStore {
  appointments: Appointment[];
  availableSlots: AppointmentSlot[];
  isLoading: boolean;
  fetchByDate: (date: string) => Promise<void>;
  fetchSlots: (doctorId: string, date: string) => Promise<void>;
  bookAppointment: (data: { doctor_id: string, service_id: string, slot_id: string, note?: string, patient_id?: string }) => Promise<void>;
  updateStatus: (id: string, status: string) => Promise<void>;
}

export const useAppointmentStore = create<AppointmentStore>((set) => ({
  appointments: [],
  availableSlots: [],
  isLoading: false,

  fetchByDate: async (date: string) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/appointments?date=${date}`);
      set({ appointments: res.data.data.items || [] });
    } catch (error) {
      console.error("Failed to fetch appointments:", error);
      set({ appointments: [] });
    } finally {
      set({ isLoading: false });
    }
  },

  fetchSlots: async (doctorId: string, date: string) => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get(`/appointments/slots?doctor_id=${doctorId}&date=${date}`);
      set({ availableSlots: res.data.data || [] });
    } catch (error) {
      console.error("Failed to fetch slots:", error);
      set({ availableSlots: [] });
    } finally {
      set({ isLoading: false });
    }
  },

  bookAppointment: async (data) => {
    set({ isLoading: true });
    try {
      await apiClient.post('/appointments', data);
    } catch (error) {
      console.error("Failed to book appointment:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  },

  updateStatus: async (id: string, status: string) => {
    set({ isLoading: true });
    try {
      await apiClient.put(`/appointments/${id}`, { status });
      // Update local state without refetching if successful
      set((state) => ({
        appointments: state.appointments.map(app => 
          app.id === id ? { ...app, status: status as any } : app
        )
      }));
    } catch (error) {
      console.error("Failed to update status:", error);
      throw error;
    } finally {
      set({ isLoading: false });
    }
  }
}));
