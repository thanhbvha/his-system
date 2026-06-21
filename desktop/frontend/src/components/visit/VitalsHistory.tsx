import React from "react";
import { Timeline, Typography, Card } from "antd";
import { useTranslation } from "react-i18next";
import { VisitVital } from "@/store/visitStore";

const { Text } = Typography;

interface VitalsHistoryProps {
  vitals: VisitVital[];
}

export const VitalsHistory: React.FC<VitalsHistoryProps> = ({ vitals }) => {
  const { t } = useTranslation();

  if (!vitals || vitals.length === 0) {
    return (
      <Card title={t("visit.vitals", "Lịch sử sinh hiệu")}>
        <Text type="secondary">Chưa có dữ liệu sinh hiệu</Text>
      </Card>
    );
  }

  return (
    <Card title={t("visit.vitals", "Lịch sử sinh hiệu")}>
      <Timeline>
        {vitals.map(vital => (
          <Timeline.Item key={vital.id}>
            <div className="font-semibold mb-1">
              {(() => {
                try {
                  return new Intl.DateTimeFormat('vi-VN', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit', year: 'numeric' }).format(new Date(vital.recorded_at));
                } catch {
                  return vital.recorded_at;
                }
              })()}
            </div>
            <div className="grid grid-cols-2 gap-2 text-sm">
              {vital.bp_systolic && vital.bp_diastolic && (
                <div>HA: <Text strong>{vital.bp_systolic}/{vital.bp_diastolic}</Text> mmHg</div>
              )}
              {vital.heart_rate && (
                <div>Mạch: <Text strong>{vital.heart_rate}</Text> bpm</div>
              )}
              {vital.temperature && (
                <div>Nhiệt độ: <Text strong>{vital.temperature}</Text> °C</div>
              )}
              {vital.spo2 && (
                <div>SpO2: <Text strong>{vital.spo2}</Text>%</div>
              )}
              {vital.weight_kg && (
                <div>Cân nặng: <Text strong>{vital.weight_kg}</Text> kg</div>
              )}
              {vital.height_cm && (
                <div>Chiều cao: <Text strong>{vital.height_cm}</Text> cm</div>
              )}
            </div>
          </Timeline.Item>
        ))}
      </Timeline>
    </Card>
  );
};
