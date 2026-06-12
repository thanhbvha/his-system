import { Outlet, Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { useTranslation } from "react-i18next";

export const PublicLayout = () => {
  const { t, i18n } = useTranslation();

  return (
    <div className="min-h-screen flex flex-col bg-slate-50">
      <header className="bg-white border-b h-16 flex items-center justify-between px-6">
        <div className="font-bold text-primary text-xl">HIS Patient Portal</div>
        <div className="space-x-4 flex items-center">
          <Button variant="ghost" onClick={() => i18n.changeLanguage(i18n.language === "vi" ? "en" : "vi")}>
            {i18n.language === "vi" ? "EN" : "VI"}
          </Button>
          <Link to="/login">
            <Button variant="outline">{t("common.login")}</Button>
          </Link>
          <Link to="/register">
            <Button>{t("common.register")}</Button>
          </Link>
        </div>
      </header>
      <main className="flex-1 container mx-auto p-6">
        <Outlet />
      </main>
      <footer className="bg-white border-t p-6 text-center text-sm text-slate-500">
        &copy; {new Date().getFullYear()} HIS System. All rights reserved.
      </footer>
    </div>
  );
};
