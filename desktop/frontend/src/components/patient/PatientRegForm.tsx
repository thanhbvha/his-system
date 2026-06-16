import React, { useState } from 'react';
import { Form, Input, DatePicker, Select, Button, message, Space, Card } from 'antd';
import { usePatientStore } from '@/store/patientStore';
import dayjs from 'dayjs';

const { Option } = Select;

interface PatientRegFormProps {
  onSuccess?: (patient: any) => void;
  onCancel?: () => void;
}

export const PatientRegForm: React.FC<PatientRegFormProps> = ({ onSuccess, onCancel }) => {
  const [form] = Form.useForm();
  const { createPatient, isLoading } = usePatientStore();

  const onFinish = async (values: any) => {
    try {
      const payload = {
        ...values,
        dob: values.dob.format('YYYY-MM-DD'),
      };
      
      const newPatient = await createPatient(payload);
      message.success(`Đăng ký thành công! Mã BN: ${newPatient.patient_code || newPatient.id}`);
      form.resetFields();
      if (onSuccess) onSuccess(newPatient);
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Đăng ký thất bại');
    }
  };

  return (
    <Card title="Đăng ký Bệnh nhân mới" bordered={false}>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{ gender: 'MALE' }}
      >
        <Form.Item
          name="full_name"
          label="Họ và tên"
          rules={[{ required: true, message: 'Vui lòng nhập họ tên!' }]}
        >
          <Input placeholder="Nguyễn Văn A" />
        </Form.Item>

        <Space style={{ display: 'flex' }} size="large">
          <Form.Item
            name="dob"
            label="Ngày sinh"
            rules={[{ required: true, message: 'Vui lòng chọn ngày sinh!' }]}
          >
            <DatePicker format="DD/MM/YYYY" style={{ width: 200 }} />
          </Form.Item>

          <Form.Item
            name="gender"
            label="Giới tính"
            rules={[{ required: true, message: 'Vui lòng chọn giới tính!' }]}
          >
            <Select style={{ width: 120 }}>
              <Option value="MALE">Nam</Option>
              <Option value="FEMALE">Nữ</Option>
              <Option value="OTHER">Khác</Option>
            </Select>
          </Form.Item>
        </Space>

        <Space style={{ display: 'flex' }} size="large">
          <Form.Item
            name="phone"
            label="Số điện thoại"
            rules={[
              { required: true, message: 'Vui lòng nhập SĐT!' },
              { pattern: /^[0-9]{10}$/, message: 'SĐT phải gồm 10 chữ số!' }
            ]}
          >
            <Input placeholder="09xxxxxxxx" maxLength={10} style={{ width: 200 }} />
          </Form.Item>

          <Form.Item
            name="cccd"
            label="CCCD / CMND"
            rules={[
              { pattern: /^[0-9]{12}$/, message: 'CCCD phải gồm 12 chữ số!' }
            ]}
          >
            <Input placeholder="001xxxxxxxx" maxLength={12} style={{ width: 200 }} />
          </Form.Item>
        </Space>

        <Form.Item name="email" label="Email" rules={[{ type: 'email', message: 'Email không hợp lệ!' }]}>
          <Input placeholder="email@example.com" />
        </Form.Item>

        <Form.Item name="address" label="Địa chỉ liên hệ">
          <Input.TextArea rows={2} placeholder="Số nhà, đường, phường/xã, quận/huyện..." />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isLoading}>
              Đăng ký
            </Button>
            {onCancel && (
              <Button onClick={onCancel} disabled={isLoading}>
                Hủy
              </Button>
            )}
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
};
