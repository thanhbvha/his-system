import { Layout, Menu, Button } from "antd";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/store/authStore";
import { useUIStore } from "@/store/uiStore";
import { 
  MenuFoldOutlined, 
  MenuUnfoldOutlined, 
  LogoutOutlined,
  DashboardOutlined,
  TeamOutlined,
  SafetyOutlined,
  BankOutlined
} from "@ant-design/icons";

const { Header, Sider, Content } = Layout;

export const RoleLayout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const role = useAuthStore((s) => s.role);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const { sidebarOpen, toggleSidebar } = useUIStore();

  const handleLogout = () => {
    clearAuth();
    navigate("/login");
  };

  const getMenuItems = () => {
    const baseItems: any[] = [];
    
    if (role === "admin") {
       baseItems.push(
         { key: "/admin/dashboard", label: "Dashboard", icon: <DashboardOutlined /> },
         { key: "/admin/users", label: "Quản lý User", icon: <TeamOutlined /> },
         { key: "/admin/roles", label: "Phân quyền", icon: <SafetyOutlined /> },
         { key: "/admin/departments", label: "Khoa/Phòng ban", icon: <BankOutlined /> }
       );
    } else if (role === "doctor") {
       baseItems.push(
         { key: "/", label: t("nav.dashboard") },
         { key: "/patients", label: t("nav.patients") },
         { key: "/appointments", label: t("nav.appointments") }
       );
    } else {
       baseItems.push({ key: "/", label: t("nav.dashboard") });
    }

    return baseItems;
  };

  // set default key to first item if possible
  const items = getMenuItems();
  const defaultKey = items.length > 0 ? items[0].key : "/";

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sider trigger={null} collapsible collapsed={!sidebarOpen}>
        <div style={{ height: 64, margin: 16, color: "white", fontSize: 20, fontWeight: "bold", textAlign: "center" }}>
          HIS App
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[location.pathname]} onClick={(e) => navigate(e.key)} items={items} />
      </Sider>
      <Layout>
        <Header style={{ padding: 0, background: "#fff", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <Button type="text" icon={sidebarOpen ? <MenuFoldOutlined /> : <MenuUnfoldOutlined />} onClick={toggleSidebar} style={{ fontSize: "16px", width: 64, height: 64 }} />
          <div style={{ paddingRight: 24 }}>
            <Button onClick={() => i18n.changeLanguage(i18n.language === "vi" ? "en" : "vi")} style={{ marginRight: 16 }}>
              {i18n.language === "vi" ? "EN" : "VI"}
            </Button>
            <Button icon={<LogoutOutlined />} onClick={handleLogout}>{t("common.logout")}</Button>
          </div>
        </Header>
        <Content style={{ margin: "24px 16px", padding: 24, minHeight: 280, background: "#fff", borderRadius: 8, overflowY: "auto" }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};
