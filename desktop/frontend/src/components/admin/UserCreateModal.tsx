import { useState } from "react";
import { Modal, Form, Input, Button, Select, message, Space } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CopyOutlined, SyncOutlined } from "@ant-design/icons";
import apiClient from "@/lib/apiClient";

interface UserCreateModalProps {
  open: boolean;
  onClose: () => void;
  roles: any[];
  departments: any[];
}

export const UserCreateModal = ({ open, onClose, roles, departments }: UserCreateModalProps) => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();
  const [generatedPassword, setGeneratedPassword] = useState("");

  const generatePassword = () => {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*";
    let pass = "";
    for (let i = 0; i < 10; i++) {
      pass += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setGeneratedPassword(pass);
    form.setFieldsValue({ password: pass });
  };

  const copyPassword = () => {
    if (generatedPassword) {
      navigator.clipboard.writeText(generatedPassword);
      message.success("Đã copy mật khẩu");
    }
  };

  const createMutation = useMutation({
    mutationFn: async (values: any) => {
      await apiClient.post("/admin/users", values);
    },
    onSuccess: () => {
      message.success("Tạo tài khoản thành công!");
      form.resetFields();
      setGeneratedPassword("");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      onClose();
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || "Lỗi khi tạo tài khoản");
    }
  });

  return (
    <Modal
      title="Tạo nhân viên mới"
      open={open}
      onCancel={onClose}
      confirmLoading={createMutation.isPending}
      onOk={() => form.submit()}
      width={600}
    >
      <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate(values)}>
        <Form.Item name="username" label="Tên đăng nhập" rules={[{ required: true }]}>
          <Input placeholder="VD: nguyenvan_a" />
        </Form.Item>

        <Form.Item name="email" label="Email" rules={[{ required: true, type: "email" }]}>
          <Input placeholder="VD: a.nguyen@hospital.com" />
        </Form.Item>

        <Form.Item label="Mật khẩu khởi tạo" required>
          <Space.Compact style={{ width: "100%" }}>
            <Form.Item name="password" noStyle rules={[{ required: true, message: "Vui lòng tạo mật khẩu" }]}>
              <Input.Password value={generatedPassword} placeholder="Nhấn nút Generate để tạo" />
            </Form.Item>
            <Button icon={<SyncOutlined />} onClick={generatePassword}>Generate</Button>
            <Button icon={<CopyOutlined />} onClick={copyPassword} disabled={!generatedPassword}>Copy</Button>
          </Space.Compact>
        </Form.Item>

        <Form.Item name="role_ids" label="Quyền (Roles)" rules={[{ required: true }]}>
          <Select mode="multiple" placeholder="Chọn một hoặc nhiều Role">
            {roles?.map(r => <Select.Option key={r.id} value={r.id}>{r.name}</Select.Option>)}
          </Select>
        </Form.Item>

        <Form.Item name="department_id" label="Khoa / Phòng ban" rules={[{ required: true }]}>
          <Select placeholder="Chọn phòng ban">
            {departments?.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
};
