import { Card, Descriptions, Avatar, Tag, Typography, Row, Col } from "antd";
import { UserOutlined, SafetyCertificateOutlined, GlobalOutlined, EnvironmentOutlined } from "@ant-design/icons";
import { useAuthStore } from "@/store/authStore";
import { useTranslation } from "react-i18next";

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

  if (!user) return null;

  // Mocking some data for visual completeness as requested in sprint plan
  const loginTime = new Date().toLocaleString();
  const ipAddress = "192.168.1.100 (Local)"; 

  return (
    <div style={{ maxWidth: 800, margin: "0 auto" }}>
      <Title level={2} style={{ marginBottom: 24 }}>
        <UserOutlined /> {t("profile.title")}
      </Title>

      <Row gutter={[24, 24]}>
        <Col span={24} md={8}>
          <Card style={{ textAlign: "center", borderRadius: 12, boxShadow: "0 4px 12px rgba(0,0,0,0.05)" }}>
            <Avatar size={100} icon={<UserOutlined />} style={{ backgroundColor: "#1890ff", marginBottom: 16 }} />
            <Title level={4} style={{ margin: 0 }}>{user.name || user.username || t("profile.unknown")}</Title>
            <Text type="secondary">{user.email || t("profile.noEmail")}</Text>
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
              <Descriptions.Item label={t("profile.username")}>{user.username}</Descriptions.Item>
              <Descriptions.Item label={t("profile.department")}>{t("profile.defaultDepartment")}</Descriptions.Item>
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
