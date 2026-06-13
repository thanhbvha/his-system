import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useLocation, Navigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/apiClient";

const registerSchema = z.object({
  full_name: z.string().min(2, "Tên phải có ít nhất 2 ký tự"),
  dob: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Định dạng ngày sinh phải là YYYY-MM-DD"),
  gender: z.enum(["male", "female", "other"], { message: "Vui lòng chọn giới tính" }),
  email: z.string().email("Email không hợp lệ").optional().or(z.literal("")),
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export const RegisterPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((s) => s.setAuth);

  const phone = location.state?.phone;
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      gender: "other",
    }
  });

  if (!phone) {
    return <Navigate to="/login" replace />;
  }

  const onSubmit = async (data: RegisterFormValues) => {
    setLoading(true);
    setErrorMsg("");
    try {
      const payload = {
        phone,
        full_name: data.full_name,
        dob: data.dob,
        gender: data.gender,
        email: data.email || undefined,
      };

      const res = await apiClient.post("/auth/register", payload);
      const resData = res.data.data;

      setAuth(resData.access_token, resData.patient);
      navigate("/dashboard", { replace: true });
    } catch (err: any) {
      if (err.response?.status === 409) {
        setErrorMsg(t("auth.errors.phoneExists", "Số điện thoại đã được đăng ký. Vui lòng đăng nhập."));
      } else {
        setErrorMsg(t("common.error"));
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-[80vh] items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl text-center">{t("auth.registerTitle", "Hoàn tất đăng ký")}</CardTitle>
          <CardDescription className="text-center">
            {t("auth.registerDesc", "Cập nhật thông tin hồ sơ cho số điện thoại: ")} <span className="font-bold">{phone}</span>
          </CardDescription>
        </CardHeader>
        <CardContent>
          {errorMsg && <div className="mb-4 text-sm text-red-500 font-medium text-center">{errorMsg}</div>}
          
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="full_name">{t("auth.fullName", "Họ và tên")}</Label>
              <Input id="full_name" {...register("full_name")} disabled={loading} placeholder="Nguyễn Văn A" />
              {errors.full_name && <p className="text-sm text-red-500">{errors.full_name.message}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="dob">{t("auth.dob", "Ngày sinh (YYYY-MM-DD)")}</Label>
              <Input id="dob" type="date" {...register("dob")} disabled={loading} />
              {errors.dob && <p className="text-sm text-red-500">{errors.dob.message}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="gender">{t("auth.gender", "Giới tính")}</Label>
              <select 
                id="gender" 
                {...register("gender")} 
                disabled={loading}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="male">{t("auth.genderMale", "Nam")}</option>
                <option value="female">{t("auth.genderFemale", "Nữ")}</option>
                <option value="other">{t("auth.genderOther", "Khác")}</option>
              </select>
              {errors.gender && <p className="text-sm text-red-500">{errors.gender.message}</p>}
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">{t("auth.email", "Email (Không bắt buộc)")}</Label>
              <Input id="email" type="email" {...register("email")} disabled={loading} />
              {errors.email && <p className="text-sm text-red-500">{errors.email.message}</p>}
            </div>

            <Button type="submit" className="w-full mt-4" disabled={loading}>
              {loading ? t("common.loading") : t("auth.createAccount", "Tạo tài khoản")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
};
