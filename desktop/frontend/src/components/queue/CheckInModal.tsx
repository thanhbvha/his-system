import React, { useState, useEffect } from "react";
import { Modal, Steps, Select, Button, notification, Typography, Card } from "antd";
import { useTranslation } from "react-i18next";
import { PatientSearchModal } from "@/components/patient/PatientSearchModal";
import { usePublicStore } from "@/store/publicStore";
import { useQueueStore } from "@/store/queueStore";

const { Text } = Typography;

interface CheckInModalProps {
  visible: boolean;
  onClose: () => void;
}

export const CheckInModal: React.FC<CheckInModalProps> = ({ visible, onClose }) => {
  const { t } = useTranslation();
  const [currentStep, setCurrentStep] = useState(0);
  const [selectedPatient, setSelectedPatient] = useState<any>(null);
  const [selectedService, setSelectedService] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  const { services, fetchServices } = usePublicStore();
  const { checkIn } = useQueueStore();

  useEffect(() => {
    if (visible && services.length === 0) {
      fetchServices();
    }
    if (!visible) {
      // Reset state on close
      setCurrentStep(0);
      setSelectedPatient(null);
      setSelectedService("");
    }
  }, [visible, services.length, fetchServices]);

  const handlePatientSelect = (patient: any) => {
    setSelectedPatient(patient);
    setCurrentStep(1);
  };

  const handleSubmit = async () => {
    if (!selectedPatient || !selectedService) return;
    
    setIsSubmitting(true);
    try {
      const entry = await checkIn({
        patient_id: selectedPatient.id,
        service_type: selectedService,
      });
      notification.success({
        message: t("queue.checkinSuccess", { number: entry.queue_number }),
        description: `${selectedPatient.full_name} - ${selectedService}`,
        duration: 5,
      });
      onClose();
    } catch (error: any) {
      notification.error({
        message: t("common.error"),
        description: error.response?.data?.message || "Check-in failed",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const steps = [
    {
      title: t("queue.searchPatient"),
      content: (
        <div className="py-12 flex flex-col items-center justify-center">
          <Text className="text-gray-500 mb-4">Vui lòng tìm kiếm và chọn bệnh nhân để tiếp tục</Text>
          <Button type="primary" size="large" onClick={() => setIsSearchOpen(true)}>
            {t("queue.searchPatient")}
          </Button>
          <PatientSearchModal
            open={isSearchOpen}
            onClose={() => setIsSearchOpen(false)}
            onSelect={handlePatientSelect}
          />
        </div>
      )
    },
    {
      title: t("queue.selectService"),
      content: (
        <div className="py-8 px-4 flex flex-col items-center">
          {selectedPatient && (
            <Card className="w-full max-w-md mb-6 bg-blue-50 border-blue-200">
              <Text className="block text-gray-500 mb-1">Bệnh nhân đã chọn:</Text>
              <Text strong className="text-lg">{selectedPatient.full_name}</Text>
              <Text className="block text-gray-500 mt-1">Mã BN: {selectedPatient.patient_code}</Text>
            </Card>
          )}
          
          <div className="w-full max-w-md">
            <Text className="block mb-2 font-medium">{t("queue.selectService")}</Text>
            <Select
              className="w-full"
              size="large"
              placeholder={t("queue.selectService")}
              value={selectedService || undefined}
              onChange={setSelectedService}
              options={services.map(s => ({ value: s.id, label: s.name }))}
            />
          </div>
        </div>
      )
    }
  ];

  return (
    <Modal
      title={<span className="text-xl font-bold">{t("queue.checkIn")}</span>}
      open={visible}
      onCancel={onClose}
      width={700}
      footer={
        currentStep === 1 ? (
          <div className="flex justify-between w-full">
            <Button onClick={() => setCurrentStep(0)}>Quay lại</Button>
            <Button 
              type="primary" 
              onClick={handleSubmit} 
              loading={isSubmitting}
              disabled={!selectedService}
              size="large"
            >
              Hoàn tất Check-in
            </Button>
          </div>
        ) : null
      }
      destroyOnClose
    >
      <div className="mt-4">
        <Steps current={currentStep} items={steps.map(s => ({ title: s.title }))} className="mb-6" />
        
        <div className="bg-white rounded-md min-h-[300px]">
          {steps[currentStep].content}
        </div>
      </div>
    </Modal>
  );
};
