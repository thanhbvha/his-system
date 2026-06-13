import { Button, Form, Input, Card, message } from "antd";
import { useNavigate, useLocation, Navigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";
import { useState } from "react";
import apiClient from "@/lib/apiClient";
import { GetPublicKey, SignData } from "../../wailsjs/go/main/App";

export const MFAPage = () => {
  const navigate = useNavigate();
  const location = useLocation();
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
        username,
        code: values.code,
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
      const role = "admin"; // Mock for now
      
      setAuth(access_token, refresh_token, user, role as any);
      navigate(getRoleRoute(role));
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401 || err.response?.status === 400) {
        message.error("Mã xác thực không đúng");
      } else {
        message.error("Lỗi hệ thống");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh", backgroundColor: "var(--color-bg)" }}>
      <Card title="Xác thực 2 bước (MFA)" style={{ width: 350 }}>
        <p>Vui lòng nhập mã 6 số từ ứng dụng Authenticator của bạn.</p>
        <Form name="mfa" onFinish={onFinish} layout="vertical">
          <Form.Item label="Mã xác thực" name="code" rules={[{ required: true, len: 6, message: "Mã phải có 6 chữ số!" }]}>
            <Input disabled={loading} maxLength={6} style={{ textAlign: 'center', letterSpacing: '0.5em', fontSize: '1.2em' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Xác nhận
            </Button>
          </Form.Item>
          <div style={{ textAlign: "center", marginTop: 16 }}>
            <Button type="link" onClick={() => message.info("Chức năng dùng mã dự phòng chưa triển khai")}>Dùng mã dự phòng</Button>
          </div>
        </Form>
      </Card>
    </div>
  );
};
