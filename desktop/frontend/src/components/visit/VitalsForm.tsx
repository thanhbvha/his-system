import React, { useState } from "react";
import { Form, InputNumber, Button, Card, Row, Col, Alert, Space } from "antd";
import { useTranslation } from "react-i18next";
import { useVisitStore } from "@/store/visitStore";

interface VitalsFormProps {
  visitId: string;
}

export const VitalsForm: React.FC<VitalsFormProps> = ({ visitId }) => {
  const { t } = useTranslation();
  const { recordVitals } = useVisitStore();
  const [form] = Form.useForm();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [warnings, setWarnings] = useState<string[]>([]);

  const checkWarnings = (values: any) => {
    const newWarnings: string[] = [];
    if (values.bp_systolic > 140 || values.bp_systolic < 90) newWarnings.push("Huyết áp tâm thu bất thường");
    if (values.bp_diastolic > 90 || values.bp_diastolic < 60) newWarnings.push("Huyết áp tâm trương bất thường");
    if (values.heart_rate > 100 || values.heart_rate < 60) newWarnings.push("Nhịp tim bất thường");
    if (values.temperature > 37.5) newWarnings.push("Sốt (Nhiệt độ cao)");
    if (values.spo2 < 95) newWarnings.push("SpO2 thấp");
    setWarnings(newWarnings);
  };

  const onValuesChange = (_: any, allValues: any) => {
    checkWarnings(allValues);
  };

  const onFinish = async (values: any) => {
    setIsSubmitting(true);
    try {
      await recordVitals(visitId, values);
      form.resetFields();
      setWarnings([]);
    } catch (error) {
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card title={t("visit.addVitals", "Ghi nhận sinh hiệu")} className="mb-6">
      {warnings.length > 0 && (
        <Alert
          message={t("visit.abnormal", "Cảnh báo chỉ số bất thường")}
          description={
            <ul className="pl-4 m-0">
              {warnings.map((w, i) => <li key={i}>{w}</li>)}
            </ul>
          }
          type="warning"
          showIcon
          className="mb-4"
        />
      )}
      
      <Form 
        form={form} 
        layout="vertical" 
        onFinish={onFinish}
        onValuesChange={onValuesChange}
      >
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item label={t("visit.bp", "Huyết áp (mmHg)")} className="mb-0">
              <Space.Compact>
                <Form.Item name="bp_systolic" style={{ width: '45%', textAlign: 'center' }} rules={[{ required: true }]}>
                  <InputNumber placeholder="Tâm thu" style={{ width: '100%' }} />
                </Form.Item>
                <div style={{ width: '10%', display: 'inline-block', textAlign: 'center', lineHeight: '32px' }}>/</div>
                <Form.Item name="bp_diastolic" style={{ width: '45%', textAlign: 'center' }} rules={[{ required: true }]}>
                  <InputNumber placeholder="Tâm trương" style={{ width: '100%' }} />
                </Form.Item>
              </Space.Compact>
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="heart_rate" label={t("visit.heartRate", "Nhịp tim (bpm)")} rules={[{ required: true }]}>
              <InputNumber className="w-full" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="temperature" label={t("visit.temp", "Nhiệt độ (°C)")} rules={[{ required: true }]}>
              <InputNumber className="w-full" step={0.1} />
            </Form.Item>
          </Col>
        </Row>
        
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item name="spo2" label={t("visit.spo2", "SpO2 (%)")}>
              <InputNumber className="w-full" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="weight_kg" label={t("visit.weight", "Cân nặng (kg)")}>
              <InputNumber className="w-full" step={0.1} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="height_cm" label={t("visit.height", "Chiều cao (cm)")}>
              <InputNumber className="w-full" />
            </Form.Item>
          </Col>
        </Row>
        
        <Button type="primary" htmlType="submit" loading={isSubmitting}>
          {t("common.save", "Lưu")}
        </Button>
      </Form>
    </Card>
  );
};
