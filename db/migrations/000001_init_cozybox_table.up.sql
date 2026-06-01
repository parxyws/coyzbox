
-- =============================================================================
-- IDENTITY LAYER
-- =============================================================================

CREATE TABLE users
(
    id                    VARCHAR(26)  NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    username              VARCHAR(100) NOT NULL,
    email                 VARCHAR(255) NOT NULL,
    password              TEXT         NOT NULL DEFAULT '',
    -- Empty password = provisioned account, login blocked until setup is complete.
    -- Check: password = '' → return "account_setup_pending" error on login.

    status                VARCHAR(30)  NOT NULL DEFAULT 'active',
    -- 'active' | 'suspended' | in_active

    force_password_change BOOLEAN      NOT NULL DEFAULT FALSE,
    -- TRUE on owner-provisioned accounts until member sets their own password.
    -- Triggers a restricted JWT scope ("setup_only") on login until cleared.

    is_verified           BOOLEAN      NOT NULL DEFAULT FALSE,
    -- FALSE until email OTP is confirmed. Self-registered users start FALSE.
    -- Owner-provisioned users start TRUE (owner vouches for the email).

    onboarding_completed  BOOLEAN      NOT NULL DEFAULT FALSE,

    last_login            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT ck_users_status CHECK (status IN ('active', 'suspended', 'in_active'))
);

CREATE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users (status) WHERE deleted_at IS NULL;