import React, { useState } from 'react';
import { Button, Space, Typography, Modal } from 'antd';
import { SearchOutlined, UserAddOutlined } from '@ant-design/icons';
import { PatientSearchModal } from '@/components/patient/PatientSearchModal';
import { PatientRegForm } from '@/components/patient/PatientRegForm';
import { PatientDetailView } from '@/components/patient/PatientDetailView';
import { Patient } from '@/store/patientStore';

const { Title } = Typography;

export const PatientsPage: React.FC = () => {
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
        <Title level={2} style={{ margin: 0 }}>Quản lý Bệnh nhân</Title>
        <Space>
          <Button type="primary" icon={<UserAddOutlined />} onClick={() => setIsRegOpen(true)}>
            Đăng ký mới
          </Button>
          <Button icon={<SearchOutlined />} onClick={() => setIsSearchOpen(true)}>
            Tìm kiếm
          </Button>
        </Space>
      </div>

      {selectedPatientId ? (
        <PatientDetailView patientId={selectedPatientId} />
      ) : (
        <div style={{ textAlign: 'center', marginTop: 100, color: '#999' }}>
          <p>Vui lòng tìm kiếm hoặc đăng ký bệnh nhân mới để xem hồ sơ.</p>
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
