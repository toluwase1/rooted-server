-- Admin authentication table
CREATE TABLE admin_users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    name            VARCHAR(100) NOT NULL,
    role            VARCHAR(20) DEFAULT 'admin',  -- admin, super_admin
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

-- Seed default admin account
-- Password: Toluwase1994 (bcrypt hash)
-- Change this after first login!
INSERT INTO admin_users (email, password_hash, name, role) VALUES
('toluwasethomas1@gmail.com', '$2a$10$.Vf6g9mnOZBOwHge.lB8COecVzdgU9fgK09bGI3vu2ICf1RoPuADm', 'Toluwase', 'super_admin');

-- Request logs table for structured API logging
CREATE TABLE request_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    method          VARCHAR(10) NOT NULL,
    path            VARCHAR(500) NOT NULL,
    status_code     INT NOT NULL,
    latency_ms      INT NOT NULL,
    user_id         UUID,
    telegram_id     BIGINT,
    error_message   TEXT,
    ip              VARCHAR(45),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_request_logs_created ON request_logs(created_at DESC);
CREATE INDEX idx_request_logs_status ON request_logs(status_code) WHERE status_code >= 400;
