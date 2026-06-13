import { Card, Col, Row, Statistic, Spin } from "antd";
import { UserOutlined, SafetyCertificateOutlined, BankOutlined } from "@ant-design/icons";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/apiClient";

export const AdminDashboardPage = () => {
  // Fetch basic stats (we could have a dedicated endpoint, but we'll try to estimate or call multiple if needed)
  // For now, let's just make dummy queries to get meta data if backend supports it.
  
  const { data: usersData, isLoading: loadingUsers } = useQuery({
    queryKey: ["users", "count"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/users?limit=1");
      return res.data;
    }
  });

  const { data: rolesData, isLoading: loadingRoles } = useQuery({
    queryKey: ["roles"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/roles");
      return res.data.data;
    }
  });

  const { data: deptsData, isLoading: loadingDepts } = useQuery({
    queryKey: ["departments"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/departments");
      return res.data.data;
    }
  });

  if (loadingUsers || loadingRoles || loadingDepts) {
    return <div style={{ textAlign: "center", marginTop: 100 }}><Spin size="large" /></div>;
  }

  const totalUsers = usersData?.meta?.total || 0;
  const totalRoles = rolesData?.length || 0;
  const totalDepts = deptsData?.length || 0;

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>Admin Dashboard</h2>
      <Row gutter={16}>
        <Col span={8}>
          <Card bordered={false} style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.08)" }}>
            <Statistic
              title="Tổng số nhân viên"
              value={totalUsers}
              prefix={<UserOutlined style={{ color: "#1677ff" }} />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false} style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.08)" }}>
            <Statistic
              title="Tổng số Roles"
              value={totalRoles}
              prefix={<SafetyCertificateOutlined style={{ color: "#52c41a" }} />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false} style={{ boxShadow: "0 2px 8px rgba(0,0,0,0.08)" }}>
            <Statistic
              title="Tổng số Khoa/Phòng"
              value={totalDepts}
              prefix={<BankOutlined style={{ color: "#faad14" }} />}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};
