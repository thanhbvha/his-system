import { HashRouter, Routes, Route, Navigate } from "react-router-dom";
import { RoleLayout } from "@/layouts/RoleLayout";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { Dashboard } from "@/pages/Dashboard";
import { Login } from "@/pages/Login";
import { MFAPage } from "@/pages/MFAPage";
import { MFASetupPage } from "@/pages/MFASetupPage";

// Admin Pages
import { AdminDashboardPage } from "@/pages/admin/AdminDashboardPage";
import { UserListPage } from "@/pages/admin/UserListPage";
import { RolePermissionPage } from "@/pages/admin/RolePermissionPage";
import { DepartmentPage } from "@/pages/admin/DepartmentPage";

const ForbiddenPage = () => <div style={{ padding: 24, textAlign: "center", fontSize: 20 }}>403 Forbidden - Bạn không có quyền truy cập</div>;

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/mfa" element={<MFAPage />} />
        <Route path="/403" element={<ForbiddenPage />} />

        <Route element={<ProtectedRoute />}>
          <Route path="/mfa-setup" element={<MFASetupPage />} />

          <Route element={<RoleLayout />}>
            <Route path="/" element={<Dashboard />} />

            {/* Admin Routes */}
            <Route path="/admin" element={<ProtectedRoute roles={["admin"]} />}>
              <Route path="dashboard" element={<AdminDashboardPage />} />
              <Route path="users" element={<UserListPage />} />
              <Route path="roles" element={<RolePermissionPage />} />
              <Route path="departments" element={<DepartmentPage />} />
            </Route>

            {/* Additional routes will be added here in future sprints */}
          </Route>
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </HashRouter>
  );
}

export default App;
