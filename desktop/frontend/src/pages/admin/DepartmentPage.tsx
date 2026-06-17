import { useState } from "react";
import { Table, Button, Modal, Form, Input, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

export const DepartmentPage = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  const { data: deptsData, isLoading } = useQuery({
    queryKey: ["departments"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/departments");
      return res.data.data;
    }
  });

  const createMutation = useMutation({
    mutationFn: async (values: any) => {
      await apiClient.post("/admin/departments", values);
    },
    onSuccess: () => {
      message.success(t("admin.departments.createSuccess"));
      setIsModalVisible(false);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ["departments"] });
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || t("common.error"));
    }
  });

  const columns = [
    { title: t("admin.departments.code"), dataIndex: "code", key: "code", width: 150 },
    { title: t("admin.departments.name"), dataIndex: "name", key: "name", width: 250 },
    { title: t("admin.departments.description"), dataIndex: "description", key: "description" },
    { title: t("admin.departments.employeeCount"), dataIndex: "employee_count", key: "employee_count", width: 150, render: (val: number) => val || 0 },
  ];

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <h2>{t("admin.departments.title")}</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalVisible(true)}>
          {t("admin.departments.addNew")}
        </Button>
      </div>

      <Table 
        columns={columns} 
        dataSource={deptsData || []} 
        rowKey="id" 
        loading={isLoading} 
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title={t("admin.departments.addTitle")}
        open={isModalVisible}
        onOk={() => form.submit()}
        onCancel={() => setIsModalVisible(false)}
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate(values)}>
          <Form.Item 
            name="code" 
            label={t("admin.departments.code")}
            rules={[{ required: true, message: t("admin.departments.requireCode") }]}
          >
            <Input placeholder={t("admin.departments.codePlaceholder")} />
          </Form.Item>
          <Form.Item 
            name="name" 
            label={t("admin.departments.name")}
            rules={[{ required: true, message: t("admin.departments.requireName") }]}
          >
            <Input placeholder={t("admin.departments.namePlaceholder")} />
          </Form.Item>
          <Form.Item 
            name="description" 
            label={t("admin.departments.description")}
          >
            <Input.TextArea rows={3} placeholder={t("admin.departments.descPlaceholder")} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};
