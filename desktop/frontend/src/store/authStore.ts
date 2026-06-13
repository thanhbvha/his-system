import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface User {
  id: string;
  name: string;
  email?: string;
  username?: string;
  mfa_enabled?: boolean;
  role_ids?: string[];
}

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  role: "admin" | "doctor" | "nurse" | "receptionist" | "pharmacist" | null;
  setAuth: (token: string, refreshToken: string, user: User, role: AuthState["role"]) => void;
  setToken: (token: string) => void;
  setTokens: (token: string, refreshToken: string) => void;
  updateAuthUser: (user: Partial<User>) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      role: null,
      setAuth: (token, refreshToken, user, role) => set({ token, refreshToken, user, role }),
      setToken: (token) => set({ token }),
      setTokens: (token, refreshToken) => set({ token, refreshToken }),
      updateAuthUser: (updates) => set((state) => ({ user: state.user ? { ...state.user, ...updates } : null })),
      clearAuth: () => set({ token: null, refreshToken: null, user: null, role: null }),
    }),
    {
      name: "auth-storage", // name of the item in the storage (must be unique)
    }
  )
);
