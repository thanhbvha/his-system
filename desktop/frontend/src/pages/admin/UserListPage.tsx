import { useState } from "react";
import { Table, Button, Input, Tag, Popconfirm, message, Space, Select, Avatar } from "antd";
import { UserOutlined, PlusOutlined, LockOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/apiClient";
import { UserCreateModal } from "@/components/admin/UserCreateModal";
import { AssignRolesModal } from "@/components/admin/AssignRolesModal";

export const UserListPage = () => {
  const queryClient = useQueryClient();
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
      return res.data;
    }
  });

  const { data: deptsData } = useQuery({
    queryKey: ["departments"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/departments");
      return res.data;
    }
  });

  const deactivateMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.put(`/admin/users/${id}/deactivate`);
    },
    onSuccess: () => {
      message.success("Đã thay đổi trạng thái tài khoản");
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
    onError: () => message.error("Lỗi khi cập nhật")
  });

  const columns = [
    {
      title: "Avatar",
      dataIndex: "username",
      key: "avatar",
      render: (val: string) => <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
    },
    { title: "Username", dataIndex: "username", key: "username" },
    { title: "Email", dataIndex: "email", key: "email" }, // Assuming backend decrypts and returns email
    { 
      title: "Roles", 
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
      title: "Status",
      dataIndex: "is_active",
      key: "is_active",
      render: (isActive: boolean) => isActive ? <Tag color="green">Hoạt động</Tag> : <Tag color="red">Đã khóa</Tag>
    },
    {
      title: "Thao tác",
      key: "action",
      render: (_: any, record: any) => (
        <Space>
          <Button 
            size="small" 
            icon={<SafetyCertificateOutlined />} 
            onClick={() => setAssignUser(record)}
          >
            Phân quyền
          </Button>
          <Popconfirm
            title="Bạn có chắc chắn muốn vô hiệu hoá tài khoản này?"
            onConfirm={() => deactivateMutation.mutate(record.id)}
            okText="Đồng ý"
            cancelText="Hủy"
          >
            <Button size="small" danger icon={<LockOutlined />} disabled={!record.is_active}>
              Khóa
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <h2>Quản lý Nhân viên (User)</h2>
        <Space>
          <Input.Search 
            placeholder="Tìm theo username/email" 
            onSearch={setSearch} 
            allowClear
            style={{ width: 300 }}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsCreateOpen(true)}>
            Thêm User
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
        roles={rolesData?.data || []}
        departments={deptsData?.data || []}
      />

      <AssignRolesModal
        open={!!assignUser}
        onClose={() => setAssignUser(null)}
        user={assignUser}
        roles={rolesData?.data || []}
      />
    </div>
  );
};
