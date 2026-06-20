import React, { useState, useEffect } from 'react';
import { Modal, Input, List, Typography, Space, Button } from 'antd';
import { SearchOutlined, UserOutlined } from '@ant-design/icons';
import { useTranslation } from "react-i18next";
import { usePatientStore, Patient } from '@/store/patientStore';

const { Text } = Typography;

interface PatientSearchModalProps {
  open: boolean;
  onClose: () => void;
  onSelect: (patient: Patient) => void;
}

export const PatientSearchModal: React.FC<PatientSearchModalProps> = ({ open, onClose, onSelect }) => {
  const { t } = useTranslation();
  const [searchTerm, setSearchTerm] = useState('');
  const { searchPatients, searchResults, isLoading } = usePatientStore();

  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchTerm.length >= 2) {
        searchPatients(searchTerm);
      } else {
        searchPatients('');
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
      title={t("patients.searchTitle")}
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnClose
    >
      <Input
        placeholder={t("patients.searchPlaceholder")}
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
                  {t("common.select")}
                </Button>
              ]}
            >
              <List.Item.Meta
                avatar={<UserOutlined style={{ fontSize: 24, color: '#1890ff' }} />}
                title={<Text strong>{patient.full_name}</Text>}
                description={
                  <Space direction="vertical" size={0}>
                    <Text type="secondary">{t("patients.codePrefix")} {patient.patient_code || 'N/A'} | {t("patients.dobPrefix")} {patient.dob}</Text>
                    <Text type="secondary">{t("patients.phonePrefix")} {patient.phone_masked}</Text>
                  </Space>
                }
              />
            </List.Item>
          )}
          locale={{ emptyText: searchTerm.length >= 2 ? t("patients.notFound") : t("patients.searchMinChars") }}
        />
      </div>
    </Modal>
  );
};
