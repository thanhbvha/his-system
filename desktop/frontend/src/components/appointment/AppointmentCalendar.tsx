import React, { useEffect, useState } from 'react';
import { Calendar, Badge, Modal, List, Typography, Button, Spin, Tag, Space, DatePicker } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import { useAppointmentStore, Appointment } from '@/store/appointmentStore';

const { Text } = Typography;

export const AppointmentCalendar: React.FC = () => {
  const { fetchByDate, appointments, isLoading, updateStatus } = useAppointmentStore();
  const [selectedDate, setSelectedDate] = useState<Dayjs>(dayjs());
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [selectedAppointments, setSelectedAppointments] = useState<Appointment[]>([]);

  useEffect(() => {
    fetchByDate(selectedDate.format('YYYY-MM-DD'));
  }, [selectedDate, fetchByDate]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'PENDING': return 'gold';
      case 'CONFIRMED': return 'blue';
      case 'CHECKED_IN': return 'green';
      case 'COMPLETED': return 'default';
      case 'CANCELLED': return 'red';
      default: return 'default';
    }
  };

  const getStatusBadgeType = (status: string): "warning" | "processing" | "success" | "default" | "error" => {
    switch (status) {
      case 'PENDING': return 'warning';
      case 'CONFIRMED': return 'processing';
      case 'CHECKED_IN': return 'success';
      case 'COMPLETED': return 'default';
      case 'CANCELLED': return 'error';
      default: return 'default';
    }
  };

  // Prepare data for the calendar cells
  const getListData = (value: Dayjs) => {
    // Ideally, we'd fetch a month's worth of data, but for this step we fetch daily.
    // If we only have daily data, the calendar cells other than selectedDate will be empty.
    // In a real implementation, we'd have a fetchByMonth action.
    if (value.format('YYYY-MM-DD') === selectedDate.format('YYYY-MM-DD')) {
      return appointments;
    }
    return [];
  };

  const onSelect = (date: Dayjs, info: { source: string }) => {
    if (info.source === 'date') {
      setSelectedDate(date);
      // Fetch will be triggered by useEffect
    }
  };

  const handleCellClick = (date: Dayjs) => {
    setSelectedDate(date);
    if (date.format('YYYY-MM-DD') === selectedDate.format('YYYY-MM-DD')) {
       setSelectedAppointments(appointments);
       setIsModalVisible(true);
    }
  };

  const handleStatusUpdate = async (id: string, newStatus: string) => {
    await updateStatus(id, newStatus);
    // Refresh modal data
    setSelectedAppointments(appointments.map(app => app.id === id ? { ...app, status: newStatus as any } : app));
  };

  const cellRender = (current: Dayjs, info: any) => {
    if (info.type === 'date') {
      const listData = getListData(current);
      return (
        <div className="events" onClick={() => handleCellClick(current)} style={{ minHeight: '80px' }}>
          {listData.map((item) => (
            <div key={item.id}>
              <Badge status={getStatusBadgeType(item.status)} text={`${item.scheduled_at} - ${item.patient?.full_name}`} />
            </div>
          ))}
          {listData.length === 0 && current.format('YYYY-MM-DD') === selectedDate.format('YYYY-MM-DD') && isLoading && (
            <Spin size="small" />
          )}
        </div>
      );
    }
    return info.originNode;
  };

  return (
    <div>
      <Calendar value={selectedDate} onSelect={onSelect} cellRender={cellRender} />

      <Modal
        title={`Lịch hẹn ngày ${selectedDate.format('DD/MM/YYYY')}`}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        footer={null}
        width={700}
      >
        <List
          loading={isLoading}
          dataSource={selectedAppointments}
          renderItem={(app) => (
            <List.Item
              actions={[
                app.status === 'PENDING' && (
                  <Button key="confirm" type="primary" size="small" onClick={() => handleStatusUpdate(app.id, 'CONFIRMED')}>
                    Xác nhận
                  </Button>
                ),
                (app.status === 'PENDING' || app.status === 'CONFIRMED') && (
                  <Button key="checkin" type="default" size="small" onClick={() => handleStatusUpdate(app.id, 'CHECKED_IN')}>
                    Check-in
                  </Button>
                ),
                <Button key="cancel" danger size="small" onClick={() => handleStatusUpdate(app.id, 'CANCELLED')} disabled={app.status === 'COMPLETED' || app.status === 'CANCELLED'}>
                  Hủy
                </Button>
              ].filter(Boolean)}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <Text strong>{app.scheduled_at}</Text>
                    <Text>{app.patient?.full_name} ({app.patient?.patient_code})</Text>
                    <Tag color={getStatusColor(app.status)}>{app.status}</Tag>
                  </Space>
                }
                description={`Bác sĩ: ${app.doctor?.full_name} | Dịch vụ: ${app.service?.name} | Ghi chú: ${app.note || 'Không'}`}
              />
            </List.Item>
          )}
          locale={{ emptyText: 'Không có lịch hẹn nào' }}
        />
      </Modal>
    </div>
  );
};
