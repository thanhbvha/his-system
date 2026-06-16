import React, { useState, useEffect } from 'react';
import { Modal, Form, Select, DatePicker, Button, message, Spin, Row, Col, Typography, Input } from 'antd';
import { useAppointmentStore, AppointmentSlot } from '@/store/appointmentStore';
import { usePublicStore } from '@/store/publicStore';
import { PatientSearchModal } from '../patient/PatientSearchModal';
import { Patient } from '@/store/patientStore';
import dayjs from 'dayjs';

const { Option } = Select;
const { Text } = Typography;

interface BookingModalProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const BookingModal: React.FC<BookingModalProps> = ({ open, onClose, onSuccess }) => {
  const [form] = Form.useForm();
  const { fetchSlots, availableSlots, bookAppointment, isLoading: isAppointmentLoading } = useAppointmentStore();
  const { fetchDoctors, fetchServices, doctors, services, isLoading: isPublicLoading } = usePublicStore();
  
  const [selectedDoctor, setSelectedDoctor] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [selectedPatient, setSelectedPatient] = useState<Patient | null>(null);
  const [isPatientSearchOpen, setIsPatientSearchOpen] = useState(false);

  useEffect(() => {
    // Fetch initial list of services and doctors
    fetchServices();
    fetchDoctors();
  }, [fetchServices, fetchDoctors]);

  useEffect(() => {
    if (selectedDoctor && selectedDate) {
      fetchSlots(selectedDoctor, selectedDate);
    }
  }, [selectedDoctor, selectedDate, fetchSlots]);

  const handleServiceChange = (serviceId: string) => {
    // Optionally refetch doctors based on the selected service
    // fetchDoctors(serviceId); 
    form.setFieldsValue({ doctor_id: undefined });
    setSelectedDoctor(null);
  };

  const onFinish = async (values: any) => {
    if (!selectedPatient) {
      message.error("Vui lòng chọn bệnh nhân!");
      return;
    }
    
    try {
      await bookAppointment({
        patient_id: selectedPatient.id,
        doctor_id: values.doctor_id,
        service_id: values.service_id,
        slot_id: values.slot_id,
        note: values.note
      });
      message.success("Đặt lịch thành công!");
      form.resetFields();
      setSelectedPatient(null);
      onSuccess();
    } catch (error: any) {
      if (error.response?.status === 409) {
        message.error("Slot vừa được đặt, vui lòng chọn giờ khác");
        // Refetch slots
        if (selectedDoctor && selectedDate) {
          fetchSlots(selectedDoctor, selectedDate);
        }
      } else {
        message.error("Đặt lịch thất bại: " + (error.response?.data?.message || error.message));
      }
    }
  };

  const handleDateChange = (date: any, dateString: string | string[] | null) => {
      if (!dateString) {
        setSelectedDate(null);
      } else {
        setSelectedDate(Array.isArray(dateString) ? dateString[0] : dateString);
      }
      // reset slot
      form.setFieldsValue({ slot_id: undefined });
  };

  return (
    <>
      <Modal
        title="Đặt lịch khám"
        open={open}
        onCancel={onClose}
        footer={null}
        width={600}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={onFinish}>
          
          <Form.Item label="Bệnh nhân" required>
            <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
              {selectedPatient ? (
                <Text strong>{selectedPatient.full_name} ({selectedPatient.phone_masked || selectedPatient.phone})</Text>
              ) : (
                <Text type="secondary">Chưa chọn bệnh nhân</Text>
              )}
              <Button onClick={() => setIsPatientSearchOpen(true)}>
                {selectedPatient ? 'Chọn lại' : 'Tìm bệnh nhân'}
              </Button>
            </div>
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="service_id" label="Dịch vụ" rules={[{ required: true }]}>
                <Select placeholder="Chọn dịch vụ" onChange={handleServiceChange} loading={isPublicLoading}>
                  {services.map(s => <Option key={s.id} value={s.id}>{s.name}</Option>)}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="doctor_id" label="Bác sĩ" rules={[{ required: true }]}>
                <Select placeholder="Chọn bác sĩ" onChange={setSelectedDoctor} loading={isPublicLoading} disabled={!form.getFieldValue('service_id')}>
                  {doctors.map(d => <Option key={d.id} value={d.id}>{d.full_name || (d as any).name}</Option>)}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="date" label="Ngày khám" rules={[{ required: true }]}>
                <DatePicker 
                  format="YYYY-MM-DD" 
                  style={{ width: '100%' }} 
                  onChange={handleDateChange}
                  disabledDate={current => current && current < dayjs().startOf('day')}
                />
              </Form.Item>
            </Col>
          </Row>

          {selectedDoctor && selectedDate && (
            <Form.Item name="slot_id" label="Khung giờ" rules={[{ required: true, message: 'Vui lòng chọn khung giờ' }]}>
              <Spin spinning={isAppointmentLoading}>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                  {availableSlots.length > 0 ? availableSlots.map(slot => (
                    <Button 
                      key={slot.id} 
                      type={form.getFieldValue('slot_id') === slot.id ? 'primary' : 'default'}
                      disabled={slot.is_booked}
                      onClick={() => form.setFieldsValue({ slot_id: slot.id })}
                    >
                      {slot.start_time}
                    </Button>
                  )) : <Text type="secondary">Không có lịch trống trong ngày này</Text>}
                </div>
              </Spin>
            </Form.Item>
          )}

          <Form.Item name="note" label="Ghi chú">
             <Input.TextArea rows={2} />
          </Form.Item>

          <Form.Item style={{ textAlign: 'right', marginTop: '20px' }}>
            <Button onClick={onClose} style={{ marginRight: '10px' }}>Hủy</Button>
            <Button type="primary" htmlType="submit" loading={isAppointmentLoading}>Đặt lịch</Button>
          </Form.Item>
        </Form>
      </Modal>

      <PatientSearchModal 
        open={isPatientSearchOpen} 
        onClose={() => setIsPatientSearchOpen(false)} 
        onSelect={setSelectedPatient} 
      />
    </>
  );
};
