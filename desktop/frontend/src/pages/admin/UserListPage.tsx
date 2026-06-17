import { useState } from "react";
import { Table, Button, Input, Tag, Popconfirm, message, Space, Select, Avatar } from "antd";
import { UserOutlined, PlusOutlined, LockOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";
import { UserCreateModal } from "@/components/admin/UserCreateModal";
import { AssignRolesModal } from "@/components/admin/AssignRolesModal";

export const UserListPage = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  
  const [assignUser, setAssignUser] = useState<any>(null);

  // Fetch data
  const { data: usersData, isLoading } = useQuery({
    queryKey: ["users", page, search],
    queryFn: async () => {
      const res = await apiClient.get("/admin/users", { params: { page, limit: 10, search } });
      return res.data; // { data: [...], meta: { ... } }
    }
  });

  const { data: rolesData } = useQuery({
    queryKey: ["roles"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/roles");
      return res.data.data;
    }
  });

  const { data: deptsData } = useQuery({
    queryKey: ["departments"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/departments");
      return res.data.data;
    }
  });

  const deactivateMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.put(`/admin/users/${id}/deactivate`);
    },
    onSuccess: () => {
      message.success(t("admin.userList.statusChanged"));
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
    onError: () => message.error(t("admin.userList.updateError"))
  });

  const columns = [
    {
      title: t("admin.userList.avatar"),
      dataIndex: "username",
      key: "avatar",
      render: (val: string) => <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
    },
    { title: t("admin.userList.username"), dataIndex: "username", key: "username" },
    { title: t("admin.userList.email"), dataIndex: "email", key: "email" }, // Assuming backend decrypts and returns email
    { 
      title: t("admin.userList.roles"), 
      dataIndex: "roles", 
      key: "roles",
      render: (roles: any[]) => (
        <>
          {roles?.map(r => (
            <Tag color="blue" key={r.id || r}>{r.name || r}</Tag>
          ))}
        </>
      )
    },
    {
      title: t("admin.userList.status"),
      dataIndex: "is_active",
      key: "is_active",
      render: (isActive: boolean) => isActive ? <Tag color="green">{t("admin.userList.active")}</Tag> : <Tag color="red">{t("admin.userList.locked")}</Tag>
    },
    {
      title: t("admin.userList.actions"),
      key: "action",
      render: (_: any, record: any) => (
        <Space>
          <Button 
            size="small" 
            icon={<SafetyCertificateOutlined />} 
            onClick={() => setAssignUser(record)}
          >
            {t("admin.userList.assignRoles")}
          </Button>
          <Popconfirm
            title={t("admin.userList.confirmDeactivate")}
            onConfirm={() => deactivateMutation.mutate(record.id)}
            okText={t("common.yes")}
            cancelText={t("common.no")}
          >
            <Button size="small" danger icon={<LockOutlined />} disabled={!record.is_active}>
              {t("admin.userList.lock")}
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <h2>{t("admin.userList.title")}</h2>
        <Space>
          <Input.Search 
            placeholder={t("admin.userList.searchPlaceholder")} 
            onSearch={setSearch} 
            allowClear
            style={{ width: 300 }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsCreateOpen(true)}>
            {t("admin.userList.addUser")}
          </Button>
        </Space>
      </div>

      <Table 
        columns={columns} 
        dataSource={usersData?.data || []} 
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: usersData?.meta?.page || page,
          pageSize: usersData?.meta?.limit || 10,
          total: usersData?.meta?.total || 0,
          onChange: setPage
        }}
      />

      <UserCreateModal 
        open={isCreateOpen} 
        onClose={() => setIsCreateOpen(false)} 
        roles={rolesData || []}
        departments={deptsData || []}
      />

      <AssignRolesModal
        open={!!assignUser}
        onClose={() => setAssignUser(null)}
        user={assignUser}
        roles={rolesData || []}
      />
    </div>
  );
};
