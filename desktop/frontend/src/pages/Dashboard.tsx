import { useTranslation } from "react-i18next";

export const Dashboard = () => {
  const { t } = useTranslation();

  return (
    <div>
      <h2>{t("nav.dashboard")}</h2>
      <p>{t("common.comingSoon")}</p>
    </div>
  );
};
