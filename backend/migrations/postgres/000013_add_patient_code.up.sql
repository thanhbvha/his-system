CREATE SEQUENCE IF NOT EXISTS patient_code_seq START 1;

ALTER TABLE patients 
ADD COLUMN IF NOT EXISTS patient_code VARCHAR(20);

-- Backfill existing patients with codes starting from BN260000001
UPDATE patients 
SET patient_code = 'BN' || to_char(now() AT TIME ZONE 'Asia/Ho_Chi_Minh', 'YY') || to_char(nextval('patient_code_seq'), 'FM0000000')
WHERE patient_code IS NULL;

ALTER TABLE patients 
ALTER COLUMN patient_code SET NOT NULL,
ADD CONSTRAINT patients_patient_code_key UNIQUE (patient_code),
ALTER COLUMN patient_code SET DEFAULT 'BN' || to_char(now() AT TIME ZONE 'Asia/Ho_Chi_Minh', 'YY') || to_char(nextval('patient_code_seq'), 'FM0000000');
