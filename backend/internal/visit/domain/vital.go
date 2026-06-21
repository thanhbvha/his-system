package domain

import (
	"time"

	"github.com/google/uuid"
)

type VisitVital struct {
	ID          uuid.UUID   `json:"id"`
	VisitID     uuid.UUID   `json:"visit_id"`
	BpSystolic  *int        `json:"bp_systolic,omitempty"`
	BpDiastolic *int        `json:"bp_diastolic,omitempty"`
	HeartRate   *int        `json:"heart_rate,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
	SpO2        *int        `json:"spo2,omitempty"`
	WeightKg    *float64    `json:"weight_kg,omitempty"`
	HeightCm    *float64    `json:"height_cm,omitempty"`
	RecordedAt  time.Time   `json:"recorded_at"`
	RecordedBy  uuid.UUID   `json:"recorded_by"`
}

// Alerts returns a list of clinical alert messages for abnormal values.
func (v *VisitVital) Alerts() []string {
	var alerts []string
	if v.BpSystolic != nil && *v.BpSystolic >= 140 {
		alerts = append(alerts, "Huyết áp tâm thu cao (≥140 mmHg)")
	}
	if v.BpDiastolic != nil && *v.BpDiastolic >= 90 {
		alerts = append(alerts, "Huyết áp tâm trương cao (≥90 mmHg)")
	}
	if v.HeartRate != nil && (*v.HeartRate > 100 || *v.HeartRate < 60) {
		alerts = append(alerts, "Mạch bất thường (<60 hoặc >100 bpm)")
	}
	if v.Temperature != nil && (*v.Temperature >= 38.0 || *v.Temperature < 35.5) {
		alerts = append(alerts, "Nhiệt độ bất thường (<35.5°C hoặc ≥38.0°C)")
	}
	if v.SpO2 != nil && *v.SpO2 < 95 {
		alerts = append(alerts, "SpO2 thấp (<95%)")
	}
	return alerts
}
