import { useEffect } from "react";
import { Layout, Menu, Button } from "antd";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/store/authStore";
import { useUIStore } from "@/store/uiStore";
import apiClient from "@/lib/apiClient";
import { 
  MenuFoldOutlined, 
  MenuUnfoldOutlined, 
  LogoutOutlined,
  DashboardOutlined,
  TeamOutlined,
  SafetyOutlined,
  BankOutlined,
  SolutionOutlined,
  CalendarOutlined,
  UserOutlined,
  OrderedListOutlined
} from "@ant-design/icons";

const { Header, Sider, Content } = Layout;

export const RoleLayout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const role = useAuthStore((s) => s.role);
  const user = useAuthStore((s) => s.user);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const updateAuthUser = useAuthStore((s) => s.updateAuthUser);
  const { sidebarOpen, toggleSidebar } = useUIStore();

  useEffect(() => {
    if (user?.preferred_language && i18n.language !== user.preferred_language) {
      i18n.changeLanguage(user.preferred_language);
    }
  }, [user?.preferred_language, i18n]);

  const handleLogout = () => {
    clearAuth();
    navigate("/login");
  };

  const handleLanguageChange = async () => {
    const newLang = i18n.language === "vi" ? "en" : "vi";
    i18n.changeLanguage(newLang);
    try {
      await apiClient.put("/auth/me/language", { language: newLang });
      updateAuthUser({ preferred_language: newLang });
    } catch (e) {
      console.error("Failed to sync language:", e);
    }
  };

  const getMenuItems = () => {
    const baseItems: any[] = [];
    
    if (role === "admin") {
       baseItems.push(
         { key: "/admin/dashboard", label: t("nav.adminDashboard"), icon: <DashboardOutlined /> },
         { key: "/admin/users", label: t("nav.manageUsers"), icon: <TeamOutlined /> },
         { key: "/admin/roles", label: t("nav.roles"), icon: <SafetyOutlined /> },
         { key: "/admin/departments", label: t("nav.departments"), icon: <BankOutlined /> },
         { type: 'divider' },
         { key: "/patients", label: t("nav.patients", "Bệnh nhân"), icon: <SolutionOutlined /> },
         { key: "/appointments", label: t("nav.appointments", "Lịch hẹn"), icon: <CalendarOutlined /> },
         { key: "/queue", label: t("queue.title", "Hàng đợi hôm nay"), icon: <OrderedListOutlined /> }
       );
    } else if (role === "doctor" || role === "receptionist") {
       baseItems.push(
         { key: "/", label: t("nav.dashboard"), icon: <DashboardOutlined /> },
         { key: "/patients", label: t("nav.patients", "Bệnh nhân"), icon: <SolutionOutlined /> },
         { key: "/appointments", label: t("nav.appointments", "Lịch hẹn"), icon: <CalendarOutlined /> },
         { key: "/queue", label: t("queue.title", "Hàng đợi hôm nay"), icon: <OrderedListOutlined /> }
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
            <Button icon={<UserOutlined />} onClick={() => navigate("/profile")} style={{ marginRight: 16 }}>
              {t("nav.profile")}
            </Button>
            <Button onClick={handleLanguageChange} style={{ marginRight: 16 }}>
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
