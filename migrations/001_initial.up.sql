-- Rooted Dating - Initial Schema
-- This migration creates all tables for the MVP

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- ============================================
-- USERS & PROFILES
-- ============================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id     BIGINT UNIQUE NOT NULL,
    telegram_username VARCHAR(64),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    last_active_at  TIMESTAMPTZ DEFAULT NOW(),
    status          VARCHAR(20) DEFAULT 'active',
    verification    VARCHAR(20) DEFAULT 'unverified',
    trust_score     SMALLINT DEFAULT 50,
    subscription    VARCHAR(20) DEFAULT 'free',
    sub_expires_at  TIMESTAMPTZ
);

CREATE TABLE profiles (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name      VARCHAR(50) NOT NULL,
    date_of_birth   DATE NOT NULL,
    gender          VARCHAR(20) NOT NULL,
    gender_pref     VARCHAR(20) NOT NULL,
    city            VARCHAR(100),
    country         VARCHAR(3),
    latitude        DECIMAL(10, 7),
    longitude       DECIMAL(10, 7),
    location        GEOGRAPHY(POINT, 4326),
    heritage        VARCHAR(50)[] DEFAULT '{}',
    diaspora_tag    VARCHAR(30),
    intention       VARCHAR(20),
    faith           VARCHAR(30),
    faith_importance VARCHAR(20),
    bio             VARCHAR(150),
    audio_bio_url   TEXT,
    cultural_prompts JSONB DEFAULT '[]',
    personality_prompts JSONB DEFAULT '[]',
    completeness    SMALLINT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profiles_location ON profiles USING GIST(location);
CREATE INDEX idx_profiles_heritage ON profiles USING GIN(heritage);
CREATE INDEX idx_profiles_diaspora ON profiles(diaspora_tag);
CREATE INDEX idx_profiles_intention ON profiles(intention);
CREATE INDEX idx_profiles_faith ON profiles(faith);
CREATE INDEX idx_profiles_country ON profiles(country);
CREATE INDEX idx_profiles_gender_pref ON profiles(gender, gender_pref);
CREATE INDEX idx_profiles_matching ON profiles(gender, intention, faith_importance)
    WHERE completeness >= 50;

CREATE TABLE photos (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_thumbnail   TEXT NOT NULL,
    url_medium      TEXT NOT NULL,
    url_large       TEXT NOT NULL,
    position        SMALLINT NOT NULL,
    is_primary      BOOLEAN DEFAULT FALSE,
    moderation_status VARCHAR(20) DEFAULT 'pending',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_photos_user ON photos(user_id, position);

-- ============================================
-- MATCHING & SWIPES
-- ============================================

CREATE TABLE swipes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    swiper_id       UUID NOT NULL REFERENCES users(id),
    swiped_id       UUID NOT NULL REFERENCES users(id),
    action          VARCHAR(10) NOT NULL,
    comment         TEXT,
    liked_element   VARCHAR(50),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(swiper_id, swiped_id)
);

CREATE INDEX idx_swipes_swiped ON swipes(swiped_id, action) WHERE action = 'like';
CREATE INDEX idx_swipes_swiper ON swipes(swiper_id, created_at DESC);

CREATE TABLE matches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_a_id       UUID NOT NULL REFERENCES users(id),
    user_b_id       UUID NOT NULL REFERENCES users(id),
    compatibility   SMALLINT,
    status          VARCHAR(20) DEFAULT 'active',
    matched_at      TIMESTAMPTZ DEFAULT NOW(),
    last_message_at TIMESTAMPTZ,
    UNIQUE(user_a_id, user_b_id)
);

CREATE INDEX idx_matches_user_a ON matches(user_a_id, status) WHERE status = 'active';
CREATE INDEX idx_matches_user_b ON matches(user_b_id, status) WHERE status = 'active';

CREATE TABLE seen_profiles (
    user_id         UUID NOT NULL REFERENCES users(id),
    seen_user_id    UUID NOT NULL REFERENCES users(id),
    seen_at         TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, seen_user_id)
);

