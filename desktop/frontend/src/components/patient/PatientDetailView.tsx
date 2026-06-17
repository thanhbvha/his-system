import React, { useEffect } from 'react';
import { Card, Descriptions, Tabs, Spin, Alert, Typography } from 'antd';
import { useTranslation } from "react-i18next";
import { usePatientStore } from '@/store/patientStore';

const { Title } = Typography;

interface PatientDetailViewProps {
  patientId: string;
}

export const PatientDetailView: React.FC<PatientDetailViewProps> = ({ patientId }) => {
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

  const items = [
    {
      key: '1',
      label: t("patients.generalInfo"),
      children: (
        <Descriptions bordered column={{ xxl: 2, xl: 2, lg: 2, md: 1, sm: 1, xs: 1 }}>
          <Descriptions.Item label={t("patients.patientCode")}>{selectedPatient.patient_code}</Descriptions.Item>
          <Descriptions.Item label={t("patients.fullName")}>{selectedPatient.full_name}</Descriptions.Item>
          <Descriptions.Item label={t("patients.dob")}>{selectedPatient.dob}</Descriptions.Item>
          <Descriptions.Item label={t("patients.gender")}>
            {selectedPatient.gender === 'MALE' ? t("patients.male") : selectedPatient.gender === 'FEMALE' ? t("patients.female") : t("patients.other")}
          </Descriptions.Item>
          {/* Note: This component is for staff, so we show the full phone/cccd instead of masked */}
          <Descriptions.Item label={t("patients.phone")}>{selectedPatient.phone || selectedPatient.phone_masked}</Descriptions.Item>
          <Descriptions.Item label={t("patients.cccd")}>{selectedPatient.cccd || 'N/A'}</Descriptions.Item>
          <Descriptions.Item label={t("patients.email")}>{selectedPatient.email || 'N/A'}</Descriptions.Item>
          <Descriptions.Item label={t("patients.address")}>{selectedPatient.address || 'N/A'}</Descriptions.Item>
        </Descriptions>
      ),
    },
    {
      key: '2',
      label: t("patients.bhyt"),
      children: <Alert message={t("patients.noBhyt")} type="info" />,
    },
    {
      key: '3',
      label: t("patients.history"),
      children: <Alert message={t("patients.noHistory")} type="info" />,
    },
  ];

  return (
    <Card title={t("patients.profileTitle", { name: selectedPatient.full_name })}>
      <Tabs defaultActiveKey="1" items={items} />
    </Card>
  );
};
