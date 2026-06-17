import { useState } from "react";
import { Modal, Form, Input, Button, Select, message, Space } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CopyOutlined, SyncOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

interface UserCreateModalProps {
  open: boolean;
  onClose: () => void;
  roles: any[];
  departments: any[];
}

export const UserCreateModal = ({ open, onClose, roles, departments }: UserCreateModalProps) => {
  const { t } = useTranslation();
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
      message.success(t("admin.userCreate.copySuccess"));
    }
  };

  const createMutation = useMutation({
    mutationFn: async (values: any) => {
      await apiClient.post("/admin/users", values);
    },
    onSuccess: () => {
      message.success(t("admin.userCreate.createSuccess"));
      form.resetFields();
      setGeneratedPassword("");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      onClose();
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || t("admin.userCreate.createFail"));
    }
  });

  return (
    <Modal
      title={t("admin.userCreate.title")}
      open={open}
      onCancel={onClose}
      confirmLoading={createMutation.isPending}
      onOk={() => form.submit()}
      width={600}
    >
      <Form form={form} layout="vertical" onFinish={(values) => createMutation.mutate(values)}>
        <Form.Item name="username" label={t("admin.userCreate.username")} rules={[{ required: true }]}>
          <Input placeholder={t("admin.userCreate.usernamePlaceholder")} />
        </Form.Item>

        <Form.Item name="email" label={t("admin.userCreate.email")} rules={[{ required: true, type: "email" }]}>
          <Input placeholder={t("admin.userCreate.emailPlaceholder")} />
        </Form.Item>

        <Form.Item label={t("admin.userCreate.passwordInit")} required>
          <Space.Compact style={{ width: "100%" }}>
            <Form.Item name="password" noStyle rules={[{ required: true, message: t("admin.userCreate.requirePassword") }]}>
              <Input.Password value={generatedPassword} placeholder={t("admin.userCreate.passwordPlaceholder")} />
            </Form.Item>
            <Button icon={<SyncOutlined />} onClick={generatePassword}>Generate</Button>
            <Button icon={<CopyOutlined />} onClick={copyPassword} disabled={!generatedPassword}>Copy</Button>
          </Space.Compact>
        </Form.Item>

        <Form.Item name="role_ids" label={t("admin.userCreate.roles")} rules={[{ required: true }]}>
          <Select mode="multiple" placeholder={t("admin.roles.selectMultiplePlaceholder")}>
            {roles?.map(r => <Select.Option key={r.id} value={r.id}>{r.name}</Select.Option>)}
          </Select>
        </Form.Item>

        <Form.Item name="department_id" label={t("admin.userCreate.department")} rules={[{ required: true }]}>
          <Select placeholder={t("admin.userCreate.deptPlaceholder")}>
            {departments?.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
};
