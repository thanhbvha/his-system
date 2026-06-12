import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

export const LandingPage = () => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col items-center justify-center text-center py-20">
      <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl mb-6 text-slate-900">
        {t("home.heroTitle")}
      </h1>
      <p className="text-xl text-slate-500 max-w-2xl mb-8">
        {t("home.heroSubtitle")}
      </p>
      <Link to="/login">
        <Button size="lg" className="text-lg px-8">
          {t("nav.book")}
        </Button>
      </Link>
    </div>
  );
};
