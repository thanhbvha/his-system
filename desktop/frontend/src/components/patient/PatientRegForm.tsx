import React, { useState } from 'react';
import { Form, Input, DatePicker, Select, Button, message, Space, Card } from 'antd';
import { usePatientStore } from '@/store/patientStore';
import dayjs from 'dayjs';
import { useTranslation } from "react-i18next";

const { Option } = Select;

interface PatientRegFormProps {
  onSuccess?: (patient: any) => void;
  onCancel?: () => void;
}

export const PatientRegForm: React.FC<PatientRegFormProps> = ({ onSuccess, onCancel }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const { createPatient, isLoading } = usePatientStore();

  const onFinish = async (values: any) => {
    try {
      const payload = {
        ...values,
        dob: values.dob.format('YYYY-MM-DD'),
      };
      
      const newPatient = await createPatient(payload);
      message.success(t("patients.regSuccess") + (newPatient.patient_code || newPatient.id));
      form.resetFields();
      if (onSuccess) onSuccess(newPatient);
    } catch (error: any) {
      message.error(error.response?.data?.message || t("patients.regFail"));
    }
  };

  return (
    <Card title={t("patients.regNewTitle")} bordered={false}>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{ gender: 'MALE' }}
      >
        <Form.Item
          name="full_name"
          label={t("patients.fullName")}
          rules={[{ required: true, message: t("patients.requireFullName") }]}
        >
          <Input placeholder="Nguyễn Văn A" />
        </Form.Item>

        <Space style={{ display: 'flex' }} size="large">
          <Form.Item
            name="dob"
            label={t("patients.dob")}
            rules={[{ required: true, message: t("patients.requireDob") }]}
          >
            <DatePicker format="DD/MM/YYYY" style={{ width: 200 }} />
          </Form.Item>

          <Form.Item
            name="gender"
            label={t("patients.gender")}
            rules={[{ required: true, message: t("patients.requireGender") }]}
          >
            <Select style={{ width: 120 }}>
              <Option value="MALE">{t("patients.male")}</Option>
              <Option value="FEMALE">{t("patients.female")}</Option>
              <Option value="OTHER">{t("patients.other")}</Option>
            </Select>
          </Form.Item>
        </Space>

        <Space style={{ display: 'flex' }} size="large">
          <Form.Item
            name="phone"
            label={t("patients.phone")}
            rules={[
              { required: true, message: t("patients.requirePhone") },
              { pattern: /^[0-9]{10}$/, message: t("patients.invalidPhone") }
            ]}
          >
            <Input placeholder="09xxxxxxxx" maxLength={10} style={{ width: 200 }} />
          </Form.Item>

          <Form.Item
            name="cccd"
            label={t("patients.cccd")}
            rules={[
              { pattern: /^[0-9]{12}$/, message: t("patients.invalidCccd") }
            ]}
          >
            <Input placeholder="001xxxxxxxx" maxLength={12} style={{ width: 200 }} />
          </Form.Item>
        </Space>

        <Form.Item name="email" label={t("patients.email")} rules={[{ type: 'email', message: t("patients.invalidEmail") }]}>
          <Input placeholder="email@example.com" />
        </Form.Item>

        <Form.Item name="address" label={t("patients.address")}>
          <Input.TextArea rows={2} placeholder="Số nhà, đường, phường/xã, quận/huyện..." />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isLoading}>
              {t("patients.register")}
            </Button>
            {onCancel && (
              <Button onClick={onCancel} disabled={isLoading}>
                {t("common.cancel")}
              </Button>
            )}
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
};
