import { useEffect } from "react";
import { usePublicStore } from "@/store/publicStore";
import { useBookingStore } from "@/store/bookingStore";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export const StepDoctor = () => {
  const { doctors, fetchDoctors, isLoading } = usePublicStore();
  const { selectedService, selectedDoctor, setDoctor, setStep } = useBookingStore();

  useEffect(() => {
    if (selectedService) {
      fetchDoctors(selectedService);
    }
  }, [fetchDoctors, selectedService]);

  const handleNext = () => {
    if (selectedDoctor) {
      setStep(3);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1">
        {isLoading ? (
          <p className="text-center text-slate-500 my-8">Đang tải danh sách bác sĩ...</p>
        ) : doctors.length === 0 ? (
          <p className="text-center text-slate-500 my-8">Không có bác sĩ nào cho dịch vụ này.</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {doctors.map((doc) => (
              <Card 
                key={doc.id} 
                className={`cursor-pointer transition-all hover:border-primary ${selectedDoctor === doc.id ? 'border-primary bg-primary/5 shadow-md' : ''}`}
                onClick={() => setDoctor(doc.id)}
              >
                <CardHeader>
                  <CardTitle className="text-lg">{doc.full_name}</CardTitle>
                  <CardDescription>{doc.specialty}</CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        )}
      </div>

      <div className="mt-8 flex justify-between">
        <Button variant="outline" onClick={() => setStep(1)}>
          Quay lại
        </Button>
        <Button onClick={handleNext} disabled={!selectedDoctor}>
          Tiếp tục
        </Button>
      </div>
    </div>
  );
};
