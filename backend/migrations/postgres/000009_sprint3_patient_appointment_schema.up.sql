-- ============================================================
-- Sprint 3: Patient & Appointment — Full Schema
-- Migration: 000009_sprint3_patient_appointment_schema.up.sql
-- ============================================================

-- ============================================================
-- 1. PATIENTS (Expand existing stub from migration 000002)
-- ============================================================
DROP TABLE IF EXISTS patient_contacts CASCADE;
DROP TABLE IF EXISTS patient_insurance CASCADE;
DROP TABLE IF EXISTS patients CASCADE;

CREATE TABLE patients (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Cleartext (non-sensitive)
    full_name                 VARCHAR(255) NOT NULL,
    full_name_search          TSVECTOR,           -- auto-update via trigger
    dob                       DATE,
    gender                    VARCHAR(10),        -- 'MALE' | 'FEMALE' | 'OTHER'
    blood_type                VARCHAR(5),         -- 'A+' | 'B-' | 'AB+' | 'O' ...
    is_active                 BOOLEAN NOT NULL DEFAULT true,

    -- Phone (PII encrypted)
    phone_encrypted           TEXT,
    phone_hmac                VARCHAR(64),        -- SHA-256 hex, indexed for exact-match

    -- CCCD / CMND (PII encrypted)
    cccd_encrypted            TEXT,
    cccd_hmac                 VARCHAR(64),

    -- Email (PII encrypted, optional)
    email_encrypted           TEXT,
    email_hmac                VARCHAR(64),

    -- Address (PII encrypted, optional)
    address_detail_encrypted  TEXT,

    -- Avatar
    avatar_url                TEXT,

    -- Audit
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Exact-match search on HMAC fields (B-tree index)
CREATE UNIQUE INDEX idx_patients_phone_hmac  ON patients(phone_hmac)  WHERE phone_hmac  IS NOT NULL;
CREATE UNIQUE INDEX idx_patients_cccd_hmac   ON patients(cccd_hmac)   WHERE cccd_hmac   IS NOT NULL;
CREATE INDEX        idx_patients_email_hmac  ON patients(email_hmac)  WHERE email_hmac  IS NOT NULL;

-- Full-text search on full_name (GIN index)
CREATE INDEX idx_patients_full_name_search ON patients USING GIN(full_name_search);

-- Trigger: auto-update tsvector when full_name changes
CREATE OR REPLACE FUNCTION update_patient_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.full_name_search := to_tsvector('simple', coalesce(NEW.full_name, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_patients_search_vector
BEFORE INSERT OR UPDATE ON patients
FOR EACH ROW EXECUTE FUNCTION update_patient_search_vector();

-- ============================================================
-- 2. PATIENT INSURANCE (BHYT)
-- ============================================================
CREATE TABLE patient_insurance (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id            UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,

    -- BHYT (PII encrypted)
    bhyt_number_encrypted TEXT,
    bhyt_hmac             VARCHAR(64),

    valid_from            DATE,
    valid_to              DATE,
    coverage_level        VARCHAR(50),    -- e.g. '80%' | '100%'
    issuing_province      VARCHAR(100),

    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patient_insurance_patient_id ON patient_insurance(patient_id);
CREATE INDEX idx_patient_insurance_bhyt_hmac  ON patient_insurance(bhyt_hmac) WHERE bhyt_hmac IS NOT NULL;

-- ============================================================
-- 3. PATIENT CONTACTS (Emergency contact, người thân)
-- ============================================================
CREATE TABLE patient_contacts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id       UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    relationship     VARCHAR(100),           -- 'FATHER' | 'MOTHER' | 'SPOUSE' | 'OTHER'
    phone_encrypted  TEXT,
    phone_hmac       VARCHAR(64),
    is_primary       BOOLEAN NOT NULL DEFAULT false,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patient_contacts_patient_id ON patient_contacts(patient_id);

-- ============================================================
-- 4. MEDICAL SERVICES (Danh mục dịch vụ/chuyên khoa)
-- ============================================================
CREATE TABLE services (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(50) NOT NULL UNIQUE,
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    price        NUMERIC(12, 2) NOT NULL DEFAULT 0,
    duration_min INT NOT NULL DEFAULT 30,   -- default slot duration in minutes
    is_active    BOOLEAN NOT NULL DEFAULT true,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 5. DOCTOR PROFILES (extend users table with medical info)
-- ============================================================
ALTER TABLE staff_profiles
    ADD COLUMN IF NOT EXISTS specialty    VARCHAR(255),
    ADD COLUMN IF NOT EXISTS avatar_url   TEXT,
    ADD COLUMN IF NOT EXISTS bio          TEXT;

-- ============================================================
-- 6. DOCTOR SCHEDULES (Weekly recurring schedule template)
-- ============================================================
DROP TABLE IF EXISTS appointment_slots CASCADE;
DROP TABLE IF EXISTS slot_templates CASCADE;
DROP TABLE IF EXISTS doctor_schedules CASCADE;

CREATE TABLE doctor_schedules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department_id    UUID REFERENCES departments(id) ON DELETE SET NULL,
    day_of_week      SMALLINT NOT NULL,     -- 0=Mon, 1=Tue, ... 6=Sun
    start_time       TIME NOT NULL,
    end_time         TIME NOT NULL,
    slot_duration_min INT NOT NULL DEFAULT 30,
    is_active        BOOLEAN NOT NULL DEFAULT true,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (doctor_id, day_of_week, start_time)
);

CREATE INDEX idx_doctor_schedules_doctor_id ON doctor_schedules(doctor_id);

-- ============================================================
-- 7. APPOINTMENT SLOTS (Concrete time slots, generated from schedule)
-- ============================================================
CREATE TABLE appointment_slots (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id    UUID REFERENCES doctor_schedules(id) ON DELETE SET NULL,
    slot_date      DATE NOT NULL,
    start_time     TIME NOT NULL,
    end_time       TIME NOT NULL,
    is_booked      BOOLEAN NOT NULL DEFAULT false,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (doctor_id, slot_date, start_time)   -- prevent duplicate slot generation
);

CREATE INDEX idx_appointment_slots_doctor_date ON appointment_slots(doctor_id, slot_date);
CREATE INDEX idx_appointment_slots_available   ON appointment_slots(doctor_id, slot_date) WHERE is_booked = false;

-- ============================================================
-- 8. APPOINTMENTS
-- ============================================================
DROP TABLE IF EXISTS appointments CASCADE;

CREATE TABLE appointments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id    UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    service_id    UUID REFERENCES services(id) ON DELETE SET NULL,
    slot_id       UUID REFERENCES appointment_slots(id) ON DELETE SET NULL,

    -- Denormalized for quick display without joins
    scheduled_at  TIMESTAMPTZ NOT NULL,

    -- State machine: PENDING → CONFIRMED → CHECKED_IN → COMPLETED | CANCELLED
    status        VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                  CHECK (status IN ('PENDING','CONFIRMED','CHECKED_IN','COMPLETED','CANCELLED')),

    note          TEXT,
    cancel_reason TEXT,

    -- Booking metadata
    booked_by     UUID REFERENCES users(id) ON DELETE SET NULL,  -- staff who booked (NULL if self-booked)
    booked_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at  TIMESTAMPTZ,
    checked_in_at TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    cancelled_at  TIMESTAMPTZ,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appointments_patient_id    ON appointments(patient_id);
CREATE INDEX idx_appointments_doctor_date   ON appointments(doctor_id, scheduled_at);
CREATE INDEX idx_appointments_status        ON appointments(status);
CREATE INDEX idx_appointments_scheduled_at  ON appointments(scheduled_at);
