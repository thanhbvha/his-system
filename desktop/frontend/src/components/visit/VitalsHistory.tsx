import React from 'react';
import { Timeline, Typography, Card, Tag } from 'antd';
import { useTranslation } from "react-i18next";
import { VisitVital } from '@/store/visitStore';
import dayjs from 'dayjs';

const { Text } = Typography;

interface VitalsHistoryProps {
  vitals: VisitVital[];
}

export const VitalsHistory: React.FC<VitalsHistoryProps> = ({ vitals }) => {
  const { t } = useTranslation();

  if (!vitals || vitals.length === 0) {
    return <Card bordered={false}><Text type="secondary">{t("common.none", "Không")}</Text></Card>;
  }

  const checkAbnormal = (field: string, value: number | undefined) => {
    if (value === undefined || value === null) return false;
    switch (field) {
      case 'bp_systolic': return value > 140 || value < 90;
      case 'bp_diastolic': return value > 90 || value < 60;
      case 'heart_rate': return value > 100 || value < 60;
      case 'temperature': return value > 37.5;
      case 'spo2': return value < 95;
      default: return false;
    }
  };

  const renderValue = (field: string, value: number | undefined, unit: string) => {
    if (value === undefined || value === null) return null;
    const isAbnormal = checkAbnormal(field, value);
    return (
      <span style={{ marginRight: 16 }}>
        <Text type="secondary">{t(`visit.${field}`, field)}: </Text>
        <Text type={isAbnormal ? 'danger' : undefined} strong={isAbnormal}>
          {value} {unit}
        </Text>
        {isAbnormal && <Tag color="error" style={{ marginLeft: 4 }}>{t("visit.abnormal", "Bất thường")}</Tag>}
      </span>
    );
  };

  return (
    <Card title={t("visit.history", "Lịch sử khám")} bordered={false}>
      <Timeline>
        {vitals.map(v => (
          <Timeline.Item key={v.id}>
            <div style={{ marginBottom: 4 }}>
              <Text strong>{dayjs(v.recorded_at).format('DD/MM/YYYY HH:mm')}</Text>
            </div>
            <div>
              {v.bp_systolic && v.bp_diastolic && (
                <span style={{ marginRight: 16 }}>
                  <Text type="secondary">{t("visit.bp", "Huyết áp")}: </Text>
                  <Text 
                    type={checkAbnormal('bp_systolic', v.bp_systolic) || checkAbnormal('bp_diastolic', v.bp_diastolic) ? 'danger' : undefined}
                  >
                    {v.bp_systolic}/{v.bp_diastolic} mmHg
                  </Text>
                </span>
              )}
              {renderValue('heart_rate', v.heart_rate, 'bpm')}
              {renderValue('temperature', v.temperature, '°C')}
              {renderValue('spo2', v.spo2, '%')}
              {renderValue('weight_kg', v.weight_kg, 'kg')}
              {renderValue('height_cm', v.height_cm, 'cm')}
            </div>
          </Timeline.Item>
        ))}
      </Timeline>
    </Card>
  );
};
