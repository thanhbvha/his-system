-- visits table
CREATE TABLE IF NOT EXISTS visits (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id       UUID NOT NULL REFERENCES patients(id),
    doctor_id        UUID NOT NULL REFERENCES users(id),
    queue_entry_id   UUID REFERENCES queue_entries(id),
    status           VARCHAR(20) NOT NULL DEFAULT 'REGISTERED',
    chief_complaint  TEXT,
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_visits_worklist ON visits(doctor_id, status, created_at);

-- visit_vitals table
CREATE TABLE IF NOT EXISTS visit_vitals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id      UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    bp_systolic   INT,
    bp_diastolic  INT,
    heart_rate    INT,
    temperature   DECIMAL(4,1),
    spo2          INT,
    weight_kg     DECIMAL(5,2),
    height_cm     DECIMAL(5,1),
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    recorded_by   UUID NOT NULL REFERENCES users(id)
);

-- visit_orders table
CREATE TABLE IF NOT EXISTS visit_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id    UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    order_type  VARCHAR(20) NOT NULL,
    ref_id      UUID,
    details     TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- icd10_codes table
CREATE TABLE IF NOT EXISTS icd10_codes (
    code            VARCHAR(10) PRIMARY KEY,
    description_vi  TEXT NOT NULL,
    category        VARCHAR(100),
    description_tsv TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('simple', code || ' ' || description_vi)
    ) STORED
);
CREATE INDEX IF NOT EXISTS idx_icd10_fts ON icd10_codes USING GIN(description_tsv);

-- Seed ICD-10 sample data
INSERT INTO icd10_codes (code, description_vi, category) VALUES
    ('I10',   'Tăng huyết áp nguyên phát',                           'Bệnh tim mạch'),
    ('I11',   'Tăng huyết áp kèm bệnh tim',                          'Bệnh tim mạch'),
    ('I21',   'Nhồi máu cơ tim cấp',                                  'Bệnh tim mạch'),
    ('I25',   'Bệnh tim thiếu máu cục bộ mạn tính',                  'Bệnh tim mạch'),
    ('I50',   'Suy tim',                                              'Bệnh tim mạch'),
    ('E10',   'Đái tháo đường type 1',                               'Rối loạn nội tiết'),
    ('E11',   'Đái tháo đường type 2',                               'Rối loạn nội tiết'),
    ('E78',   'Rối loạn chuyển hoá lipoprotein và các rối loạn lipid khác', 'Rối loạn nội tiết'),
    ('J06',   'Nhiễm khuẩn hô hấp trên cấp tính không đặc hiệu',    'Bệnh hô hấp'),
    ('J18',   'Viêm phổi không đặc hiệu',                            'Bệnh hô hấp'),
    ('J44',   'Bệnh phổi tắc nghẽn mạn tính khác',                  'Bệnh hô hấp'),
    ('K21',   'Bệnh trào ngược dạ dày thực quản',                    'Bệnh tiêu hoá'),
    ('K25',   'Loét dạ dày',                                         'Bệnh tiêu hoá'),
    ('K59',   'Các rối loạn chức năng đường ruột khác',              'Bệnh tiêu hoá'),
    ('M54',   'Đau lưng',                                            'Bệnh cơ xương'),
    ('M79',   'Các rối loạn mô mềm khác không phân loại nơi khác',  'Bệnh cơ xương'),
    ('N39',   'Các rối loạn khác của hệ tiết niệu',                  'Bệnh tiết niệu'),
    ('R05',   'Ho',                                                   'Triệu chứng và dấu hiệu'),
    ('R10',   'Đau bụng và vùng chậu',                               'Triệu chứng và dấu hiệu'),
    ('R51',   'Đau đầu',                                             'Triệu chứng và dấu hiệu'),
    ('R50',   'Sốt không rõ nguyên nhân',                            'Triệu chứng và dấu hiệu'),
    ('Z00',   'Khám sức khoẻ định kỳ',                               'Chăm sóc sức khoẻ dự phòng')
ON CONFLICT (code) DO NOTHING;
