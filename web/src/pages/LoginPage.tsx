import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";

export const LoginPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const setAuth = useAuthStore((s) => s.setAuth);

  const handleMockLogin = () => {
    // Mock login for Sprint 1
    setAuth("dummy_patient_token", { id: "p1", name: "Nguyen Van A", email: "a@patient.his" });
    navigate("/account");
  };

  return (
    <div className="max-w-md mx-auto mt-20 p-6 bg-white rounded-lg shadow-sm border">
      <h2 className="text-2xl font-bold mb-6 text-center">{t("common.login")}</h2>
      <p className="text-sm text-slate-500 mb-6 text-center">
        This is a mock login for Sprint 1. Click below to simulate login.
      </p>
      <Button className="w-full" onClick={handleMockLogin}>
        Mock Login as Patient
      </Button>
    </div>
  );
};
