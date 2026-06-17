import { useState, useEffect } from "react";
import { Table, Checkbox, Button, message, Space } from "antd";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { SaveOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

export const RolePermissionPage = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const [matrixState, setMatrixState] = useState<Record<string, Record<string, boolean>>>({});
  const [dirtyRoles, setDirtyRoles] = useState<Set<string>>(new Set());

  // Fetch all predefined permissions
  const { data: permsData, isLoading: isLoadingPerms } = useQuery({
    queryKey: ["permissions"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/permissions");
      return res.data.data; // array of { id, resource, action }
    }
  });

  // Fetch roles (which include permissions)
  const { data: rolesData, isLoading: isLoadingRoles } = useQuery({
    queryKey: ["roles"],
    queryFn: async () => {
      const res = await apiClient.get("/admin/roles");
      return res.data.data;
    }
  });

  useEffect(() => {
    if (rolesData) {
      const newState: Record<string, Record<string, boolean>> = {};
      rolesData.forEach((role: any) => {
        newState[role.id] = {};
        role.permissions?.forEach((p: any) => {
          const key = `${p.resource}:${p.action}`;
          newState[role.id][key] = true;
        });
      });
      setMatrixState(newState);
      setDirtyRoles(new Set());
    }
  }, [rolesData]);

  const updatePermissionsMutation = useMutation({
    mutationFn: async ({ roleId, permissions }: { roleId: string, permissions: any[] }) => {
      await apiClient.put(`/admin/roles/${roleId}/permissions`, { permissions });
    },
    onSuccess: () => {
      message.success(t("admin.roles.saveSuccess"));
      queryClient.invalidateQueries({ queryKey: ["roles"] });
    },
    onError: () => {
      message.error(t("admin.roles.saveError"));
    }
  });

  const handleCheckboxChange = (roleId: string, permKey: string, checked: boolean) => {
    setMatrixState(prev => ({
      ...prev,
      [roleId]: {
        ...prev[roleId],
        [permKey]: checked
      }
    }));
    
    setDirtyRoles(prev => {
      const newSet = new Set(prev);
      newSet.add(roleId);
      return newSet;
    });
  };

  const handleSave = async () => {
    if (dirtyRoles.size === 0) {
      message.info(t("admin.roles.noChanges"));
      return;
    }

    try {
      const promises = Array.from(dirtyRoles).map(roleId => {
        const rolePermsMap = matrixState[roleId] || {};
        const permissionsToSave = Object.keys(rolePermsMap)
          .filter(key => rolePermsMap[key])
          .map(key => {
            const [resource, action] = key.split(":");
            return { resource, action };
          });
        
        return updatePermissionsMutation.mutateAsync({ roleId, permissions: permissionsToSave });
      });

      await Promise.all(promises);
      setDirtyRoles(new Set());
    } catch (err) {
      // errors handled in mutation onError
    }
  };

  // Build unique permissions from API
  const allPermKeys = new Set<string>();
  permsData?.forEach((p: any) => {
    allPermKeys.add(`${p.resource}:${p.action}`);
  });
  
  // Sort permission keys
  const sortedPermKeys = Array.from(allPermKeys).sort();

  const dataSource = sortedPermKeys.map(key => ({
    key,
    permission: key
  }));

  // Columns: Permission name first, then 1 column per role
  const columns: any[] = [
    {
      title: "Permission",
      dataIndex: "permission",
      key: "permission",
      fixed: "left",
      width: 200,
    }
  ];

  rolesData?.forEach((role: any) => {
    columns.push({
      title: role.name,
      dataIndex: role.id,
      key: role.id,
      align: "center",
      render: (_: any, record: any) => {
        const isChecked = matrixState[role.id]?.[record.key] || false;
        return (
          <Checkbox 
            checked={isChecked} 
            onChange={(e) => handleCheckboxChange(role.id, record.key, e.target.checked)} 
          />
        );
      }
    });
  });

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <h2>{t("admin.roles.title")}</h2>
        <Space>
          {dirtyRoles.size > 0 && <span style={{ color: "#faad14" }}>{t("admin.roles.unsavedChanges", { count: dirtyRoles.size })}</span>}
          <Button 
            type="primary" 
            icon={<SaveOutlined />} 
            onClick={handleSave} 
            loading={updatePermissionsMutation.isPending}
            disabled={dirtyRoles.size === 0}
          >
            {t("admin.roles.saveChanges")}
          </Button>
        </Space>
      </div>

      <Table 
        columns={columns} 
        dataSource={dataSource} 
        loading={isLoadingRoles || isLoadingPerms}
        pagination={false}
        scroll={{ x: 'max-content' }}
        bordered
      />
    </div>
  );
};
