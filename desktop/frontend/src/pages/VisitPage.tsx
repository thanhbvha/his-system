import React from 'react';
import { useParams } from 'react-router-dom';
import { DoctorWorklist } from '@/components/visit/DoctorWorklist';
import { VisitScreen } from '@/components/visit/VisitScreen';

export const VisitPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  if (id) {
    return <VisitScreen visitId={id} />;
  }

  return <DoctorWorklist />;
};
