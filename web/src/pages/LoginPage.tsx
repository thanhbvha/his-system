import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { InputOTP, InputOTPGroup, InputOTPSlot, InputOTPSeparator } from "@/components/ui/input-otp";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/apiClient";

export const LoginPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((s) => s.setAuth);

  const [phase, setPhase] = useState<"enter_phone" | "enter_otp">("enter_phone");
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");
  
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    let timer: any;
    if (countdown > 0) {
      timer = setTimeout(() => setCountdown(c => c - 1), 1000);
    }
    return () => clearTimeout(timer);
  }, [countdown]);

  const handleSendOTP = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!phone || phone.length < 10 || !phone.startsWith("0")) {
      setErrorMsg(t("auth.errors.invalidPhone", "Số điện thoại không hợp lệ"));
      return;
    }

    setLoading(true);
    setErrorMsg("");
    try {
      await apiClient.post("/auth/otp/send", { phone });
      setPhase("enter_otp");
      setCountdown(60);
      setOtp("");
    } catch (err: any) {
      if (err.response?.status === 429) {
        setErrorMsg(t("auth.errors.tooManyAttempts"));
      } else {
        setErrorMsg(t("common.error"));
      }
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOTP = async (otpValue: string) => {
    if (otpValue.length !== 6) return;
    setLoading(true);
    setErrorMsg("");

    try {
      const res = await apiClient.post("/auth/otp/verify", { phone, code: otpValue });
      const data = res.data.data;

      if (data.needs_register) {
        navigate("/register", { state: { phone } });
        return;
      }

      setAuth(data.access_token, data.patient);
      
      const returnUrl = location.state?.returnUrl || "/dashboard";
      navigate(returnUrl, { replace: true });

    } catch (err: any) {
      if (err.response?.status === 401 || err.response?.status === 400) {
        setErrorMsg(t("auth.errors.invalidOTP", "Mã OTP không đúng"));
      } else if (err.response?.status === 410) {
        setErrorMsg(t("auth.errors.expiredOTP", "Mã OTP đã hết hạn"));
        setCountdown(0);
      } else if (err.response?.status === 429) {
        setErrorMsg(t("auth.errors.tooManyAttempts"));
      } else {
        setErrorMsg(t("common.error"));
      }
      setOtp(""); // reset input so user can type again
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-[80vh] items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl text-center">{t("auth.loginTitle")}</CardTitle>
          <CardDescription className="text-center">
            {phase === "enter_phone" 
              ? t("auth.enterPhoneDesc", "Nhập số điện thoại để nhận mã xác thực") 
              : t("auth.enterOTPDesc", "Mã 6 số đã được gửi tới " + phone)}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {errorMsg && <div className="mb-4 text-sm text-red-500 font-medium text-center">{errorMsg}</div>}
          
          {phase === "enter_phone" && (
            <form onSubmit={handleSendOTP} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="phone">{t("auth.phone", "Số điện thoại")}</Label>
                <Input
                  id="phone"
                  placeholder="09..."
                  value={phone}
                  onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                  disabled={loading}
                  maxLength={11}
                  autoFocus
                />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? t("common.loading") : t("auth.sendOTP", "Gửi mã OTP")}
              </Button>
            </form>
          )}

          {phase === "enter_otp" && (
            <div className="flex flex-col items-center space-y-6">
              <InputOTP 
                maxLength={6} 
                value={otp} 
                onChange={setOtp} 
                onComplete={handleVerifyOTP}
                disabled={loading}
              >
                <InputOTPGroup>
                  <InputOTPSlot index={0} />
                  <InputOTPSlot index={1} />
                  <InputOTPSlot index={2} />
                </InputOTPGroup>
                <InputOTPSeparator />
                <InputOTPGroup>
                  <InputOTPSlot index={3} />
                  <InputOTPSlot index={4} />
                  <InputOTPSlot index={5} />
                </InputOTPGroup>
              </InputOTP>

              <div className="text-sm text-slate-500 text-center space-y-2">
                {countdown > 0 ? (
                  <p>{t("auth.resendIn", "Gửi lại sau")} <span className="font-medium text-slate-900">{countdown}s</span></p>
                ) : (
                  <Button variant="link" className="p-0 h-auto" onClick={() => handleSendOTP()} disabled={loading}>
                    {t("auth.resend", "Gửi lại mã OTP")}
                  </Button>
                )}
                <div>
                  <Button variant="link" className="p-0 h-auto text-slate-400" onClick={() => { setPhase("enter_phone"); setOtp(""); setCountdown(0); }} disabled={loading}>
                    {t("auth.changePhone", "Đổi số điện thoại")}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
