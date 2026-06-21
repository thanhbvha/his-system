import React, { useEffect } from "react";
import { Table, Button, Badge, Space, Typography } from "antd";
import { useTranslation } from "react-i18next";
import { useQueueStore, QueueEntry } from "@/store/queueStore";
import { useAuthStore } from "@/store/authStore";
import { useQueueWS } from "@/components/ws/useQueueWS";
import { useNavigate } from "react-router-dom";
import { useVisitStore } from "@/store/visitStore";

const { Title } = Typography;

export const DoctorWorklist: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { entries, isLoading, fetchQueue, callNext, skip } = useQueueStore();
  const { createVisit } = useVisitStore();
  const { token, user, roomId } = useAuthStore();

  useQueueWS(token);

  useEffect(() => {
    fetchQueue();
  }, [fetchQueue]);

  // Filter entries to only show those for the doctor's active service (room)
  // In a real app, this would filter by the doctor's assigned service type.
  // We'll use the roomId as the service_type if they are a doctor.
  const myQueue = ((roomId === "global_reception" || !roomId) ? entries : entries.filter(e => e.service_type === roomId))
    .filter(e => e.status !== "DONE");

  const handleCallAndVisit = async (entry: QueueEntry) => {
    try {
      // If the queue entry already has a visit_id, navigate directly to it
      if (entry.visit_id) {
        navigate(`/visits/${entry.visit_id}`);
        return;
      }
      
      // If waiting, call it first
      if (entry.status === 'WAITING' || entry.status === 'SKIPPED') {
        await callNext(entry.id);
      }
      
      const visit = await createVisit({
        patient_id: entry.patient.id,
        doctor_id: user!.id,
        queue_entry_id: entry.id,
      });
      
      navigate(`/visits/${visit.id}`);
    } catch (error) {
      console.error("Failed to start visit:", error);
    }
  };

  const columns = [
    {
      title: 'Số TT',
      dataIndex: 'queue_number',
      key: 'queue_number',
      render: (text: string) => <strong className="text-lg text-blue-600">{text}</strong>,
    },
    {
      title: 'Tên bệnh nhân',
      key: 'patient',
      render: (_: any, record: QueueEntry) => (
        <div>
          <div className="font-semibold">{record.patient?.full_name}</div>
          <div className="text-gray-500 text-xs">Mã BN: {record.patient?.patient_code}</div>
        </div>
      ),
    },
    {
      title: 'Giờ đăng ký',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => {
        try {
          return new Intl.DateTimeFormat('vi-VN', { hour: '2-digit', minute: '2-digit' }).format(new Date(date));
        } catch (e) {
          return date;
        }
      },
    },
    {
      title: 'Trạng thái',
      key: 'status',
      dataIndex: 'status',
      render: (status: string) => {
        switch (status) {
          case 'WAITING':
            return <Badge status="processing" text="Đang chờ" />;
          case 'CALLED':
            return <Badge status="warning" text="Đang gọi" />;
          case 'IN_PROGRESS':
            return <Badge status="success" text="Đang khám" />;
          case 'SKIPPED':
            return <Badge status="default" text="Bỏ qua" />;
          case 'DONE':
            return <Badge status="success" text="Hoàn tất" />;
          default:
            return <Badge status="default" text={status} />;
        }
      },
    },
    {
      title: 'Thao tác',
      key: 'action',
      render: (_: any, record: QueueEntry) => (
        <Space size="middle">
          {(record.status === 'WAITING' || record.status === 'CALLED' || record.status === 'SKIPPED') && (
            <Button 
              type="primary" 
              onClick={() => handleCallAndVisit(record)}
            >
              Gọi vào khám
            </Button>
          )}
          {record.status === 'IN_PROGRESS' && (
            <Button 
              onClick={() => handleCallAndVisit(record)}
            >
              Tiếp tục khám
            </Button>
          )}
          {record.status === 'WAITING' && (
            <Button onClick={() => skip(record.id)}>Bỏ qua</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="p-6 bg-white rounded-lg shadow-sm">
      <div className="flex justify-between items-center mb-6">
        <Title level={3} style={{ margin: 0 }}>{t("visit.worklist", "Danh sách bệnh nhân")}</Title>
        <Button onClick={() => fetchQueue()}>Làm mới</Button>
      </div>
      
      <Table 
        columns={columns} 
        dataSource={myQueue} 
        rowKey="id" 
        loading={isLoading}
        pagination={{ pageSize: 10 }}
      />
    </div>
  );
};
