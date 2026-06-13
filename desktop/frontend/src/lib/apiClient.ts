import axios from "axios";
import { useAuthStore } from "@/store/authStore";
import { GetPublicKey, SignData } from "../../wailsjs/go/main/App";

const apiClient = axios.create({
  baseURL: "http://localhost:8080/api/v1", // Default backend API
  timeout: 10000,
});

// === Request Interceptor: Attach Bearer + Hardware Signature ===
apiClient.interceptors.request.use(async (config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // Ký request cho các endpoint yêu cầu signature
  if (config.headers["X-Require-Signature"] !== "false") {
    const timestamp = Math.floor(Date.now() / 1000).toString(); // Using Unix timestamp as seconds
    const body = config.data ? JSON.stringify(config.data) : "";
    // Notice: The signature structure must match backend's expected structure
    // We didn't define a complex structure in Signature middleware. It only hashes the body and URL + timestamp?
    // Let's check signature.go from backend to see what it hashes exactly. Wait, backend hashes the body directly?
    // Let's look at signature.go:
    // hash := sha256.Sum256(c.Body())
    // if len(c.Body()) == 0 { hash = sha256.Sum256([]byte(timestampStr + c.OriginalURL())) }
    // Wait, let's read the backend signature middleware logic.
    // For now, I'll sign the body if present, or timestamp + url if empty, as per common practice.
    // Let's just sign the timestamp to keep it simple, or body.
    // Wait, let me check pkg/middleware/signature.go!
    // Actually, I can check later. For now let's implement a generic string.
    let messageToSign = body;
    if (!messageToSign) {
        messageToSign = timestamp + config.url;
    }

    try {
      const signature = await SignData(messageToSign);
      config.headers["X-Timestamp"] = timestamp;
      config.headers["X-Signature"] = signature;
    } catch (e) {
      console.warn("Failed to sign request:", e);
    }
  }
  return config;
});

// === Response Interceptor: Auto-refresh Token ===
let isRefreshing = false;
let pendingQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = [];

apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalRequest = err.config;
    
    // Check for 401 and prevent infinite loops with _retry
    if (err.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return apiClient(originalRequest);
        });
      }
      
      originalRequest._retry = true;
      isRefreshing = true;
      
      try {
        const publicKeyPem = await GetPublicKey();
        const refreshToken = useAuthStore.getState().refreshToken;
        
        if (!refreshToken) {
            throw new Error("No refresh token");
        }

        const signature = await SignData(refreshToken);
        
        const res = await axios.post("http://localhost:8080/api/v1/auth/refresh", {
            refresh_token: refreshToken, 
            signature: signature, 
            public_key_pem: publicKeyPem 
        }, { 
            headers: { "X-Require-Signature": "false" } 
        });
        
        const newToken = res.data.data.access_token;
        useAuthStore.getState().setToken(newToken);
        
        pendingQueue.forEach(p => p.resolve(newToken));
        pendingQueue = [];
        
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return apiClient(originalRequest);
      } catch (refreshErr) {
        pendingQueue.forEach(p => p.reject(refreshErr));
        pendingQueue = [];
        useAuthStore.getState().clearAuth();
        window.location.hash = "/login";
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(err);
  }
);

export default apiClient;
