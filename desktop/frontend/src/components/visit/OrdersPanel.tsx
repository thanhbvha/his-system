import React, { useState } from 'react';
import { Form, Input, Select, Button, Table, Space, Card, Tag, Typography, message } from 'antd';
import { useTranslation } from "react-i18next";
import { useVisitStore, VisitOrder } from '@/store/visitStore';
import { ICD10Search } from '../icd10/ICD10Search';

const { Option } = Select;
const { Title } = Typography;

interface OrdersPanelProps {
  visitId: string;
  orders: VisitOrder[];
}

export const OrdersPanel: React.FC<OrdersPanelProps> = ({ visitId, orders }) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const { createOrder, isLoading } = useVisitStore();
  const [selectedIcd10, setSelectedIcd10] = useState<string | null>(null);

  const handleIcd10Select = (code: string, desc: string) => {
    setSelectedIcd10(code);
    const currentDetails = form.getFieldValue('details') || '';
    form.setFieldsValue({
      details: currentDetails ? `${currentDetails}\n[${code}] ${desc}` : `[${code}] ${desc}`
    });
  };

  const onFinish = async (values: any) => {
    try {
      await createOrder(visitId, {
        order_type: values.order_type,
        details: values.details
      });
      message.success(t("common.success", "Tạo chỉ định thành công"));
      form.resetFields();
    } catch (err) {
      message.error(t("common.error", "Tạo chỉ định thất bại"));
    }
  };

  const columns = [
    {
      title: 'Loại',
      dataIndex: 'order_type',
      key: 'order_type',
      render: (type: string) => <Tag color="blue">{type}</Tag>
    },
    {
      title: 'Chi tiết',
      dataIndex: 'details',
      key: 'details',
    },
    {
      title: 'Trạng thái',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={status === 'COMPLETED' ? 'green' : 'orange'}>{status}</Tag>
    }
  ];

  return (
    <Card bordered={false}>
      <div style={{ marginBottom: 24 }}>
        <Title level={5}>Chẩn đoán (ICD-10)</Title>
        <ICD10Search onSelect={handleIcd10Select} />
      </div>

      <div style={{ marginBottom: 24 }}>
        <Title level={5}>{t("visit.addOrder", "Thêm chỉ định")}</Title>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Space align="start" size="large">
            <Form.Item 
              name="order_type" 
              label="Loại chỉ định"
              rules={[{ required: true }]}
            >
              <Select style={{ width: 200 }}>
                <Option value="LAB">Xét nghiệm (Lab)</Option>
                <Option value="RADIOLOGY">Chẩn đoán hình ảnh</Option>
                <Option value="PROCEDURE">Thủ thuật</Option>
              </Select>
            </Form.Item>
            <Form.Item 
              name="details" 
              label="Chi tiết / Yêu cầu"
              rules={[{ required: true }]}
              style={{ width: 400 }}
            >
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item label=" ">
              <Button type="primary" htmlType="submit" loading={isLoading}>
                {t("common.confirm", "Xác nhận")}
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </div>

      <Title level={5}>Danh sách chỉ định</Title>
      <Table 
        columns={columns} 
        dataSource={orders} 
        rowKey="id" 
        pagination={false}
        locale={{ emptyText: t("visit.noOrders", "Chưa có chỉ định nào") }}
      />
    </Card>
  );
};
