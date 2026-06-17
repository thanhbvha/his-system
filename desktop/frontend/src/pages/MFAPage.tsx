import { Button, Form, Input, Card, message } from "antd";
import { useNavigate, useLocation, Navigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";
import { GetPublicKey, SignData } from "../../wailsjs/go/main/App";

export const MFAPage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { t } = useTranslation();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [loading, setLoading] = useState(false);

  const state = location.state as { challenge_string: string; username: string } | null;
  if (!state) {
    return <Navigate to="/login" replace />;
  }

  const { challenge_string, username } = state;

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
      // Step 1: Verify MFA
      const verifyRes = await apiClient.post("/auth/mfa/verify", {
        user_id: username,
        totp_code: values.code,
      });

      const { mfa_token } = verifyRes.data.data;

      // Step 2: Hardware Signature
      const publicKeyPem = await GetPublicKey();
      const signature = await SignData(challenge_string);

      // Step 3: Complete Login
      const compRes = await apiClient.post("/auth/login/complete", {
        username,
        challenge_string,
        signature,
        public_key_pem: publicKeyPem,
        mfa_token,
      });

      const { access_token, refresh_token, user } = compRes.data.data;
      const role = user.role_ids && user.role_ids.length > 0 ? "admin" : "receptionist";
      
      setAuth(access_token, refresh_token, user, role as any);
      navigate("/profile");
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401 || err.response?.status === 400) {
        message.error(t("auth.errors.invalidMFA"));
      } else {
        message.error(t("common.error"));
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh", backgroundColor: "var(--color-bg)" }}>
      <Card title={t("auth.mfaTitle")} style={{ width: 350 }}>
        <p>{t("auth.mfaInstruction")}</p>
        <Form name="mfa" onFinish={onFinish} layout="vertical">
          <Form.Item label={t("auth.mfaCode")} name="code" rules={[{ required: true, len: 6, message: t("auth.errors.require6Digits") }]}>
            <Input disabled={loading} maxLength={6} style={{ textAlign: 'center', letterSpacing: '0.5em', fontSize: '1.2em' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              {t("common.confirm")}
            </Button>
          </Form.Item>
          <div style={{ textAlign: "center", marginTop: 16 }}>
            <Button type="link" onClick={() => message.info(t("auth.backupCodeNotImplemented"))}>{t("auth.useBackupCode")}</Button>
          </div>
        </Form>
      </Card>
    </div>
  );
};
