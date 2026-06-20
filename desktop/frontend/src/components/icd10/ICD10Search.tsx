import React, { useState, useMemo, useRef } from 'react';
import { Select } from 'antd';
import { useTranslation } from "react-i18next";
import apiClient from '@/lib/apiClient';

function debounce<T extends (...args: any[]) => any>(func: T, wait: number) {
  let timeout: ReturnType<typeof setTimeout>;
  return function(this: any, ...args: Parameters<T>) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
}

interface ICD10SearchProps {
  onSelect: (code: string, description: string) => void;
  style?: React.CSSProperties;
}

export const ICD10Search: React.FC<ICD10SearchProps> = ({ onSelect, style }) => {
  const { t } = useTranslation();
  const [options, setOptions] = useState<any[]>([]);
  const [fetching, setFetching] = useState(false);

  const fetchOptions = useMemo(
    () => debounce(async (query: string) => {
      if (!query || query.length < 2) {
        setOptions([]);
        return;
      }
      setFetching(true);
      try {
        const res = await apiClient.get(`/icd10/search?q=${encodeURIComponent(query)}`);
        const data = res.data.data || [];
        setOptions(data.map((item: any) => ({
          label: `${item.code} — ${item.description_vi}`,
          value: item.code,
          itemData: item
        })));
      } catch (error) {
        console.error("Failed to search ICD10:", error);
      } finally {
        setFetching(false);
      }
    }, 500), // debounce 500ms
    []
  );

  const handleSelect = (value: string, option: any) => {
    onSelect(value, option.itemData.description_vi);
  };

  return (
    <Select
      showSearch
      placeholder={t("visit.icd10Search", "Tìm mã ICD-10")}
      filterOption={false}
      onSearch={fetchOptions}
      onSelect={handleSelect}
      options={options}
      loading={fetching}
      style={style || { width: '100%' }}
      allowClear
    />
  );
};
