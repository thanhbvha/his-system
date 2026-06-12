import { create } from "zustand";

export interface User {
  id: string;
  name: string;
  email: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  role: "admin" | "doctor" | "nurse" | "receptionist" | "pharmacist" | null;
  setAuth: (token: string, user: User, role: AuthState["role"]) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  role: null,
  setAuth: (token, user, role) => set({ token, user, role }),
  clearAuth: () => set({ token: null, user: null, role: null }),
}));
