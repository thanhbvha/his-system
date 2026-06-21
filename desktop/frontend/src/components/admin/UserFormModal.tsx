import { useState, useEffect } from "react";
import { Modal, Form, Input, Button, Select, message, Space } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CopyOutlined, SyncOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

interface UserFormModalProps {
  open: boolean;
  onClose: () => void;
  roles: any[];
  departments: any[];
  editUser?: any; // If provided, we are in Edit mode
}

export const UserFormModal = ({ open, onClose, roles, departments, editUser }: UserFormModalProps) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const queryClient = useQueryClient();
  const [generatedPassword, setGeneratedPassword] = useState("");

  const isEdit = !!editUser;

  useEffect(() => {
    if (open) {
      if (isEdit) {
        form.setFieldsValue({
          username: editUser.username,
          email: editUser.email,
          full_name: editUser.full_name,
          department_id: editUser.department_id,
          role_ids: editUser.roles?.map((r: any) => r.id),
        });
      } else {
        form.resetFields();
        setGeneratedPassword("");
      }
    }
  }, [open, isEdit, editUser, form]);

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

  const saveMutation = useMutation({
    mutationFn: async (values: any) => {
      if (isEdit) {
        await apiClient.put(`/admin/users/${editUser.id}/profile`, {
          full_name: values.full_name,
          department_id: values.department_id,
          role_ids: values.role_ids,
          email: values.email, // Always send, backend will decide whether to process it
        });
      } else {
        await apiClient.post("/admin/users", values);
      }
    },
    onSuccess: () => {
      message.success(isEdit ? t("common.success") : t("admin.userCreate.createSuccess"));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      onClose();
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || (isEdit ? t("common.error") : t("admin.userCreate.createFail")));
    }
  });

  const isEmailDisabled = isEdit && !!editUser?.email;

  return (
    <Modal
      title={isEdit ? t("common.edit") : t("admin.userCreate.title")}
      open={open}
      onCancel={onClose}
      confirmLoading={saveMutation.isPending}
      onOk={() => form.submit()}
      width={600}
    >
      <Form form={form} layout="vertical" onFinish={(values) => saveMutation.mutate(values)}>
        <Form.Item name="username" label={t("admin.userCreate.username")} rules={[{ required: true }]}>
          <Input placeholder={t("admin.userCreate.usernamePlaceholder")} disabled={isEdit} />
        </Form.Item>

        <Form.Item name="email" label={t("admin.userCreate.email")} rules={[{ required: true, type: "email" }]}>
          <Input placeholder={t("admin.userCreate.emailPlaceholder")} disabled={isEmailDisabled} />
        </Form.Item>

        <Form.Item name="full_name" label={t("profile.fullName") || "Họ và Tên"} rules={[{ required: true }]}>
          <Input placeholder="Nhập họ và tên..." />
        </Form.Item>

        {!isEdit && (
          <Form.Item label={t("admin.userCreate.passwordInit")} required>
            <Space.Compact style={{ width: "100%" }}>
              <Form.Item name="password" noStyle rules={[{ required: true, message: t("admin.userCreate.requirePassword") }]}>
                <Input.Password value={generatedPassword} placeholder={t("admin.userCreate.passwordPlaceholder")} />
              </Form.Item>
              <Button icon={<SyncOutlined />} onClick={generatePassword}>Generate</Button>
              <Button icon={<CopyOutlined />} onClick={copyPassword} disabled={!generatedPassword}>Copy</Button>
            </Space.Compact>
          </Form.Item>
        )}

        <Form.Item name="department_id" label={t("admin.userCreate.department")} rules={[{ required: true }]}>
          <Select placeholder={t("admin.userCreate.deptPlaceholder")}>
            {departments?.map(d => <Select.Option key={d.id} value={d.id}>{d.name}</Select.Option>)}
          </Select>
        </Form.Item>

        <Form.Item name="role_ids" label={t("admin.userCreate.roles")} rules={[{ required: true }]}>
          <Select mode="multiple" placeholder={t("admin.roles.selectMultiplePlaceholder")}>
            {roles?.map(r => <Select.Option key={r.id} value={r.id}>{r.name}</Select.Option>)}
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
};
