import React, { useState } from 'react';
import { useTranslation } from "react-i18next";
import { Button, Typography, Space } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { AppointmentCalendar } from '@/components/appointment/AppointmentCalendar';
import { BookingModal } from '@/components/appointment/BookingModal';

const { Title } = Typography;

export const AppointmentsPage: React.FC = () => {
  const { t } = useTranslation();
  const [isBookingOpen, setIsBookingOpen] = useState(false);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>{t("appointments.title")}</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsBookingOpen(true)}>
          {t("appointments.bookNew")}
        </Button>
      </div>

      <div style={{ background: '#fff', padding: 24, borderRadius: 8 }}>
        <AppointmentCalendar />
      </div>

      <BookingModal
        open={isBookingOpen}
        onClose={() => setIsBookingOpen(false)}
        onSuccess={() => setIsBookingOpen(false)}
      />
    </div>
  );
};
