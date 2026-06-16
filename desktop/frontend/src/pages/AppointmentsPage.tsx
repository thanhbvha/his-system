import React, { useState } from 'react';
import { Button, Typography, Space } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { AppointmentCalendar } from '@/components/appointment/AppointmentCalendar';
import { BookingModal } from '@/components/appointment/BookingModal';

const { Title } = Typography;

export const AppointmentsPage: React.FC = () => {
  const [isBookingOpen, setIsBookingOpen] = useState(false);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>Lịch hẹn</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsBookingOpen(true)}>
          Đặt lịch khám
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
