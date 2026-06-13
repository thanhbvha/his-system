import { create } from "zustand";

export interface Patient {
  id: string;
  name: string;
  email: string;
}

interface AuthState {
  token: string | null;
  patient: Patient | null;
  setAuth: (token: string, patient: Patient) => void;
  setToken: (token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  patient: null,
  setAuth: (token, patient) => set({ token, patient }),
  setToken: (token) => set({ token }),
  clearAuth: () => set({ token: null, patient: null }),
}));
