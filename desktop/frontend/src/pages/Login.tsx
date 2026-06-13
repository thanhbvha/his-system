import { Button, Form, Input, Card, message } from "antd";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";
import { useState } from "react";
import apiClient from "@/lib/apiClient";
import { GetPublicKey, SignData } from "../../wailsjs/go/main/App";

export const Login = () => {
  const navigate = useNavigate();
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

      if (user.mfa_enabled === false) {
        navigate("/mfa-setup");
      } else {
        navigate(getRoleRoute(role));
      }
    } catch (err: any) {
      console.error(err);
      if (err.response?.status === 401) {
        message.error("Tên đăng nhập hoặc mật khẩu không đúng");
      } else if (err.response?.status === 429) {
        message.error("Quá nhiều lần thử. Vui lòng đợi vài phút.");
      } else if (err.response?.status === 423) {
        message.error("Tài khoản bị khóa. Liên hệ quản trị viên.");
      } else {
        message.error("Lỗi hệ thống");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh", backgroundColor: "var(--color-bg)" }}>
      <Card title="HIS Login" style={{ width: 350 }}>
        <Form name="login" onFinish={onFinish} layout="vertical">
          <Form.Item label="Username" name="username" rules={[{ required: true, message: "Please input your username!" }]}>
            <Input disabled={loading} />
          </Form.Item>
          <Form.Item label="Password" name="password" rules={[{ required: true, message: "Please input your password!" }]}>
            <Input.Password disabled={loading} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Đăng nhập
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};
