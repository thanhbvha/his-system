import React, { useEffect } from 'react';
import { Card, Spin, Alert, Timeline, Typography } from 'antd';
import { useTranslation } from "react-i18next";
import { usePatientStore } from '@/store/patientStore';
import dayjs from 'dayjs';

const { Text } = Typography;

interface PatientHistoryTabProps {
  patientId: string;
}

export const PatientHistoryTab: React.FC<PatientHistoryTabProps> = ({ patientId }) => {
  const { t } = useTranslation();
  const { selectedPatient, getPatientDetail, isLoading } = usePatientStore();

  useEffect(() => {
    if (patientId) {
      getPatientDetail(patientId);
    }
  }, [patientId, getPatientDetail]);

  if (isLoading) {
    return <Spin size="large" style={{ display: 'block', margin: '40px auto' }} />;
  }

  if (!selectedPatient) {
    return <Alert message={t("patients.notFound")} type="warning" />;
  }

  return (
    <Card bordered={false}>
      <Alert message="Chức năng xem chi tiết lịch sử khám (đơn thuốc, kết quả xét nghiệm cũ) sẽ được hoàn thiện trong Sprint 5." type="info" style={{ marginBottom: 24 }} />
      <Timeline>
        <Timeline.Item color="blue">
          <Text strong>{dayjs().format('DD/MM/YYYY')}</Text>
          <br />
          <Text>Lần khám hiện tại</Text>
        </Timeline.Item>
        {/* In reality we would map through selectedPatient's historical visits */}
        <Timeline.Item color="gray">
          <Text type="secondary">Chưa có dữ liệu lịch sử cũ trong hệ thống</Text>
        </Timeline.Item>
      </Timeline>
    </Card>
  );
};
