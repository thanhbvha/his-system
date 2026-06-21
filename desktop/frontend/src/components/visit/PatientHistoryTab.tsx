import React from "react";
import { Spin, Timeline, Typography, Card, Tag } from "antd";
import { useTranslation } from "react-i18next";
import { usePatientStore } from "@/store/patientStore";

const { Text } = Typography;

interface PatientHistoryTabProps {
  patientId: string;
}

export const PatientHistoryTab: React.FC<PatientHistoryTabProps> = ({ patientId }) => {
  const { t } = useTranslation();
  const [history, setHistory] = React.useState<any[]>([]);
  const [loading, setLoading] = React.useState(true);
  const getPatientHistory = usePatientStore(state => state.getPatientHistory);

  React.useEffect(() => {
    let mounted = true;
    if (patientId) {
      setLoading(true);
      getPatientHistory(patientId).then((data) => {
        if (mounted) {
          setHistory(data);
          setLoading(false);
        }
      });
    }
    return () => { mounted = false; };
  }, [patientId, getPatientHistory]);

  if (loading) return <Spin className="w-full mt-4" />;
  if (!history.length) return <div className="text-gray-500 mt-4 text-center">{t("visit.noHistory", "Chưa có lịch sử khám bệnh")}</div>;

  return (
    <Card title={t("visit.history", "Lịch sử khám bệnh")}>
      <Timeline>
        {history.map((visit, index) => (
          <Timeline.Item key={visit.appointment_id || index} color="blue">
            <div className="font-semibold text-base mb-1">
              {(() => {
                try {
                  return new Intl.DateTimeFormat('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(visit.date));
                } catch {
                  return visit.date;
                }
              })()}
            </div>
            <div className="text-gray-700">
              <div>Bác sĩ: <Text strong>{visit.doctor_name || "Chưa có"}</Text></div>
              <div>Khoa/Phòng: <Tag color="geekblue">{visit.department_name || "Phòng khám"}</Tag></div>
              <div className="mt-1">Trạng thái: <Text className="text-blue-600">{visit.status}</Text></div>
            </div>
          </Timeline.Item>
        ))}
      </Timeline>
    </Card>
  );
};
