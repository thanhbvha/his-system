import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "react-router-dom";
import { usePublicStore } from "@/store/publicStore";

export const LandingPage = () => {
  const { t } = useTranslation();
  const { clinicInfo, services, doctors, fetchClinicInfo, fetchServices, fetchDoctors, isLoading } = usePublicStore();

  useEffect(() => {
    fetchClinicInfo();
    fetchServices();
    fetchDoctors();
  }, [fetchClinicInfo, fetchServices, fetchDoctors]);

  return (
    <div className="flex flex-col min-h-screen">
      {/* Hero Section */}
      <div className="flex flex-col items-center justify-center text-center py-20 bg-slate-50">
        <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl mb-6 text-slate-900">
          {clinicInfo?.name || t("home.heroTitle")}
        </h1>
        <p className="text-xl text-slate-500 max-w-2xl mb-8">
          {clinicInfo ? `${clinicInfo.address} - Tel: ${clinicInfo.phone}` : t("home.heroSubtitle")}
        </p>
        <Link to="/book">
          <Button size="lg" className="text-lg px-8">
            {t("nav.book")}
          </Button>
        </Link>
      </div>

      {/* Services Section */}
      <div className="py-16 px-8 max-w-7xl mx-auto w-full">
        <h2 className="text-3xl font-bold mb-8 text-center">Dịch vụ nổi bật</h2>
        {isLoading ? (
          <p className="text-center text-slate-500">Đang tải dữ liệu...</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {services.map((srv) => (
              <Card key={srv.id}>
                <CardHeader>
                  <CardTitle>{srv.name}</CardTitle>
                  <CardDescription>Thời gian: {srv.duration_minutes} phút</CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-lg font-semibold text-primary">
                    {srv.price.toLocaleString()} VNĐ
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Doctors Section */}
      <div className="py-16 px-8 max-w-7xl mx-auto w-full bg-slate-50">
        <h2 className="text-3xl font-bold mb-8 text-center">Đội ngũ Bác sĩ</h2>
        {isLoading ? (
          <p className="text-center text-slate-500">Đang tải dữ liệu...</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {doctors.map((doc) => (
              <Card key={doc.id}>
                <CardHeader>
                  <CardTitle className="text-lg">{doc.full_name}</CardTitle>
                  <CardDescription>{doc.specialty}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
