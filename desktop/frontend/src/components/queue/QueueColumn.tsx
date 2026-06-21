import React from "react";
import { QueueEntry } from "@/store/queueStore";
import { Button, Card, Badge, Typography, Popconfirm } from "antd";
import { useTranslation } from "react-i18next";
import { BellOutlined, FastForwardOutlined } from "@ant-design/icons";

const { Text } = Typography;

interface QueueColumnProps {
  serviceType: string;
  entries: QueueEntry[];
  onCallNext: (id: string) => void;
  onSkip: (id: string) => void;
  readOnly?: boolean;
}

export const QueueColumn: React.FC<QueueColumnProps> = ({ serviceType, entries, onCallNext, onSkip, readOnly = false }) => {
  const { t } = useTranslation();

  const calledEntry = entries.find(e => e.status === "CALLED" || e.status === "IN_PROGRESS");
  const waitingEntries = entries.filter(e => e.status === "WAITING").sort((a, b) => a.queue_number.localeCompare(b.queue_number));
  
  const nextWaiting = waitingEntries.length > 0 ? waitingEntries[0] : null;

  return (
    <Card 
      title={<span className="text-lg font-bold text-gray-800">{serviceType}</span>} 
      className="shadow-md h-full flex flex-col"
      bodyStyle={{ padding: '16px', flex: 1, display: 'flex', flexDirection: 'column' }}
    >
      {/* Curently Called Section */}
      <div className="mb-6 p-4 rounded-lg bg-blue-50 border border-blue-200 text-center relative overflow-hidden">
        <Text type="secondary" className="block mb-2 uppercase text-xs font-semibold tracking-wider">
          {t("queue.called")}
        </Text>
        
        {calledEntry ? (
          <div className="animate-pulse-slow">
            <div className="text-3xl font-black text-blue-600 mb-2">{calledEntry.queue_number}</div>
            <div className="text-lg font-medium text-gray-700">{calledEntry.patient.full_name}</div>
            <div className="mt-4 flex justify-center gap-2">
              {!readOnly && (
                <Popconfirm
                  title="Bạn có chắc chắn muốn bỏ qua bệnh nhân này?"
                  description="Hành động này sẽ đẩy bệnh nhân vào cuối hàng đợi hoặc xóa khỏi hàng đợi hiện tại."
                  onConfirm={() => onSkip(calledEntry.id)}
                  okText="Đồng ý"
                  cancelText="Hủy"
                >
                  <Button size="small" icon={<FastForwardOutlined />}>
                    {t("queue.skip")}
                  </Button>
                </Popconfirm>
              )}
            </div>
          </div>
        ) : (
          <div className="py-6 text-gray-400 italic">
            {t("queue.noQueue")}
          </div>
        )}
      </div>

      {/* Action Button */}
      {!readOnly && (
        <Popconfirm
          title="Gọi bệnh nhân tiếp theo?"
          description="Hành động này sẽ cập nhật trạng thái bệnh nhân thành 'Đang gọi'."
          onConfirm={() => nextWaiting ? onCallNext(nextWaiting.id) : null}
          okText="Đồng ý"
          cancelText="Hủy"
          disabled={!nextWaiting && !calledEntry}
        >
          <Button 
            type="primary" 
            size="large" 
            block 
            disabled={!nextWaiting && !calledEntry}
            className="mb-6 h-12 text-base font-medium shadow-sm"
            icon={<BellOutlined />}
          >
            {t("queue.callNext")} {nextWaiting ? `(${nextWaiting.queue_number})` : ''}
          </Button>
        </Popconfirm>
      )}

      {/* Waiting List */}
      <div className="flex-1 overflow-y-auto">
        <div className="flex justify-between items-center mb-3 pb-2 border-b border-gray-100">
          <Text className="font-semibold text-gray-600">{t("queue.waiting")}</Text>
          <Badge count={waitingEntries.length} style={{ backgroundColor: '#52c41a' }} />
        </div>
        
        <div className="space-y-2">
          {waitingEntries.map(entry => (
            <div key={entry.id} className="flex justify-between items-center p-3 rounded-md bg-gray-50 hover:bg-gray-100 transition-colors">
              <div className="flex items-center gap-3">
                <span className="inline-flex items-center justify-center px-3 py-1 rounded-md bg-blue-50 shadow-sm border border-blue-200 font-bold text-blue-700 text-sm">
                  {entry.queue_number}
                </span>
                <span className="font-medium text-gray-800 line-clamp-1">{entry.patient.full_name}</span>
              </div>
              <span className="text-xs text-gray-400">
                {new Date(entry.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
            </div>
          ))}
          {waitingEntries.length === 0 && (
            <div className="text-center py-4 text-gray-400 text-sm">
              {t("queue.noQueue")}
            </div>
          )}
        </div>
      </div>
    </Card>
  );
};
