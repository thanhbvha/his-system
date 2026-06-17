import React, { useState } from 'react';
import { useTranslation } from "react-i18next";
import { Button, Space, Typography, Modal } from 'antd';
import { SearchOutlined, UserAddOutlined } from '@ant-design/icons';
import { PatientSearchModal } from '@/components/patient/PatientSearchModal';
import { PatientRegForm } from '@/components/patient/PatientRegForm';
import { PatientDetailView } from '@/components/patient/PatientDetailView';
import { Patient } from '@/store/patientStore';

const { Title } = Typography;

export const PatientsPage: React.FC = () => {
  const { t } = useTranslation();
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isRegOpen, setIsRegOpen] = useState(false);
  const [selectedPatientId, setSelectedPatientId] = useState<string | null>(null);

  const handlePatientSelect = (patient: Patient) => {
    setSelectedPatientId(patient.id);
  };

  const handleRegSuccess = (patient: Patient) => {
    setIsRegOpen(false);
    setSelectedPatientId(patient.id);
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t("patients.title")}</Title>
        <Space>
          <Button type="primary" icon={<UserAddOutlined />} onClick={() => setIsRegOpen(true)}>
            {t("patients.registerNew")}
          </Button>
          <Button icon={<SearchOutlined />} onClick={() => setIsSearchOpen(true)}>
            {t("common.search")}
          </Button>
        </Space>
      </div>

      {selectedPatientId ? (
        <PatientDetailView patientId={selectedPatientId} />
      ) : (
        <div style={{ textAlign: 'center', marginTop: 100, color: '#999' }}>
          <p>{t("patients.noPatientSelected")}</p>
        </div>
      )}

      <PatientSearchModal
        open={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
        onSelect={handlePatientSelect}
      />

      <Modal
        open={isRegOpen}
        onCancel={() => setIsRegOpen(false)}
        footer={null}
        width={800}
        destroyOnClose
      >
        <PatientRegForm onSuccess={handleRegSuccess} onCancel={() => setIsRegOpen(false)} />
      </Modal>
    </div>
  );
};
