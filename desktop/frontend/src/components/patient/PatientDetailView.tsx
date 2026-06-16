import React, { useEffect } from 'react';
import { Card, Descriptions, Tabs, Spin, Alert, Typography } from 'antd';
import { usePatientStore } from '@/store/patientStore';

const { Title } = Typography;

interface PatientDetailViewProps {
  patientId: string;
}

export const PatientDetailView: React.FC<PatientDetailViewProps> = ({ patientId }) => {
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
    return <Alert message="Không tìm thấy thông tin bệnh nhân." type="warning" />;
  }

  const items = [
    {
      key: '1',
      label: 'Thông tin chung',
      children: (
        <Descriptions bordered column={{ xxl: 2, xl: 2, lg: 2, md: 1, sm: 1, xs: 1 }}>
          <Descriptions.Item label="Mã BN">{selectedPatient.patient_code}</Descriptions.Item>
          <Descriptions.Item label="Họ và tên">{selectedPatient.full_name}</Descriptions.Item>
          <Descriptions.Item label="Ngày sinh">{selectedPatient.dob}</Descriptions.Item>
          <Descriptions.Item label="Giới tính">
            {selectedPatient.gender === 'MALE' ? 'Nam' : selectedPatient.gender === 'FEMALE' ? 'Nữ' : 'Khác'}
          </Descriptions.Item>
          {/* Note: This component is for staff, so we show the full phone/cccd instead of masked */}
          <Descriptions.Item label="SĐT">{selectedPatient.phone || selectedPatient.phone_masked}</Descriptions.Item>
          <Descriptions.Item label="CCCD">{selectedPatient.cccd || 'N/A'}</Descriptions.Item>
          <Descriptions.Item label="Email">{selectedPatient.email || 'N/A'}</Descriptions.Item>
          <Descriptions.Item label="Địa chỉ">{selectedPatient.address || 'N/A'}</Descriptions.Item>
        </Descriptions>
      ),
    },
    {
      key: '2',
      label: 'BHYT',
      children: <Alert message="Chưa có dữ liệu BHYT." type="info" />,
    },
    {
      key: '3',
      label: 'Lịch sử khám',
      children: <Alert message="Chưa có lịch sử khám bệnh." type="info" />,
    },
  ];

  return (
    <Card title={`Hồ sơ bệnh nhân: ${selectedPatient.full_name}`}>
      <Tabs defaultActiveKey="1" items={items} />
    </Card>
  );
};
