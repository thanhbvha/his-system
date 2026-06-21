import { Card, Descriptions, Avatar, Tag, Typography, Row, Col, Spin } from "antd";
import { UserOutlined, SafetyCertificateOutlined, GlobalOutlined, EnvironmentOutlined } from "@ant-design/icons";
import { useAuthStore } from "@/store/authStore";
import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/apiClient";

const { Title, Text } = Typography;

export const ProfilePage = () => {
  const { t } = useTranslation();
  const { user, role } = useAuthStore();

  const getRoleColor = (r: string | null) => {
    switch (r) {
      case "admin": return "red";
      case "doctor": return "blue";
      case "receptionist": return "green";
      default: return "default";
    }
  };

  const { data: profileData, isLoading } = useQuery({
    queryKey: ["auth_me"],
    queryFn: async () => {
      const res = await apiClient.get("/auth/me");
      return res.data.data;
    },
    enabled: !!user,
  });

  if (!user) return null;

  if (isLoading) {
    return <div style={{ textAlign: "center", padding: "50px" }}><Spin size="large" /></div>;
  }

  // Mocking some data for visual completeness
  const loginTime = new Date().toLocaleString();
  const ipAddress = "192.168.1.100 (Local)"; 

  const displayName = profileData?.staff_profile?.full_name || profileData?.username || user.username || t("profile.unknown");
  const departmentName = profileData?.department?.name || t("profile.defaultDepartment");
  const avatarUrl = profileData?.staff_profile?.avatar_url || null;

  return (
    <div style={{ maxWidth: 800, margin: "0 auto" }}>
      <Title level={2} style={{ marginBottom: 24 }}>
        <UserOutlined /> {t("profile.title")}
      </Title>

      <Row gutter={[24, 24]}>
        <Col span={24} md={8}>
          <Card style={{ textAlign: "center", borderRadius: 12, boxShadow: "0 4px 12px rgba(0,0,0,0.05)" }}>
            {avatarUrl ? (
              <Avatar size={100} src={avatarUrl} style={{ marginBottom: 16 }} />
            ) : (
              <Avatar size={100} icon={<UserOutlined />} style={{ backgroundColor: "#1890ff", marginBottom: 16 }} />
            )}
            <Title level={4} style={{ margin: 0 }}>{displayName}</Title>
            {profileData?.staff_profile?.title && (
              <Text type="secondary" style={{ display: "block", marginBottom: 8 }}>{profileData.staff_profile.title}</Text>
            )}
            <div style={{ marginTop: 16 }}>
              <Tag color={getRoleColor(role)} style={{ fontSize: 14, padding: "4px 12px", borderRadius: 16 }}>
                {role ? role.toUpperCase() : t("profile.noRole")}
              </Tag>
            </div>
          </Card>
        </Col>

        <Col span={24} md={16}>
          <Card 
            title={<><SafetyCertificateOutlined /> {t("profile.systemInfo")}</>} 
            style={{ borderRadius: 12, boxShadow: "0 4px 12px rgba(0,0,0,0.05)", marginBottom: 16 }}
          >
            <Descriptions column={1} bordered size="small" labelStyle={{ width: "200px" }}>
              <Descriptions.Item label={t("profile.systemId")}>{user.id}</Descriptions.Item>
              <Descriptions.Item label={t("profile.username")}>{profileData?.username || user.username}</Descriptions.Item>
              <Descriptions.Item label={t("profile.department")}>{departmentName}</Descriptions.Item>
              <Descriptions.Item label={t("profile.mfa")}>
                {user.mfa_enabled ? <Tag color="success">{t("profile.mfaEnabled")}</Tag> : <Tag color="warning">{t("profile.mfaDisabled")}</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label={t("profile.preferredLanguage")}>
                <Tag color="geekblue">{user.preferred_language === "en" ? "English" : "Tiếng Việt"}</Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card 
            title={<><GlobalOutlined /> {t("profile.currentSession")}</>} 
            style={{ borderRadius: 12, boxShadow: "0 4px 12px rgba(0,0,0,0.05)" }}
          >
            <Descriptions column={1} bordered size="small" labelStyle={{ width: "200px" }}>
              <Descriptions.Item label={t("profile.loginTime")}>{loginTime}</Descriptions.Item>
              <Descriptions.Item label={t("profile.ipAddress")}><EnvironmentOutlined /> {ipAddress}</Descriptions.Item>
              <Descriptions.Item label={t("profile.status")}><Tag color="processing">{t("profile.active")}</Tag></Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>
    </div>
  );
};
