import { useTranslation } from "react-i18next";

export const RegisterPage = () => {
  const { t } = useTranslation();

  return (
    <div className="max-w-md mx-auto mt-20 p-6 bg-white rounded-lg shadow-sm border text-center">
      <h2 className="text-2xl font-bold mb-4">{t("common.register")}</h2>
      <p className="text-slate-500">{t("common.comingSoon")}</p>
    </div>
  );
};
