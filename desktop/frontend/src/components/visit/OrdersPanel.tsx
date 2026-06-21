import React, { useState } from "react";
import { Form, Select, Input, Button, Table, Badge, Typography, Card } from "antd";
import { useTranslation } from "react-i18next";
import { useVisitStore, VisitOrder } from "@/store/visitStore";
import { ICD10Search } from "../icd10/ICD10Search";

const { Text } = Typography;

interface OrdersPanelProps {
  visitId: string;
  orders: VisitOrder[];
}

export const OrdersPanel: React.FC<OrdersPanelProps> = ({ visitId, orders }) => {
  const { t } = useTranslation();
  const { createOrder } = useVisitStore();
  const [form] = Form.useForm();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const onFinish = async (values: any) => {
    setIsSubmitting(true);
    try {
      await createOrder(visitId, values);
      form.resetFields();
    } catch (error) {
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const columns = [
    {
      title: 'Loại chỉ định',
      dataIndex: 'order_type',
      key: 'order_type',
      render: (type: string) => {
        const types: Record<string, string> = {
          'LAB': 'Xét nghiệm',
          'RADIOLOGY': 'Chẩn đoán hình ảnh',
          'PROCEDURE': 'Thủ thuật'
        };
        return <Text strong>{types[type] || type}</Text>;
      }
    },
    {
      title: 'Mô tả chi tiết',
      dataIndex: 'details',
      key: 'details',
    },
    {
      title: 'Trạng thái',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Badge status={status === 'COMPLETED' ? 'success' : 'processing'} text={status} />
    }
  ];

  return (
    <div>
      <Card title="Chẩn đoán ICD-10" className="mb-6">
        <div className="mb-4 text-gray-500">Chức năng kê đơn thuốc và tích hợp hệ thống Dược sẽ có trong Sprint 5.</div>
        <Form layout="vertical">
          <Form.Item label="Tìm kiếm mã bệnh (ICD-10)">
            <ICD10Search onSelect={(val: string) => console.log("Selected ICD-10", val)} />
          </Form.Item>
        </Form>
      </Card>

      <Card title={t("visit.orders", "Chỉ định cận lâm sàng")}>
        <Form 
          form={form} 
          layout="inline" 
          onFinish={onFinish}
          className="mb-6"
        >
          <Form.Item name="order_type" rules={[{ required: true }]} style={{ width: 200 }}>
            <Select placeholder="Chọn loại chỉ định">
              <Select.Option value="LAB">Xét nghiệm</Select.Option>
              <Select.Option value="RADIOLOGY">Chẩn đoán hình ảnh</Select.Option>
              <Select.Option value="PROCEDURE">Thủ thuật</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="details" rules={[{ required: true }]} style={{ width: 300 }}>
            <Input placeholder="Chi tiết chỉ định (VD: Siêu âm bụng tổng quát)" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={isSubmitting}>
              {t("visit.addOrder", "Thêm chỉ định")}
            </Button>
          </Form.Item>
        </Form>

        <Table 
          dataSource={orders} 
          columns={columns} 
          rowKey="id" 
          pagination={false}
          locale={{ emptyText: t("visit.noOrders", "Chưa có chỉ định nào") }}
        />
      </Card>
    </div>
  );
};
