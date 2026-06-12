CREATE TABLE IF NOT EXISTS patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    dob DATE,
    gender VARCHAR(20),
    phone_encrypted VARCHAR(255),
    phone_hmac VARCHAR(255),
    cccd_encrypted VARCHAR(255),
    cccd_hmac VARCHAR(255),
    email_encrypted VARCHAR(255),
    email_hmac VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patients_phone_hmac ON patients(phone_hmac);
CREATE INDEX IF NOT EXISTS idx_patients_cccd_hmac ON patients(cccd_hmac);

CREATE TABLE IF NOT EXISTS patient_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    relationship VARCHAR(100),
    phone VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS patient_insurance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE,
    insurance_code VARCHAR(100) NOT NULL,
    valid_from DATE,
    valid_to DATE
);
