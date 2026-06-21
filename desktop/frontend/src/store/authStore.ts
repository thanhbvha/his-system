import { create } from "zustand";
import { persist } from "zustand/middleware";

export function isTokenExpired(token: string | null): boolean {
  if (!token) return true;
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return false; // Encrypted token, assume valid until backend returns 401
    const payloadBase64 = parts[1];
    let base64 = payloadBase64.replace(/-/g, '+').replace(/_/g, '/');
    while (base64.length % 4) {
      base64 += '=';
    }
    const decodedStr = atob(base64);
    
    // Instead of JSON.parse which crashes on raw binary AES-GCM payloads,
    // we use a regex to safely extract the 'exp' claim.
    const match = decodedStr.match(/"exp"\s*:\s*(\d+)/);
    if (match && match[1]) {
      const exp = parseInt(match[1], 10);
      return (Date.now() / 1000) >= (exp - 5);
    }
    
    // For encrypted tokens without standard JWT parts, we default to false 
    // to let the backend dictate expiration via 401 responses.
    return false;
  } catch (e) {
    console.error("isTokenExpired error:", e);
    // If we can't parse it, return false so we don't spam the backend
    // The backend JWT middleware will catch real expirations anyway.
    return false;
  }
}

export interface User {
  id: string;
  name: string;
  email?: string;
  username?: string;
  mfa_enabled?: boolean;
  preferred_language?: string;
  role_ids?: string[];
  roles?: string[];
}

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  role: "admin" | "doctor" | "nurse" | "receptionist" | "pharmacist" | null;
  roomId: string | null;
  setAuth: (token: string, refreshToken: string, user: User, role: AuthState["role"]) => void;
  setToken: (token: string) => void;
  setTokens: (token: string, refreshToken: string) => void;
  setRoomId: (roomId: string) => void;
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
      roomId: "global_reception", // Default to global_reception for Receptionists
      setAuth: (token, refreshToken, user, role) => set({ token, refreshToken, user, role }),
      setToken: (token) => set({ token }),
      setTokens: (token, refreshToken) => set({ token, refreshToken }),
      setRoomId: (roomId) => set({ roomId }),
      updateAuthUser: (updates) => set((state) => ({ user: state.user ? { ...state.user, ...updates } : null })),
      clearAuth: () => {
        set({ token: null, refreshToken: null, user: null, role: null, roomId: null });
        localStorage.clear();
        sessionStorage.clear();
      },
    }),
    {
      name: "auth-storage", // name of the item in the storage (must be unique)
    }
  )
);
