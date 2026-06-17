import { Button, Form, Input, Card, message } from "antd";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";
import { GetPublicKey, SignData } from "../../wailsjs/go/main/App";

export const Login = () => {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [loading, setLoading] = useState(false);

  const getRoleRoute = (role: string) => {
    const roleRouteMap: Record<string, string> = {
      receptionist: "/receptionist/queue",
      doctor: "/doctor/worklist",
      lab_tech: "/lab/worklist",
      pharmacist: "/pharmacy/prescriptions",
      admin: "/admin/dashboard",
    };
    return roleRouteMap[role] ?? "/";
  };

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      // Step 1: Init Login
      const initRes = await apiClient.post("/auth/login/init", {
        username: values.username,
        password: values.password,
      });

      const { challenge_string, mfa_required } = initRes.data.data;

      if (mfa_required) {
        // Navigate to MFA page
        navigate("/mfa", { state: { challenge_string, username: values.username } });
        return;
      }

      // Step 2: Hardware Signature
      const publicKeyPem = await GetPublicKey();
      const signature = await SignData(challenge_string);

      // Step 3: Complete Login
      const compRes = await apiClient.post("/auth/login/complete", {
        username: values.username,
        challenge_string,
        signature,
        public_key_pem: publicKeyPem,
      });

      const { access_token, refresh_token, user } = compRes.data.data;
      
      // Usually user object from backend doesn't contain 'role' explicitly at the top level or contains 'roles'
      // We'll extract role from the first role in the array
      const role = user.role_ids && user.role_ids.length > 0 ? "admin" : "receptionist"; // TODO: Map real roles
      
      setAuth(access_token, refresh_token, user, role as any);

      if (user.preferred_language) {
        i18n.changeLanguage(user.preferred_language);
      }

      if (user.mfa_enabled === false) {
        navigate("/mfa-setup");
      } else {
        navigate("/profile");
      }
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401) {
        message.error(t("auth.errors.invalidCredentials"));
      } else if (err.response?.status === 429) {
        message.error(t("auth.errors.tooManyAttempts"));
      } else if (err.response?.status === 423) {
        message.error(t("auth.errors.accountLocked"));
      } else {
        const backendMsg = err.response?.data?.error?.message || err.response?.data?.message;
        message.error(backendMsg || t("common.error"));
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh", backgroundColor: "var(--color-bg)" }}>
      <Card title={t("auth.loginTitle")} style={{ width: 350 }}>
        <Form name="login" onFinish={onFinish} layout="vertical">
          <Form.Item label={t("auth.username")} name="username" rules={[{ required: true, message: t("auth.errors.requireUsername") }]}>
            <Input disabled={loading} />
          </Form.Item>
          <Form.Item label={t("auth.password")} name="password" rules={[{ required: true, message: t("auth.errors.requirePassword") }]}>
            <Input.Password disabled={loading} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              {t("auth.loginBtn")}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};
