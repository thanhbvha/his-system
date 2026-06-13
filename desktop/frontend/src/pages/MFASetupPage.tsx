import { Button, Form, Input, Card, message, Typography, List } from "antd";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import apiClient from "@/lib/apiClient";
import { QRCodeSVG } from "qrcode.react"; // Requires: npm install qrcode.react

import { useAuthStore } from "@/store/authStore";

const { Title, Text, Paragraph } = Typography;

export const MFASetupPage = () => {
  const navigate = useNavigate();
  const user = useAuthStore(s => s.user);
  const updateAuthUser = useAuthStore(s => s.updateAuthUser);
  const [loading, setLoading] = useState(false);
  const [setupData, setSetupData] = useState<{ qr_uri: string; backup_codes: string[] } | null>(null);

  useEffect(() => {
    // Call setup endpoint on load
    apiClient.post("/auth/mfa/setup")
      .then(res => {
        setSetupData(res.data.data);
      })
      .catch(err => {
        message.error("Lỗi khi tải cấu hình MFA");
        console.error(err);
      });
  }, []);

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      await apiClient.post("/auth/mfa/verify", {
        user_id: user?.username,
        totp_code: values.code
      });

      // Cập nhật lại state trong store
      if (updateAuthUser && user) {
        updateAuthUser({ ...user, mfa_enabled: true });
      }

      message.success("Thiết lập MFA thành công");
      
      const role = user?.role_ids && user.role_ids.length > 0 ? "admin" : "receptionist";
      const getRoleRoute = (r: string) => {
        const map: Record<string, string> = {
          receptionist: "/receptionist/queue",
          doctor: "/doctor/worklist",
          admin: "/admin/dashboard",
        };
        return map[r] ?? "/";
      };

      navigate(getRoleRoute(role)); // Redirect to correct dashboard
    } catch (err: any) {
      console.error(err);
      message.error("Mã xác thực không đúng");
    } finally {
      setLoading(false);
    }
  };

  if (!setupData) {
    return <div style={{ textAlign: "center", marginTop: 100 }}>Loading MFA Setup...</div>;
  }

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", backgroundColor: "var(--color-bg)", padding: 20 }}>
      <Card style={{ maxWidth: 500 }}>
        <Title level={3} style={{ textAlign: "center" }}>Thiết lập Xác thực 2 bước</Title>
        
        <div style={{ textAlign: "center", margin: "20px 0" }}>
          <QRCodeSVG value={setupData.qr_uri} size={200} />
        </div>

        <Paragraph>
          1. Cài đặt ứng dụng Google Authenticator hoặc Authy.<br />
          2. Quét mã QR ở trên để thêm tài khoản.<br />
          3. Lưu lại các mã dự phòng (Backup Codes) ở nơi an toàn.
        </Paragraph>

        <div style={{ backgroundColor: "#f5f5f5", padding: 15, borderRadius: 8, marginBottom: 20 }}>
          <Text strong>Backup Codes (Chỉ hiển thị 1 lần duy nhất):</Text>
          <List
            size="small"
            dataSource={setupData.backup_codes}
            renderItem={(item) => <List.Item><Text code>{item}</Text></List.Item>}
            style={{ marginTop: 10 }}
          />
        </div>

        <Form name="mfa_setup" onFinish={onFinish} layout="vertical">
          <Form.Item label="Nhập mã 6 số để xác nhận" name="code" rules={[{ required: true, len: 6, message: "Mã phải có 6 chữ số!" }]}>
            <Input disabled={loading} maxLength={6} style={{ textAlign: 'center', letterSpacing: '0.5em', fontSize: '1.2em' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              Xác nhận & Hoàn tất
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};
