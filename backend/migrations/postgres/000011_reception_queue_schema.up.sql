CREATE TABLE IF NOT EXISTS queue_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id),
    visit_id        UUID,
    appointment_id  UUID REFERENCES appointments(id),
    service_type    VARCHAR(50) NOT NULL DEFAULT 'GENERAL',
    queue_number    VARCHAR(10) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'WAITING',
    called_at       TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_queue_today ON queue_entries(((created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date), service_type, status);
