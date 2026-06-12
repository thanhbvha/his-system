import { useAuthStore } from "@/store/authStore";

export const AccountPage = () => {
  const patient = useAuthStore((s) => s.patient);

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 min-h-[400px]">
      <h2 className="text-2xl font-bold mb-4">Patient Profile</h2>
      <div className="space-y-2">
        <p><strong>ID:</strong> {patient?.id}</p>
        <p><strong>Name:</strong> {patient?.name}</p>
        <p><strong>Email:</strong> {patient?.email}</p>
      </div>
    </div>
  );
};
