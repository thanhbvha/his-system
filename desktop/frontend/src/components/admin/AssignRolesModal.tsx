import { useEffect } from "react";
import { Modal, Form, Select, message } from "antd";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

interface AssignRolesModalProps {
  open: boolean;
  onClose: () => void;
  user: any;
  roles: any[];
}

export const AssignRolesModal = ({ open, onClose, user, roles }: AssignRolesModalProps) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  useEffect(() => {
    if (open && user) {
      // API might return role strings or role objects, depending on backend ListUsers
      const currentRoleIds = user.roles?.map((r: any) => r.id || r) || [];
      form.setFieldsValue({ role_ids: currentRoleIds });
    }
  }, [open, user, form]);

  const updateMutation = useMutation({
    mutationFn: async (values: any) => {
      await apiClient.put(`/admin/users/${user.id}/roles`, values);
    },
    onSuccess: () => {
      message.success(t("admin.roles.updateSuccess"));
      queryClient.invalidateQueries({ queryKey: ["users"] });
      onClose();
    },
    onError: (err: any) => {
      message.error(err.response?.data?.message || t("admin.roles.updateFail"));
    }
  });

  return (
    <Modal
      title={t("admin.roles.assignTo", { username: user?.username })}
      open={open}
      onCancel={onClose}
      onOk={() => form.submit()}
      confirmLoading={updateMutation.isPending}
    >
      <Form form={form} layout="vertical" onFinish={(values) => updateMutation.mutate(values)}>
        <Form.Item name="role_ids" label={t("admin.roles.selectRoles")} rules={[{ required: true }]}>
          <Select mode="multiple" placeholder={t("admin.roles.selectMultiplePlaceholder")}>
            {roles?.map(r => <Select.Option key={r.id} value={r.id}>{r.name}</Select.Option>)}
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
};
