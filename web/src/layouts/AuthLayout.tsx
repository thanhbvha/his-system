import { Outlet, Link, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/store/authStore";

export const AuthLayout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const clearAuth = useAuthStore((s) => s.clearAuth);

  const handleLogout = () => {
    clearAuth();
    navigate("/");
  };

  return (
    <div className="min-h-screen flex flex-col bg-slate-50">
      <header className="bg-white border-b h-16 flex items-center justify-between px-6">
        <div className="font-bold text-primary text-xl flex items-center gap-6">
          <Link to="/account">HIS Patient Portal</Link>
          <nav className="hidden md:flex space-x-4 text-sm font-medium">
            <Link to="/book" className="text-slate-600 hover:text-primary">{t("nav.book")}</Link>
            <Link to="/my-appointments" className="text-slate-600 hover:text-primary">{t("nav.myAppointments")}</Link>
            <Link to="/results" className="text-slate-600 hover:text-primary">{t("nav.results")}</Link>
          </nav>
        </div>
        <div className="space-x-4 flex items-center">
          <Button variant="ghost" onClick={() => i18n.changeLanguage(i18n.language === "vi" ? "en" : "vi")}>
            {i18n.language === "vi" ? "EN" : "VI"}
          </Button>
          <Button variant="outline" onClick={handleLogout}>{t("common.logout")}</Button>
        </div>
      </header>
      <main className="flex-1 container mx-auto p-6">
        <Outlet />
      </main>
    </div>
  );
};
