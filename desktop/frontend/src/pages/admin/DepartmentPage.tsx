import { useState } from "react";
import { Table, Button, Modal, Form, Input, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/apiClient";

export const DepartmentPage = () => {
  const queryClient = useQueryClient();
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
      message.success("Đã thêm Khoa/Phòng ban mới");
      setIsModalVisible(false);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ["departments"] });
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || "Có lỗi xảy ra");
    }
  });

  const columns = [
    { title: "Mã khoa", dataIndex: "code", key: "code", width: 150 },
    { title: "Tên khoa/phòng", dataIndex: "name", key: "name" },
    { title: "Số nhân viên", dataIndex: "employee_count", key: "employee_count", width: 150, render: (val: number) => val || 0 },
  ];

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <h2>Quản lý Khoa / Phòng ban</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalVisible(true)}>
          Thêm mới
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
        title="Thêm Khoa / Phòng ban mới"
        open={isModalVisible}
        onOk={() => form.submit()}
        onCancel={() => setIsModalVisible(false)}
        confirmLoading={createMutation.isPending}
      >
        <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate(values)}>
          <Form.Item 
            name="code" 
            label="Mã khoa" 
            rules={[{ required: true, message: "Vui lòng nhập mã khoa" }]}
          >
            <Input placeholder="VD: KHOA_NOI" />
          </Form.Item>
          <Form.Item 
            name="name" 
            label="Tên khoa/phòng" 
            rules={[{ required: true, message: "Vui lòng nhập tên khoa" }]}
          >
            <Input placeholder="VD: Khoa Nội tổng hợp" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};
