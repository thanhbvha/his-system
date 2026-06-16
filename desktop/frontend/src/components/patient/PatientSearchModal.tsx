import React, { useState, useEffect } from 'react';
import { Modal, Input, List, Typography, Space, Button } from 'antd';
import { SearchOutlined, UserOutlined } from '@ant-design/icons';
import { usePatientStore, Patient } from '@/store/patientStore';

const { Text } = Typography;

interface PatientSearchModalProps {
  open: boolean;
  onClose: () => void;
  onSelect: (patient: Patient) => void;
}

export const PatientSearchModal: React.FC<PatientSearchModalProps> = ({ open, onClose, onSelect }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const { searchPatients, searchResults, isLoading } = usePatientStore();

  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchTerm.length >= 2) {
        searchPatients(searchTerm);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm, searchPatients]);

  const handleSelect = (patient: Patient) => {
    onSelect(patient);
    onClose();
  };

  return (
    <Modal
      title="Tìm kiếm Bệnh nhân"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <Input
        placeholder="Nhập tên, số điện thoại (10 số) hoặc CCCD"
        prefix={<SearchOutlined />}
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        size="large"
        allowClear
      />

      <div style={{ marginTop: 20, maxHeight: 400, overflowY: 'auto' }}>
        <List
          loading={isLoading}
          dataSource={searchResults}
          renderItem={(patient) => (
            <List.Item
              actions={[
                <Button key="select" type="primary" size="small" onClick={() => handleSelect(patient)}>
                  Chọn
                </Button>
              ]}
            >
              <List.Item.Meta
                avatar={<UserOutlined style={{ fontSize: 24, color: '#1890ff' }} />}
                title={<Text strong>{patient.full_name}</Text>}
                description={
                  <Space direction="vertical" size={0}>
                    <Text type="secondary">Mã: {patient.patient_code || 'N/A'} | NS: {patient.dob}</Text>
                    <Text type="secondary">SĐT: {patient.phone_masked}</Text>
                  </Space>
                }
              />
            </List.Item>
          )}
          locale={{ emptyText: searchTerm.length >= 2 ? 'Không tìm thấy bệnh nhân' : 'Nhập ít nhất 2 ký tự để tìm kiếm' }}
        />
      </div>
    </Modal>
  );
};
