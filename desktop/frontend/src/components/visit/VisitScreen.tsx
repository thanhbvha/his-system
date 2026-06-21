import React, { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Tabs, Button, Spin, Tag, Typography, Layout, Space } from "antd";
import { ArrowLeftOutlined, CheckCircleOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import { useVisitStore } from "@/store/visitStore";

import { VitalsForm } from "./VitalsForm";
import { VitalsHistory } from "./VitalsHistory";
import { OrdersPanel } from "./OrdersPanel";
import { PatientHistoryTab } from "./PatientHistoryTab";

const { Header, Content } = Layout;
const { Title, Text } = Typography;

export const VisitScreen: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  
  const { selectedVisit, isLoading, fetchVisitDetail, closeVisit } = useVisitStore();

  useEffect(() => {
    if (id) {
      fetchVisitDetail(id);
    }
  }, [id, fetchVisitDetail]);

  if (isLoading && !selectedVisit) {
    return (
      <div className="flex justify-center items-center h-screen bg-gray-50">
        <Spin size="large" />
      </div>
    );
  }

  if (!selectedVisit) {
    return (
      <div className="p-8 text-center text-red-500">
        Không tìm thấy thông tin ca khám!
      </div>
    );
  }

  const handleCloseVisit = async () => {
    try {
      await closeVisit(selectedVisit.id);
      navigate('/visits');
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <Layout className="min-h-screen bg-gray-50">
      <Header style={{ background: '#fff', padding: '0 24px', height: 'auto', minHeight: '64px', lineHeight: 'normal' }} className="border-b flex items-center justify-between sticky top-0 z-10 shadow-sm py-3">
        <div className="flex items-center gap-4">
          <Button 
            type="text" 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/visits')}
          />
          <div className="flex flex-col justify-center">
            <Title level={4} style={{ margin: 0 }}>{selectedVisit.patient?.full_name}</Title>
            <Text type="secondary" className="text-sm">
              Mã BN: {selectedVisit.patient?.patient_code || selectedVisit.patient?.id?.substring(0, 8)} | Bác sĩ: {selectedVisit.doctor?.full_name || 'Đang chờ phân công'}
            </Text>
          </div>
        </div>
        <Space>
          <Tag color={selectedVisit.status === 'COMPLETED' ? 'green' : 'blue'} className="text-sm py-1 px-3">
            {selectedVisit.status}
          </Tag>
          {selectedVisit.status !== 'COMPLETED' && (
            <Button 
              type="primary" 
              icon={<CheckCircleOutlined />} 
              onClick={handleCloseVisit}
              loading={isLoading}
            >
              {t("visit.closeVisit", "Kết thúc khám")}
            </Button>
          )}
        </Space>
      </Header>

      <Content className="p-6 max-w-7xl mx-auto w-full">
        <Tabs 
          defaultActiveKey="1" 
          type="card"
          items={[
            {
              key: "1",
              label: t("visit.vitals", "Sinh hiệu"),
              children: (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="md:col-span-2">
                    {selectedVisit.status !== 'COMPLETED' && <VitalsForm visitId={selectedVisit.id} />}
                  </div>
                  <div className="md:col-span-1">
                    <VitalsHistory vitals={selectedVisit.vitals || []} />
                  </div>
                </div>
              ),
            },
            {
              key: "2",
              label: t("visit.orders", "Chỉ định"),
              children: <OrdersPanel visitId={selectedVisit.id} orders={selectedVisit.orders || []} />,
            },
            {
              key: "3",
              label: t("visit.history", "Lịch sử khám"),
              children: <PatientHistoryTab patientId={selectedVisit.patient?.id} />,
            }
          ]}
        />
      </Content>
    </Layout>
  );
};
