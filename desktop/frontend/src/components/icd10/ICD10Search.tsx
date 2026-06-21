import React, { useState, useMemo } from "react";
import { AutoComplete, Spin } from "antd";
import { useTranslation } from "react-i18next";
import apiClient from "@/lib/apiClient";

interface ICD10SearchProps {
  onSelect: (code: string) => void;
}

export const ICD10Search: React.FC<ICD10SearchProps> = ({ onSelect }) => {
  const { t } = useTranslation();
  const [options, setOptions] = useState<{ label: string; value: string }[]>([]);
  const [fetching, setFetching] = useState(false);

  const timeoutRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleSearch = (query: string) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(async () => {
      if (!query || query.length < 2) {
        setOptions([]);
        return;
      }
      setFetching(true);
      try {
        const res = await apiClient.get(`/icd10/search?q=${query}`);
        const data = res.data.data || [];
        
        setOptions(data.map((item: any) => ({
          label: `${item.code} — ${item.description_vi}`,
          value: item.code,
        })));
        setFetching(false);
      } catch (error) {
        console.error("ICD-10 search error", error);
        setFetching(false);
      }
    }, 500);
  };

  return (
    <AutoComplete
      style={{ width: '100%' }}
      options={options}
      onSearch={handleSearch}
      onSelect={onSelect}
      placeholder={t("visit.icd10Search", "Nhập mã hoặc tên bệnh (VD: Tim mạch)")}
      notFoundContent={fetching ? <Spin size="small" /> : null}
    />
  );
};
