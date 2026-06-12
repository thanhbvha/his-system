import { useTranslation } from "react-i18next";

export const BookingPage = () => {
  const { t } = useTranslation();

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 min-h-[400px]">
      <h2 className="text-2xl font-bold mb-4">{t("nav.book")}</h2>
      <p className="text-slate-500">{t("common.comingSoon")} (Multi-step form goes here)</p>
    </div>
  );
};
