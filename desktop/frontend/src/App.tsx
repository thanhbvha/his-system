import { HashRouter, Routes, Route, Navigate } from "react-router-dom";
import { RoleLayout } from "@/layouts/RoleLayout";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { Dashboard } from "@/pages/Dashboard";
import { Login } from "@/pages/Login";

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        
        <Route element={<ProtectedRoute />}>
          <Route element={<RoleLayout />}>
            <Route path="/" element={<Dashboard />} />
            {/* Additional routes will be added here in future sprints */}
          </Route>
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </HashRouter>
  );
}

export default App;
