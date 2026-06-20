import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Row, Col, Typography, Spin, Alert } from "antd";
import { PlusOutlined, SyncOutlined } from "@ant-design/icons";
import { useQueueStore } from "@/store/queueStore";
import { useAuthStore } from "@/store/authStore";
import { useQueueWS } from "@/components/ws/useQueueWS";
import { QueueColumn } from "./QueueColumn";
import { CheckInModal } from "./CheckInModal";
import { usePublicStore } from "@/store/publicStore";

const { Title } = Typography;

export const QueueDashboard: React.FC = () => {
  const { t } = useTranslation();
  const token = useAuthStore(state => state.token);
  const { entries, isLoading, fetchQueue, callNext, skip } = useQueueStore();
  const { services, fetchServices } = usePublicStore();
  
  const [isCheckInVisible, setIsCheckInVisible] = useState(false);

  // Initialize WS connection
  useQueueWS(token);

  useEffect(() => {
    fetchQueue();
    if (services.length === 0) {
      fetchServices();
    }
  }, [fetchQueue, fetchServices, services.length]);

  // Group entries by service type
  const entriesByService = entries.reduce((acc, entry) => {
    if (!acc[entry.service_type]) {
      acc[entry.service_type] = [];
    }
    acc[entry.service_type].push(entry);
    return acc;
  }, {} as Record<string, typeof entries>);

  return (
    <div className="h-full flex flex-col p-6 bg-gray-50 min-h-screen">
      <div className="flex justify-between items-center mb-6">
        <div>
          <Title level={2} style={{ margin: 0 }}>{t("queue.title")}</Title>
          <div className="text-gray-500 mt-1">
            Realtime WebSocket Sync <span className="inline-block w-2 h-2 rounded-full bg-green-500 ml-1 animate-pulse"></span>
          </div>
        </div>
        <div className="flex gap-3">
          <Button icon={<SyncOutlined />} onClick={() => fetchQueue()}>Làm mới</Button>
          <Button 
            type="primary" 
            size="large" 
            icon={<PlusOutlined />} 
            onClick={() => setIsCheckInVisible(true)}
            className="shadow-md"
          >
            {t("queue.checkIn")}
          </Button>
        </div>
      </div>

      {isLoading && entries.length === 0 ? (
        <div className="flex-1 flex justify-center items-center">
          <Spin size="large" />
        </div>
      ) : (
        <div className="flex-1 overflow-x-auto pb-4">
          <div className="flex gap-6 h-full min-w-max">
            {services.map(service => (
              <div key={service.id} className="w-[320px] shrink-0 h-[calc(100vh-180px)]">
                <QueueColumn 
                  serviceType={service.name} 
                  entries={entriesByService[service.id] || []}
                  onCallNext={callNext}
                  onSkip={skip}
                />
              </div>
            ))}
            
            {services.length === 0 && !isLoading && (
              <Alert message="Chưa có dịch vụ nào được định nghĩa trong hệ thống." type="warning" showIcon className="w-full" />
            )}
          </div>
        </div>
      )}

      <CheckInModal 
        visible={isCheckInVisible} 
        onClose={() => setIsCheckInVisible(false)} 
      />
    </div>
  );
};
