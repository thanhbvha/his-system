DROP TABLE IF EXISTS device_registry CASCADE;

CREATE TABLE device_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,
    public_key_pem TEXT NOT NULL,
    public_key_hash VARCHAR(255) NOT NULL,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, device_fingerprint)
);
