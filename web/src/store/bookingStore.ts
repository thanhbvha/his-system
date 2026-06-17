import { create } from "zustand";
import apiClient from '@/lib/apiClient';

export interface Slot {
  id: string;
  start_time: string;
  end_time: string;
  is_booked: boolean;
}

interface BookingState {
  step: 1 | 2 | 3 | 4;
  selectedService: string | null;
  selectedDoctor: string | null;
  selectedDate: string | null;
  selectedSlot: string | null;
  note: string;
  
  availableSlots: Slot[];
  isLoadingSlots: boolean;
  isBooking: boolean;

  setStep: (step: 1 | 2 | 3 | 4) => void;
  setService: (serviceId: string) => void;
  setDoctor: (doctorId: string) => void;
  setDate: (date: string) => void;
  setSlot: (slotId: string) => void;
  setNote: (note: string) => void;
  
  fetchSlots: (doctorId: string, date: string) => Promise<void>;
  bookAppointment: () => Promise<void>;
  reset: () => void;
}

export const useBookingStore = create<BookingState>((set, get) => ({
  step: 1,
  selectedService: null,
  selectedDoctor: null,
  selectedDate: null,
  selectedSlot: null,
  note: "",
  
  availableSlots: [],
  isLoadingSlots: false,
  isBooking: false,

  setStep: (step) => set({ step }),
  setService: (serviceId) => set({ selectedService: serviceId, selectedDoctor: null, selectedSlot: null, availableSlots: [] }),
  setDoctor: (doctorId) => set({ selectedDoctor: doctorId, selectedSlot: null, availableSlots: [] }),
  setDate: (date) => set({ selectedDate: date, selectedSlot: null }),
  setSlot: (slotId) => set({ selectedSlot: slotId }),
  setNote: (note) => set({ note }),

  fetchSlots: async (doctorId: string, date: string) => {
    set({ isLoadingSlots: true });
    try {
      const response = await apiClient.get(`/appointments/slots?doctor_id=${doctorId}&date=${date}`);
      set({ availableSlots: response.data.data });
    } catch (error) {
      console.error('Failed to fetch slots', error);
      set({ availableSlots: [] });
    } finally {
      set({ isLoadingSlots: false });
    }
  },

  bookAppointment: async () => {
    const { selectedService, selectedDoctor, selectedDate, selectedSlot, note } = get();
    set({ isBooking: true });
    try {
      await apiClient.post('/appointments', {
        service_id: selectedService,
        doctor_id: selectedDoctor,
        date: selectedDate,
        slot_id: selectedSlot,
        note: note
      });
    } finally {
      set({ isBooking: false });
    }
  },

  reset: () => set({
    step: 1,
    selectedService: null,
    selectedDoctor: null,
    selectedDate: null,
    selectedSlot: null,
    note: "",
    availableSlots: [],
  }),
}));
