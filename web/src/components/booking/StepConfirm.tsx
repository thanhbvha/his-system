import { useState } from "react";
import { useBookingStore } from "@/store/bookingStore";
import { usePublicStore } from "@/store/publicStore";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

export const StepConfirm = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { 
    selectedService, selectedDoctor, selectedDate, selectedSlot, note, 
    setNote, setStep, bookAppointment, isBooking, fetchSlots, reset
  } = useBookingStore();
  
  const { services, doctors } = usePublicStore();
  const [error, setError] = useState("");

  const service = services.find(s => s.id === selectedService);
  const doctor = doctors.find(d => d.id === selectedDoctor);
  // Slot details have to be found inside availableSlots if we want start_time, 
  // but it's okay to just submit slot_id.
  const { availableSlots } = useBookingStore.getState();
  const slot = availableSlots.find(s => s.id === selectedSlot);

  const handleConfirm = async () => {
    setError("");
    try {
      await bookAppointment();
      reset();
      navigate('/my-appointments');
    } catch (err: any) {
      if (err.response?.status === 409) {
        setError("Slot này vừa được người khác đặt. Vui lòng chọn lại giờ khác.");
        fetchSlots(selectedDoctor!, selectedDate!);
        setTimeout(() => {
          setStep(3);
        }, 3000);
      } else {
        setError(err.response?.data?.message || "Đã xảy ra lỗi khi đặt lịch.");
      }
    }
  };

  return (
    <div className="flex flex-col h-full max-w-2xl mx-auto w-full">
      <div className="flex-1 space-y-6">
        {error && (
          <div className="bg-destructive/10 text-destructive px-4 py-3 rounded-md">
            {error}
          </div>
        )}

        <Card>
          <CardContent className="pt-6 space-y-4">
            <h3 className="text-lg font-semibold border-b pb-2 mb-4">Thông tin đặt khám</h3>
            
            <div className="grid grid-cols-3 gap-4">
              <div className="font-medium text-slate-500">Dịch vụ:</div>
              <div className="col-span-2 font-semibold">{service?.name}</div>
            </div>
            
            <div className="grid grid-cols-3 gap-4">
              <div className="font-medium text-slate-500">Bác sĩ:</div>
              <div className="col-span-2">{doctor?.full_name}</div>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="font-medium text-slate-500">Ngày khám:</div>
              <div className="col-span-2">{selectedDate}</div>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="font-medium text-slate-500">Giờ khám:</div>
              <div className="col-span-2">{slot?.start_time} - {slot?.end_time}</div>
            </div>

            <div className="pt-4">
              <Label htmlFor="note">Ghi chú cho bác sĩ (Tùy chọn)</Label>
              <textarea 
                id="note"
                className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm mt-2" 
                rows={3} 
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Triệu chứng bệnh, tiền sử bệnh lý..."
              />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="mt-8 flex justify-between">
        <Button variant="outline" onClick={() => setStep(3)} disabled={isBooking}>
          Quay lại
        </Button>
        <Button onClick={handleConfirm} disabled={isBooking} className="w-40">
          {isBooking ? t("common.loading") : "Xác nhận đặt lịch"}
        </Button>
      </div>
    </div>
  );
};
