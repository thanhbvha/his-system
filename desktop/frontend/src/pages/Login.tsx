import { Button, Form, Input, Card } from "antd";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "@/store/authStore";

export const Login = () => {
  const navigate = useNavigate();
  const setAuth = useAuthStore((s) => s.setAuth);

  const onFinish = (values: any) => {
    console.log("Success:", values);
    // Mock login for Sprint 1
    setAuth("dummy_token", { id: "1", name: values.username, email: "test@his.com" }, "admin");
    navigate("/");
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", height: "100vh", backgroundColor: "var(--color-bg)" }}>
      <Card title="HIS Login" style={{ width: 350 }}>
        <Form name="login" onFinish={onFinish} layout="vertical">
          <Form.Item label="Username" name="username" rules={[{ required: true, message: "Please input your username!" }]}>
            <Input />
          </Form.Item>
          <Form.Item label="Password" name="password" rules={[{ required: true, message: "Please input your password!" }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              Submit
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};
