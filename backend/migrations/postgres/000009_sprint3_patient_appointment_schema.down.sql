-- ============================================================
-- Sprint 3: Rollback
-- Migration: 000009_sprint3_patient_appointment_schema.down.sql
-- ============================================================

DROP TABLE IF EXISTS appointments CASCADE;
DROP TABLE IF EXISTS appointment_slots CASCADE;
DROP TABLE IF EXISTS doctor_schedules CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS patient_contacts CASCADE;
DROP TABLE IF EXISTS patient_insurance CASCADE;
DROP TABLE IF EXISTS patients CASCADE;

DROP FUNCTION IF EXISTS update_patient_search_vector CASCADE;

ALTER TABLE staff_profiles
    DROP COLUMN IF EXISTS specialty,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS bio;
