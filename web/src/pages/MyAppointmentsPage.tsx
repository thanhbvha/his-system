import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface Appointment {
  id: string;
  status: string;
  scheduled_at: string;
  note: string;
  doctor: { full_name: string };
  service: { name: string };
}

export const MyAppointmentsPage = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'upcoming' | 'past'>('upcoming');
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchAppointments = async () => {
    setIsLoading(true);
    try {
      const res = await apiClient.get(`/appointments?patient_id=me&status=${activeTab}`);
      setAppointments(res.data.data.items || []);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAppointments();
  }, [activeTab]);

  const handleCancel = async (id: string) => {
    if (!window.confirm("Bạn có chắc chắn muốn hủy lịch hẹn này?")) return;
    try {
      await apiClient.put(`/appointments/${id}/status`, { status: 'CANCELLED' });
      fetchAppointments();
    } catch (error) {
      alert("Hủy lịch thất bại");
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'PENDING': return 'bg-yellow-100 text-yellow-800';
      case 'CONFIRMED': return 'bg-blue-100 text-blue-800';
      case 'CHECKED_IN': return 'bg-green-100 text-green-800';
      case 'COMPLETED': return 'bg-gray-100 text-gray-800';
      case 'CANCELLED': return 'bg-red-100 text-red-800';
      default: return 'bg-slate-100 text-slate-800';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 min-h-[400px]">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">{t("nav.myAppointments")}</h2>
      </div>

      <div className="flex gap-4 border-b mb-6">
        <button 
          className={`pb-2 px-1 ${activeTab === 'upcoming' ? 'border-b-2 border-primary font-semibold text-primary' : 'text-slate-500'}`}
          onClick={() => setActiveTab('upcoming')}
        >
          Sắp tới
        </button>
        <button 
          className={`pb-2 px-1 ${activeTab === 'past' ? 'border-b-2 border-primary font-semibold text-primary' : 'text-slate-500'}`}
          onClick={() => setActiveTab('past')}
        >
          Lịch sử khám
        </button>
      </div>

      {isLoading ? (
        <p className="text-center text-slate-500 my-8">Đang tải...</p>
      ) : appointments.length === 0 ? (
        <p className="text-center text-slate-500 my-8">Không có lịch hẹn nào.</p>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {appointments.map(app => (
            <Card key={app.id}>
              <CardHeader className="pb-2 flex flex-row items-start justify-between space-y-0">
                <div>
                  <CardTitle className="text-lg">{app.service?.name}</CardTitle>
                  <CardDescription className="font-semibold text-slate-900 mt-1">{app.scheduled_at}</CardDescription>
                </div>
                <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(app.status)}`}>
                  {app.status}
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-sm space-y-2 text-slate-600 mt-2">
                  <p><strong>Bác sĩ:</strong> {app.doctor?.full_name}</p>
                  <p><strong>Ghi chú:</strong> {app.note || 'Không có'}</p>
                </div>
                
                {app.status === 'PENDING' || app.status === 'CONFIRMED' ? (
                  <div className="mt-4 pt-4 border-t flex justify-end">
                    <Button variant="destructive" size="sm" onClick={() => handleCancel(app.id)}>
                      Hủy lịch
                    </Button>
                  </div>
                ) : null}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};
