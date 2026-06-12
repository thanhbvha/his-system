import React from 'react'
import {createRoot} from 'react-dom/client'
import { ConfigProvider } from "antd";
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/lib/queryClient";
import "@/i18n";
import "@/styles/global.css";
import "@/styles/typography.css";
import App from './App'

const container = document.getElementById('root')

const theme = {
  token: {
    colorPrimary: "#1677ff",
    colorSuccess: "#52c41a",
    colorWarning: "#faad14",
    colorError:   "#ff4d4f",
    borderRadius: 6,
    fontFamily:   "Inter, Roboto, sans-serif",
  },
};

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <QueryClientProvider client={queryClient}>
            <ConfigProvider theme={theme}>
                <App/>
            </ConfigProvider>
        </QueryClientProvider>
    </React.StrictMode>
)
