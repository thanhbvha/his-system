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
    
    // Extract path and query from full URL to match backend's c.OriginalURL()
    const fullUrl = apiClient.getUri(config);
    // getUri might return a relative URL if baseURL is not absolute, but here baseURL is http://localhost:8080/api/v1
    const urlObj = new URL(fullUrl, window.location.origin);
    const pathAndQuery = urlObj.pathname + urlObj.search;
    
    const method = (config.method || "GET").toUpperCase();
    
    let bodyStr = "";
    if (config.data) {
      if (typeof config.data === "string") {
        bodyStr = config.data;
      } else {
        bodyStr = JSON.stringify(config.data);
      }
    }

    const messageToSign = `${method}${pathAndQuery}${timestamp}${bodyStr}`;

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
        });
        
        const newToken = res.data.data.access_token;
        const newRefreshToken = res.data.data.refresh_token;
        
        // Save BOTH tokens because backend rotates them!
        useAuthStore.getState().setTokens(newToken, newRefreshToken);
        
        pendingQueue.forEach(p => p.resolve(newToken));
        pendingQueue = [];
        
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        // Since we are returning the promise, we MUST await it to catch its errors in this try-catch block!
        return await apiClient(originalRequest);
      } catch (refreshErr) {
        console.error("Refresh token failed:", refreshErr);
        pendingQueue.forEach(p => p.reject(refreshErr));
        pendingQueue = [];
        useAuthStore.getState().clearAuth();
        window.location.hash = "/login";
        // Return the original 401 error so the calling function can handle it properly
        return Promise.reject(err);
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(err);
  }
);

export default apiClient;
