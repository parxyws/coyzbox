-- =============================================================================
-- TENANT & ORGANIZATION LAYER
-- =============================================================================

CREATE TABLE tenants
(
    id         VARCHAR(26)  NOT NULL,
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(255) NOT NULL,
    status     VARCHAR(30)  NOT NULL DEFAULT 'active',
    owner_id   VARCHAR(26)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_tenants PRIMARY KEY (id),
    CONSTRAINT uq_tenants_slug UNIQUE (slug),
    CONSTRAINT fk_tenants_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT ck_tenants_status CHECK (status IN ('active', 'suspended', 'inactive'))
);

CREATE INDEX idx_tenants_slug ON tenants (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_owner ON tenants (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON tenants (status) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE tenant_members
(
    id        VARCHAR(26) NOT NULL,
    tenant_id VARCHAR(26) NOT NULL,
    user_id   VARCHAR(26) NOT NULL,
    role      VARCHAR(30) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_tenant_members PRIMARY KEY (id),
    CONSTRAINT fk_tm_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT fk_tm_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT uq_tenant_user UNIQUE (tenant_id, user_id),
    CONSTRAINT ck_tm_role CHECK (role IN ('owner', 'member'))
);

CREATE INDEX idx_tm_tenant ON tenant_members (tenant_id);
CREATE INDEX idx_tm_user ON tenant_members (user_id);

-- -----------------------------------------------------------------------------

CREATE TABLE organizations
(
    id               VARCHAR(26)  NOT NULL,
    tenant_id        VARCHAR(26)  NOT NULL,
    name             VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL DEFAULT '',
    phone            VARCHAR(50)  NOT NULL DEFAULT '',
    address_line1    TEXT         NOT NULL DEFAULT '',
    address_line2    TEXT         NOT NULL DEFAULT '',
    city             VARCHAR(100) NOT NULL DEFAULT '',
    state            VARCHAR(100) NOT NULL DEFAULT '',
    postal_code      VARCHAR(20)  NOT NULL DEFAULT '',
    country          VARCHAR(100) NOT NULL DEFAULT '',
    tax_id           VARCHAR(100) NOT NULL DEFAULT '',
    logo_s3_key      TEXT         NOT NULL DEFAULT '',
    website          VARCHAR(255) NOT NULL DEFAULT '',
    timezone         VARCHAR(50)  NOT NULL DEFAULT 'UTC',
    default_currency VARCHAR(3)   NOT NULL DEFAULT 'USD',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,

    CONSTRAINT pk_organizations PRIMARY KEY (id),
    CONSTRAINT fk_org_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT uq_org_tenant UNIQUE (tenant_id)
);

CREATE INDEX idx_org_tenant ON organizations (tenant_id) WHERE deleted_at IS NULL;
