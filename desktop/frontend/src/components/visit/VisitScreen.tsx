import React, { useEffect } from 'react';
import { Card, Tabs, Typography, Button, Space, Tag, Spin } from 'antd';
import { useTranslation } from "react-i18next";
import { useVisitStore } from '@/store/visitStore';
import { VitalsForm } from './VitalsForm';
import { VitalsHistory } from './VitalsHistory';
import { OrdersPanel } from './OrdersPanel';
import { PatientHistoryTab } from './PatientHistoryTab';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

interface VisitScreenProps {
  visitId: string;
}

export const VisitScreen: React.FC<VisitScreenProps> = ({ visitId }) => {
  const { t } = useTranslation();
  const { selectedVisit, fetchVisitDetail, closeVisit, isLoading } = useVisitStore();

  useEffect(() => {
    if (visitId) {
      fetchVisitDetail(visitId);
    }
  }, [visitId, fetchVisitDetail]);

  if (isLoading && !selectedVisit) {
    return <div style={{ textAlign: 'center', padding: '50px' }}><Spin size="large" /></div>;
  }

  if (!selectedVisit) {
    return <div>{t('common.error', 'Không tìm thấy thông tin bệnh nhân')}</div>;
  }

  const patient = selectedVisit.patient;
  const age = dayjs().diff(dayjs(patient.dob), 'year');

  const getStatusColor = (status: string) => {
    if (status === 'WAITING') return 'warning';
    if (status === 'IN_PROGRESS') return 'processing';
    if (status === 'COMPLETED') return 'success';
    return 'default';
  };

  const items = [
    {
      key: 'vitals',
      label: t('visit.vitals', 'Sinh hiệu'),
      children: (
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <VitalsForm visitId={visitId} />
          <VitalsHistory vitals={selectedVisit.vitals || []} />
        </Space>
      ),
    },
    {
      key: 'orders',
      label: t('visit.orders', 'Chỉ định & Chẩn đoán'),
      children: <OrdersPanel visitId={visitId} orders={selectedVisit.orders || []} />,
    },
    {
      key: 'history',
      label: t('visit.history', 'Lịch sử khám'),
      children: <PatientHistoryTab patientId={patient.id} />,
    },
  ];

  return (
    <Card 
      title={
        <Space direction="vertical" size={0}>
          <Title level={4} style={{ margin: 0 }}>
            {patient.full_name} 
            <Tag color={getStatusColor(selectedVisit.status)} style={{ marginLeft: 16 }}>
              {selectedVisit.status}
            </Tag>
          </Title>
          <Text type="secondary">
            {t('patients.age', 'Tuổi')}: {age} | {t('patients.gender', 'Giới tính')}: {patient.gender} | ID: {patient.id}
          </Text>
        </Space>
      }
      extra={
        <Space>
          {selectedVisit.status !== 'COMPLETED' && (
            <Button 
              type="primary" 
              danger 
              onClick={() => closeVisit(visitId)}
            >
              {t('visit.closeVisit', 'Kết thúc khám')}
            </Button>
          )}
        </Space>
      }
    >
      <Tabs defaultActiveKey="vitals" items={items} />
    </Card>
  );
};