-- ============================================
-- CHAT (Bot-Proxied)
-- ============================================

CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id        UUID NOT NULL REFERENCES matches(id),
    user_a_id       UUID NOT NULL,
    user_b_id       UUID NOT NULL,
    status          VARCHAR(20) DEFAULT 'active',
    last_message_at TIMESTAMPTZ,
    message_count   INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ
);

CREATE INDEX idx_conversations_user_a ON conversations(user_a_id);
CREATE INDEX idx_conversations_user_b ON conversations(user_b_id);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    sender_id       UUID NOT NULL REFERENCES users(id),
    content_type    VARCHAR(20) NOT NULL,
    content         TEXT,
    moderation_status VARCHAR(20) DEFAULT 'clean',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);

CREATE TABLE telegram_chat_routing (
    telegram_chat_id BIGINT PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    active_conversation_id UUID REFERENCES conversations(id)
);

CREATE INDEX idx_routing_user ON telegram_chat_routing(user_id);

-- ============================================
-- PAYMENTS
-- ============================================

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    plan            VARCHAR(20) NOT NULL,
    star_amount     INT NOT NULL,
    telegram_payment_charge_id TEXT,
    status          VARCHAR(20) DEFAULT 'active',
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    type            VARCHAR(30) NOT NULL,
    star_amount     INT,
    fiat_amount     DECIMAL(10, 2),
    fiat_currency   VARCHAR(3),
    provider        VARCHAR(20),
    provider_ref    TEXT,
    status          VARCHAR(20) DEFAULT 'completed',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_user ON transactions(user_id, created_at DESC);

-- ============================================
-- COMMUNITIES
-- ============================================

CREATE TABLE communities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    type            VARCHAR(20),
    telegram_group_id BIGINT,
    member_count    INT DEFAULT 0,
    is_official     BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE community_members (
    community_id    UUID NOT NULL REFERENCES communities(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    role            VARCHAR(20) DEFAULT 'member',
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (community_id, user_id)
);

-- ============================================
-- MODERATION
-- ============================================

CREATE TABLE reports (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id     UUID NOT NULL REFERENCES users(id),
    reported_id     UUID NOT NULL REFERENCES users(id),
    category        VARCHAR(30) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) DEFAULT 'pending',
    reviewed_by     UUID,
    reviewed_at     TIMESTAMPTZ,
    action_taken    VARCHAR(30),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_reports_reported ON reports(reported_id, status);
CREATE INDEX idx_reports_pending ON reports(status) WHERE status = 'pending';

CREATE TABLE blocklist (
    user_id         UUID NOT NULL REFERENCES users(id),
    blocked_user_id UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, blocked_user_id)
);

-- ============================================
-- EVENTS
-- ============================================

CREATE TABLE events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    event_type      VARCHAR(20),
    format          VARCHAR(20),
    city            VARCHAR(100),
    country         VARCHAR(3),
    venue           TEXT,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    capacity        INT,
    ticket_price    DECIMAL(10, 2),
    ticket_currency VARCHAR(3),
    attendee_count  INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE event_attendees (
    event_id        UUID NOT NULL REFERENCES events(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    transaction_id  UUID REFERENCES transactions(id),
    status          VARCHAR(20) DEFAULT 'confirmed',
    PRIMARY KEY (event_id, user_id)
);

-- ============================================
-- DAILY MATCHES
-- ============================================

CREATE TABLE daily_circles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    candidate_ids   UUID[] NOT NULL,
    delivered_at    TIMESTAMPTZ DEFAULT NOW(),
    opened          BOOLEAN DEFAULT FALSE,
    actions_taken   INT DEFAULT 0
);

CREATE INDEX idx_daily_circles_user ON daily_circles(user_id, delivered_at DESC);

-- ============================================
-- ADMIN CONFIGURATION
-- ============================================

