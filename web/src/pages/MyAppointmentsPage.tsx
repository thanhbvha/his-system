import { useTranslation } from "react-i18next";

export const MyAppointmentsPage = () => {
  const { t } = useTranslation();

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 min-h-[400px]">
      <h2 className="text-2xl font-bold mb-4">{t("nav.myAppointments")}</h2>
      <p className="text-slate-500">{t("common.comingSoon")}</p>
    </div>
  );
};
