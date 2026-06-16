import { Navigate, Outlet } from "react-router-dom";
import { useEffect } from "react";
import { useAuthStore, isTokenExpired } from "@/store/authStore";
import apiClient from "@/lib/apiClient";

interface ProtectedRouteProps {
  roles?: string[];
}

export const ProtectedRoute = ({ roles }: ProtectedRouteProps) => {
  const { token, role } = useAuthStore();
  
  useEffect(() => {
    const checkAuth = async () => {
      if (token && isTokenExpired(token)) {
        try {
          // Trigger a backend request to force 401 if refresh token is also expired.
          // The apiClient interceptor will catch the 401, attempt refresh, and logout if failed.
          await apiClient.get('/patients?limit=1');
        } catch (error) {
          // Interceptor handles logout, no need to do anything here
        }
      }
    };

    checkAuth();

    window.addEventListener('focus', checkAuth);
    const interval = setInterval(checkAuth, 60000); // Also check every minute

    return () => {
      window.removeEventListener('focus', checkAuth);
      clearInterval(interval);
    };
  }, [token]);

  if (!token) {
    return <Navigate to="/login" replace />;
  }
  
  if (roles && !roles.includes(role ?? "")) {
    return <Navigate to="/403" replace />; // Will just redirect to 403 or home
  }
  
  return <Outlet />;
};
