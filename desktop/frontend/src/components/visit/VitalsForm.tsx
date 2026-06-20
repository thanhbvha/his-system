import React from 'react';
import { Form, InputNumber, Button, Space, Card, Row, Col, Typography, message } from 'antd';
import { useTranslation } from "react-i18next";
import { useVisitStore } from '@/store/visitStore';

const { Text } = Typography;

interface VitalsFormProps {
  visitId: string;
}

export const VitalsForm: React.FC<VitalsFormProps> = ({ visitId }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const { recordVitals, isLoading } = useVisitStore();

  const onFinish = async (values: any) => {
    try {
      await recordVitals(visitId, values);
      message.success(t("visit.addVitals", "Ghi nhận sinh hiệu thành công"));
      form.resetFields();
    } catch (err) {
      message.error(t("common.error", "Đã xảy ra lỗi"));
    }
  };

  const checkAbnormal = (field: string, value: number | undefined) => {
    if (value === undefined || value === null) return false;
    switch (field) {
      case 'bp_systolic': return value > 140 || value < 90;
      case 'bp_diastolic': return value > 90 || value < 60;
      case 'heart_rate': return value > 100 || value < 60;
      case 'temperature': return value > 37.5;
      case 'spo2': return value < 95;
      default: return false;
    }
  };

  return (
    <Card title={t("visit.addVitals", "Ghi nhận sinh hiệu")} bordered={false}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item name="bp_systolic" label={`${t("visit.bp", "Huyết áp")} Tâm thu (mmHg)`}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="bp_diastolic" label={`${t("visit.bp", "Huyết áp")} Tâm trương (mmHg)`}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="heart_rate" label={`${t("visit.heartRate", "Nhịp tim")} (bpm)`}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="temperature" label={`${t("visit.temp", "Nhiệt độ")} (°C)`}>
              <InputNumber style={{ width: '100%' }} step={0.1} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="spo2" label={`${t("visit.spo2", "SpO2")} (%)`}>
              <InputNumber style={{ width: '100%' }} max={100} />
            </Form.Item>
          </Col>
          <Col span={4}>
            <Form.Item name="weight_kg" label={`${t("visit.weight", "Cân nặng")} (kg)`}>
              <InputNumber style={{ width: '100%' }} step={0.1} />
            </Form.Item>
          </Col>
          <Col span={4}>
            <Form.Item name="height_cm" label={`${t("visit.height", "Chiều cao")} (cm)`}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={isLoading}>
            {t("common.confirm", "Xác nhận")}
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};
