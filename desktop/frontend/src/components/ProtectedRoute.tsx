import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";

interface ProtectedRouteProps {
  roles?: string[];
}

export const ProtectedRoute = ({ roles }: ProtectedRouteProps) => {
  const { token, role } = useAuthStore();
  
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  
  if (roles && !roles.includes(role ?? "")) {
    return <Navigate to="/403" replace />; // Will just redirect to 403 or home
  }
  
  return <Outlet />;
};
