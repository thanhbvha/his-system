import { useTranslation } from "react-i18next";
import { useBookingStore } from "@/store/bookingStore";
import { StepService } from "@/components/booking/StepService";
import { StepDoctor } from "@/components/booking/StepDoctor";
import { StepSlot } from "@/components/booking/StepSlot";
import { StepConfirm } from "@/components/booking/StepConfirm";

export const BookingPage = () => {
  const { t } = useTranslation();
  const { step } = useBookingStore();

  const renderStep = () => {
    switch (step) {
      case 1: return <StepService />;
      case 2: return <StepDoctor />;
      case 3: return <StepSlot />;
      case 4: return <StepConfirm />;
      default: return <StepService />;
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 min-h-[400px]">
      <div className="mb-6 border-b pb-4">
        <h2 className="text-2xl font-bold">{t("nav.book")}</h2>
        <div className="flex gap-4 mt-4 text-sm text-slate-500">
          <span className={step >= 1 ? "text-primary font-medium" : ""}>1. Chọn dịch vụ</span>
          <span>&gt;</span>
          <span className={step >= 2 ? "text-primary font-medium" : ""}>2. Chọn bác sĩ</span>
          <span>&gt;</span>
          <span className={step >= 3 ? "text-primary font-medium" : ""}>3. Chọn ngày & giờ</span>
          <span>&gt;</span>
          <span className={step >= 4 ? "text-primary font-medium" : ""}>4. Xác nhận</span>
        </div>
      </div>
      
      {renderStep()}
    </div>
  );
};