CREATE TABLE admin_config (
    key             VARCHAR(100) PRIMARY KEY,
    value           JSONB NOT NULL,
    category        VARCHAR(50) NOT NULL,
    description     TEXT,
    value_type      VARCHAR(20) NOT NULL,
    constraints     JSONB,
    updated_by      UUID REFERENCES users(id),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE admin_config_audit (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_key      VARCHAR(100) NOT NULL,
    old_value       JSONB,
    new_value       JSONB NOT NULL,
    changed_by      UUID NOT NULL,
    reason          TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_config_audit_key ON admin_config_audit(config_key, created_at DESC);

CREATE TABLE admin_config_regional (
    config_key      VARCHAR(100) NOT NULL REFERENCES admin_config(key),
    region          VARCHAR(10) NOT NULL,
    value           JSONB NOT NULL,
    updated_by      UUID REFERENCES users(id),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (config_key, region)
);

CREATE TABLE feature_flags (
    key             VARCHAR(100) PRIMARY KEY,
    enabled         BOOLEAN DEFAULT FALSE,
    description     TEXT,
    rollout_percent SMALLINT DEFAULT 0,
    target_regions  VARCHAR(10)[] DEFAULT '{}',
    target_plans    VARCHAR(20)[] DEFAULT '{}',
    updated_by      UUID REFERENCES users(id),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- SEED: Default Config Values
-- ============================================

INSERT INTO admin_config (key, value, category, description, value_type) VALUES
-- Matching
('daily_circle_size', '8', 'matching', 'Daily matches for free users', 'int'),
('daily_circle_size_plus', '15', 'matching', 'Daily matches for Plus/Premium', 'int'),
('explore_swipe_limit_free', '15', 'matching', 'Daily explore swipes for free users', 'int'),
('explore_swipe_limit_plus', '999', 'matching', 'Daily explore swipes for paid users', 'int'),
('min_profile_completeness', '50', 'matching', 'Min profile % to appear in matching', 'int'),
('match_weight_cultural', '0.45', 'matching', 'Cultural affinity weight (MVP)', 'float'),
('match_weight_values', '0.0', 'matching', 'Values alignment weight (0 until v1.1)', 'float'),
('match_weight_diaspora', '0.20', 'matching', 'Diaspora fit weight', 'float'),
('match_weight_proximity', '0.10', 'matching', 'Proximity weight', 'float'),
('match_weight_quality', '0.10', 'matching', 'Profile quality weight', 'float'),
('match_weight_recency', '0.10', 'matching', 'Recency weight', 'float'),
('heritage_exact_match_points', '15', 'matching', 'Points for shared heritage', 'int'),
('heritage_same_country_points', '7', 'matching', 'Points for same country', 'int'),
('heritage_same_region_points', '4', 'matching', 'Points for same region', 'int'),
('faith_both_important_points', '12', 'matching', 'Points for shared important faith', 'int'),
('faith_same_points', '8', 'matching', 'Points for same faith', 'int'),
('faith_both_unimportant_points', '3', 'matching', 'Points when neither cares about faith', 'int'),
('diaspora_same_tag_points', '10', 'matching', 'Points for same diaspora tag', 'int'),
('diaspora_similar_tag_points', '7', 'matching', 'Points for similar diaspora experience', 'int'),
('diaspora_cross_continental_points', '5', 'matching', 'Points for cross-continental + heritage', 'int'),
('proximity_10km_points', '10', 'matching', 'Points within 10km', 'int'),
('proximity_50km_points', '8', 'matching', 'Points within 50km', 'int'),
('proximity_200km_points', '5', 'matching', 'Points within 200km', 'int'),
('proximity_same_country_points', '3', 'matching', 'Points for same country', 'int'),
('new_user_boost_percent', '10', 'matching', 'New user visibility boost %', 'int'),
('new_user_boost_days', '7', 'matching', 'New user boost duration in days', 'int'),
('spotlight_boost_percent', '20', 'matching', 'Spotlight visibility boost %', 'int'),
('reciprocity_boost_percent', '15', 'matching', 'Boost when they liked you', 'int'),
('diversity_max_same_city', '3', 'matching', 'Max from same city in circle', 'int'),
('diversity_max_same_heritage', '2', 'matching', 'Max from same heritage in circle', 'int'),
('trust_score_penalty_threshold', '30', 'matching', 'Trust score below this gets penalized', 'int'),
('trust_score_penalty_factor', '0.5', 'matching', 'Penalty multiplier for low trust', 'float'),

-- Monetization
('plus_price_stars', '250', 'monetization', 'Rooted Plus monthly price in Stars', 'int'),
('premium_price_stars', '500', 'monetization', 'Rooted Premium monthly price in Stars', 'int'),
('rose_price_stars_1', '15', 'monetization', 'Single Rose price', 'int'),
('rose_price_stars_5', '60', 'monetization', '5-pack Rose price', 'int'),
('boost_price_stars', '75', 'monetization', 'Spotlight Boost price', 'int'),
('super_boost_price_stars', '200', 'monetization', 'Super Boost price', 'int'),
('reopen_chat_price_stars', '30', 'monetization', 'Reopen expired chat price', 'int'),
('free_roses_per_week_premium', '3', 'monetization', 'Roses included in Premium per week', 'int'),
('free_boosts_per_month_plus', '1', 'monetization', 'Boosts included in Plus per month', 'int'),
('free_boosts_per_month_premium', '3', 'monetization', 'Boosts included in Premium per month', 'int'),

-- Chat
('conversation_expiry_days', '14', 'chat', 'Days before inactive chat expires', 'int'),
('chat_nudge_day', '7', 'chat', 'Day to send nudge about fading chat', 'int'),
('voice_note_max_seconds_free', '60', 'chat', 'Max voice note length for free users', 'int'),
('voice_note_max_seconds_premium', '180', 'chat', 'Max voice note length for Premium', 'int'),
('max_active_conversations', '20', 'chat', 'Max simultaneous active chats', 'int'),
('message_rate_limit_per_minute', '30', 'chat', 'Max messages per minute per user', 'int'),

-- Moderation
('auto_ban_report_threshold', '3', 'moderation', 'Unique reports before auto-suspension', 'int'),
('moderation_sensitivity', '"medium"', 'moderation', 'AI moderation threshold', 'string'),
('report_sla_hours', '4', 'moderation', 'Target report response time', 'int'),

-- Verification
('face_match_confidence', '90', 'verification', 'Face verification confidence threshold', 'int'),
('verification_deadline_hours', '48', 'verification', 'Hours to complete verification', 'int'),
('unverified_score_penalty', '0.7', 'verification', 'Score multiplier for unverified users', 'float'),

-- Notifications
('circle_delivery_hour_utc', '18', 'notification', 'Hour to deliver daily circle (UTC)', 'int'),
('weekly_summary_day', '"sunday"', 'notification', 'Day for weekly summary', 'string'),
('weekly_summary_enabled', 'true', 'notification', 'Enable weekly summaries', 'bool'),
('notification_cooldown_minutes', '5', 'notification', 'Min gap between notifications', 'int'),

-- Communities
('free_communities_limit', '2', 'community', 'Max communities for free users', 'int'),
('community_auto_moderation', 'true', 'community', 'Enable AI moderation in groups', 'bool');

-- Seed feature flags (all disabled for MVP)
INSERT INTO feature_flags (key, enabled, description, rollout_percent) VALUES
('compatibility_questions', false, 'Enable compatibility questions (v1.1)', 0),
('audio_bio', false, 'Enable audio bio upload (v1.1)', 0),
('communities', false, 'Enable communities (v1.1)', 0),
('crossed_continents', false, 'Enable Crossed Continents (v1.2)', 0),
('events', false, 'Enable events marketplace (v1.2)', 0),
('referral_program', false, 'Enable referral rewards (v1.2)', 0),
('message_translation', false, 'Enable message translation (v1.2)', 0),
('virtual_dates', false, 'Enable virtual date mode (v1.2)', 0),
('video_chat', false, 'Enable video/voice calls (v1.2)', 0),
('incognito_mode', false, 'Enable incognito browsing (v1.2)', 0);
