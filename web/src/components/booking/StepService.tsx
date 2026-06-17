import { useEffect } from "react";
import { usePublicStore } from "@/store/publicStore";
import { useBookingStore } from "@/store/bookingStore";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export const StepService = () => {
  const { services, fetchServices, isLoading } = usePublicStore();
  const { selectedService, setService, setStep } = useBookingStore();

  useEffect(() => {
    if (services.length === 0) {
      fetchServices();
    }
  }, [fetchServices, services.length]);

  const handleNext = () => {
    if (selectedService) {
      setStep(2);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1">
        {isLoading ? (
          <p className="text-center text-slate-500 my-8">Đang tải danh sách dịch vụ...</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {services.map((srv) => (
              <Card 
                key={srv.id} 
                className={`cursor-pointer transition-all hover:border-primary ${selectedService === srv.id ? 'border-primary bg-primary/5 shadow-md' : ''}`}
                onClick={() => setService(srv.id)}
              >
                <CardHeader className="pb-2">
                  <CardTitle className="text-lg">{srv.name}</CardTitle>
                  <CardDescription>{srv.duration_minutes} phút</CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="font-semibold text-primary">{srv.price.toLocaleString()} VNĐ</p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      <div className="mt-8 flex justify-end">
        <Button onClick={handleNext} disabled={!selectedService}>
          Tiếp tục
        </Button>
      </div>
    </div>
  );
};
