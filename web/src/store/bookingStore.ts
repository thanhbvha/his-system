import { create } from "zustand";

interface Slot {
  id: string;
  time: string;
}

interface PatientInfo {
  name: string;
  phone: string;
  dob: string;
}

interface BookingState {
  step: 1 | 2 | 3 | 4;
  selectedDepartment: string | null;
  selectedDoctor: string | null;
  selectedSlot: Slot | null;
  patientInfo: PatientInfo | null;
  setStep: (step: 1 | 2 | 3 | 4) => void;
  setDepartment: (dept: string) => void;
  setDoctor: (doc: string) => void;
  setSlot: (slot: Slot) => void;
  setPatientInfo: (info: PatientInfo) => void;
  reset: () => void;
}

export const useBookingStore = create<BookingState>((set) => ({
  step: 1,
  selectedDepartment: null,
  selectedDoctor: null,
  selectedSlot: null,
  patientInfo: null,
  setStep: (step) => set({ step }),
  setDepartment: (dept) => set({ selectedDepartment: dept }),
  setDoctor: (doc) => set({ selectedDoctor: doc }),
  setSlot: (slot) => set({ selectedSlot: slot }),
  setPatientInfo: (info) => set({ patientInfo: info }),
  reset: () => set({
    step: 1,
    selectedDepartment: null,
    selectedDoctor: null,
    selectedSlot: null,
    patientInfo: null,
  }),
}));
