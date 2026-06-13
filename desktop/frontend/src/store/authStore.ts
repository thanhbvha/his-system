import { create } from "zustand";

export interface User {
  id: string;
  name: string;
  email: string;
}

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  role: "admin" | "doctor" | "nurse" | "receptionist" | "pharmacist" | null;
  setAuth: (token: string, refreshToken: string, user: User, role: AuthState["role"]) => void;
  setToken: (token: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  refreshToken: null,
  user: null,
  role: null,
  setAuth: (token, refreshToken, user, role) => set({ token, refreshToken, user, role }),
  setToken: (token) => set({ token }),
  clearAuth: () => set({ token: null, refreshToken: null, user: null, role: null }),
}));
