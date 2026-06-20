import { useEffect, useRef, useCallback } from "react";
import { useQueueStore } from "@/store/queueStore";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/apiClient";
import { GetPublicKey, SignData } from "../../../wailsjs/go/main/App";

export function useQueueWS(token: string | null) {
  const { applyWSEvent } = useQueueStore();
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef(0);
  const DELAYS = [1000, 2000, 4000, 8000, 30000]; // exponential backoff

  const refreshAndReconnect = async () => {
    try {
      const publicKeyPem = await GetPublicKey();
      const refreshToken = useAuthStore.getState().refreshToken;
      
      if (!refreshToken) throw new Error("No refresh token");

      const signature = await SignData(refreshToken);
      const res = await apiClient.post("/auth/refresh", {
        refresh_token: refreshToken, 
        signature: signature, 
        public_key_pem: publicKeyPem 
      }, { 
        headers: { "X-Require-Signature": "false" } 
      });
      
      const newToken = res.data.data.access_token;
      const newRefreshToken = res.data.data.refresh_token;
      
      useAuthStore.getState().setTokens(newToken, newRefreshToken);
      
      // Reconnect with new token
      connect(newToken);
    } catch (err) {
      console.error("WS Token refresh failed:", err);
      useAuthStore.getState().clearAuth();
      window.location.hash = "/login";
    }
  };

  const connect = useCallback((wsToken: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8081/ws/queue";
    const ws = new WebSocket(`${wsUrl}?token=${wsToken}`);
    
    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data);
        applyWSEvent(event);
      } catch (err) {
        console.error("WS parse error:", err);
      }
    };

    ws.onclose = (e) => {
      // 1008 = Policy Violation, which we use for token expiration in custom fiber ws adapter
      if (e.code === 1008) {
        refreshAndReconnect();
        return;
      }

      const delay = DELAYS[Math.min(retryRef.current, DELAYS.length - 1)];
      retryRef.current++;
      setTimeout(() => connect(useAuthStore.getState().token || ""), delay);
    };

    ws.onopen = () => {
      retryRef.current = 0;
      console.log("Queue WS Connected");
    };

    wsRef.current = ws;
  }, [applyWSEvent]);

  useEffect(() => {
    if (token) {
      connect(token);
    }
    return () => {
      if (wsRef.current) {
        wsRef.current.onclose = null; // prevent reconnect on unmount
        wsRef.current.close();
      }
    };
  }, [token, connect]);
}
