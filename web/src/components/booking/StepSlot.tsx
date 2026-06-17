import { useEffect } from "react";
import { useBookingStore } from "@/store/bookingStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export const StepSlot = () => {
  const { selectedDoctor, selectedDate, selectedSlot, availableSlots, isLoadingSlots, setDate, setSlot, setStep, fetchSlots } = useBookingStore();

  useEffect(() => {
    if (selectedDoctor && selectedDate) {
      fetchSlots(selectedDoctor, selectedDate);
    }
  }, [selectedDoctor, selectedDate, fetchSlots]);

  const handleNext = () => {
    if (selectedSlot) {
      setStep(4);
    }
  };

  // Lấy ngày hiện tại ở định dạng YYYY-MM-DD
  const today = new Date().toISOString().split('T')[0];

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 space-y-6">
        <div>
          <label className="block text-sm font-medium mb-2">Chọn ngày khám</label>
          <Input 
            type="date" 
            min={today}
            value={selectedDate || ''} 
            onChange={(e) => setDate(e.target.value)} 
            className="w-full md:w-1/3"
          />
        </div>

        {selectedDate && (
          <div>
            <label className="block text-sm font-medium mb-2">Khung giờ trống</label>
            {isLoadingSlots ? (
              <p className="text-slate-500">Đang tải khung giờ...</p>
            ) : availableSlots.length === 0 ? (
              <p className="text-slate-500">Không có khung giờ trống trong ngày này.</p>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
                {availableSlots.map(slot => (
                  <Button
                    key={slot.id}
                    variant={selectedSlot === slot.id ? "default" : "outline"}
                    className={selectedSlot === slot.id ? "ring-2 ring-primary ring-offset-2" : ""}
                    disabled={slot.is_booked}
                    onClick={() => setSlot(slot.id)}
                  >
                    {slot.start_time}
                  </Button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      <div className="mt-8 flex justify-between">
        <Button variant="outline" onClick={() => setStep(2)}>
          Quay lại
        </Button>
        <Button onClick={handleNext} disabled={!selectedSlot}>
          Tiếp tục
        </Button>
      </div>
    </div>
  );
};
