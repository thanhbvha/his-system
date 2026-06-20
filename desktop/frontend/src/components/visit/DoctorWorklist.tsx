import React, { useEffect } from 'react';
import { Table, Button, Tag, Space, Typography, Card } from 'antd';
import { useTranslation } from "react-i18next";
import { useVisitStore, Visit } from '@/store/visitStore';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';

const { Title } = Typography;

export const DoctorWorklist: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { worklist, fetchWorklist, isLoading } = useVisitStore();

  useEffect(() => {
    fetchWorklist();

    // Listen for queue updates to refresh worklist
    const handleQueueUpdated = () => fetchWorklist();
    window.addEventListener('queue.updated', handleQueueUpdated);

    return () => {
      window.removeEventListener('queue.updated', handleQueueUpdated);
    };
  }, [fetchWorklist]);

  const columns = [
    {
      title: 'TT',
      dataIndex: 'id',
      key: 'id',
      render: (text: string, record: Visit, index: number) => index + 1,
      width: 60,
    },
    {
      title: t("patients.fullName", "Họ và tên"),
      dataIndex: ['patient', 'full_name'],
      key: 'patient_name',
    },
    {
      title: t("patients.dob", "Ngày sinh"),
      dataIndex: ['patient', 'dob'],
      key: 'dob',
    },
    {
      title: t("common.status", "Trạng thái"),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        let color = 'default';
        if (status === 'WAITING') color = 'warning';
        else if (status === 'IN_PROGRESS') color = 'processing';
        else if (status === 'COMPLETED') color = 'success';
        return <Tag color={color}>{status}</Tag>;
      }
    },
    {
      title: t("admin.userList.actions", "Thao tác"),
      key: 'action',
      render: (_: any, record: Visit) => (
        <Space size="middle">
          <Button 
            type="primary" 
            onClick={() => navigate(`/visits/${record.id}`)}
          >
            {record.status === 'WAITING' ? t("visit.startVisit", "Bắt đầu khám") : t("common.select", "Chọn")}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Card title={<Title level={4} style={{ margin: 0 }}>{t("visit.worklist", "Danh sách bệnh nhân")}</Title>}>
      <Table 
        columns={columns} 
        dataSource={worklist} 
        rowKey="id" 
        loading={isLoading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  );
};
