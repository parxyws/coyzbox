
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
    status                VARCHAR(30)  NOT NULL DEFAULT 'active',
    force_password_change BOOLEAN      NOT NULL DEFAULT FALSE,
    is_verified           BOOLEAN      NOT NULL DEFAULT FALSE,
    onboarding_completed  BOOLEAN      NOT NULL DEFAULT FALSE,
    last_login            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT ck_users_status CHECK (status IN ('active', 'suspended', 'inactive'))
);

CREATE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users (status) WHERE deleted_at IS NULL;

-- =============================================================================
-- TENANT & ORGANIZATION LAYER
-- =============================================================================

CREATE TABLE tenants
(
    id         VARCHAR(26)  NOT NULL,
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(255) NOT NULL,
    status     VARCHAR(30)  NOT NULL DEFAULT 'active',
    type       VARCHAR(30)  NOT NULL DEFAULT 'personal',
    owner_id   VARCHAR(26)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_tenants PRIMARY KEY (id),
    CONSTRAINT uq_tenants_slug UNIQUE (slug),
    CONSTRAINT fk_tenants_owner FOREIGN KEY (owner_id) REFERENCES users (id),
    CONSTRAINT ck_tenants_status CHECK (status IN ('active', 'suspended', 'inactive')),
    CONSTRAINT ck_tenants_type CHECK (type IN ('personal', 'team'))
);

CREATE INDEX idx_tenants_slug ON tenants (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_owner ON tenants (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON tenants (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_type ON tenants (type) WHERE deleted_at IS NULL;

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

-- =============================================================================
-- DOCUMENT LAYER
-- =============================================================================

CREATE TABLE contacts
(
    id              VARCHAR(26)  NOT NULL,
    tenant_id       VARCHAR(26)  NOT NULL,
    organization_id VARCHAR(26)  NOT NULL,
    type            VARCHAR(30)  NOT NULL,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL DEFAULT '',
    phone           VARCHAR(50)  NOT NULL DEFAULT '',
    address_line1   TEXT         NOT NULL DEFAULT '',
    address_line2   TEXT         NOT NULL DEFAULT '',
    city            VARCHAR(100) NOT NULL DEFAULT '',
    state           VARCHAR(100) NOT NULL DEFAULT '',
    postal_code     VARCHAR(20)  NOT NULL DEFAULT '',
    country         VARCHAR(100) NOT NULL DEFAULT '',
    tax_id          VARCHAR(100) NOT NULL DEFAULT '',
    notes           TEXT         NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT pk_contacts PRIMARY KEY (id),
    CONSTRAINT fk_contact_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT fk_contact_org FOREIGN KEY (organization_id) REFERENCES organizations (id),
    CONSTRAINT ck_contact_type CHECK (type IN ('client', 'supplier', 'dual'))
);

CREATE INDEX idx_contacts_tenant ON contacts (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_contacts_org ON contacts (organization_id) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE documents
(
    id              VARCHAR(26)     NOT NULL,
    tenant_id       VARCHAR(26)     NOT NULL,
    organization_id VARCHAR(26)     NOT NULL,
    contact_id      VARCHAR(26),
    parent_id       VARCHAR(26),

    -- Classification
    type            VARCHAR(30)     NOT NULL,
    state           VARCHAR(30)     NOT NULL DEFAULT 'draft',
    flow_status     VARCHAR(30),
    document_ref    VARCHAR(100)    NOT NULL DEFAULT '',

    -- Dates
    issue_date      TIMESTAMPTZ,
    due_date        TIMESTAMPTZ,
    valid_until     TIMESTAMPTZ,

    -- Financial
    currency        VARCHAR(3)      NOT NULL DEFAULT 'USD',
    subtotal        NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    discount_amount NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    tax_amount      NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    total           NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    amount_paid     NUMERIC(15, 4)  NOT NULL DEFAULT 0,

    -- Content
    notes           TEXT            NOT NULL DEFAULT '',
    terms           TEXT            NOT NULL DEFAULT '',
    footer          TEXT            NOT NULL DEFAULT '',
    metadata        JSONB           NOT NULL DEFAULT '{}',

    -- File storage
    pdf_s3_key      TEXT            NOT NULL DEFAULT '',
    pdf_generated_at TIMESTAMPTZ,
    pdf_size_bytes  INTEGER         NOT NULL DEFAULT 0,

    -- Audit
    created_by      VARCHAR(26)     NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT pk_documents PRIMARY KEY (id),
    CONSTRAINT fk_doc_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT fk_doc_org FOREIGN KEY (organization_id) REFERENCES organizations (id),
    CONSTRAINT fk_doc_contact FOREIGN KEY (contact_id) REFERENCES contacts (id),
    CONSTRAINT fk_doc_parent FOREIGN KEY (parent_id) REFERENCES documents (id),
    CONSTRAINT fk_doc_creator FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT ck_doc_type CHECK (type IN ('quotation', 'invoice', 'receipt', 'purchase_order', 'sales_order', 'debit_note')),
    CONSTRAINT ck_doc_state CHECK (state IN ('draft', 'published', 'cancelled', 'expired'))
);

CREATE INDEX idx_docs_tenant ON documents (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_org ON documents (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_contact ON documents (contact_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_parent ON documents (parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_type ON documents (type) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_state ON documents (state) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_ref ON documents (document_ref) WHERE deleted_at IS NULL;
CREATE INDEX idx_docs_creator ON documents (created_by) WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------------------

CREATE TABLE document_items
(
    id           VARCHAR(26)     NOT NULL,
    document_id  VARCHAR(26)     NOT NULL,
    sort_order   INTEGER         NOT NULL DEFAULT 0,
    description  TEXT            NOT NULL DEFAULT '',
    quantity     NUMERIC(15, 4)  NOT NULL DEFAULT 1,
    unit         VARCHAR(50)     NOT NULL DEFAULT '',
    unit_price   NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    discount_pct NUMERIC(5, 2)   NOT NULL DEFAULT 0,
    tax_pct      NUMERIC(5, 2)   NOT NULL DEFAULT 0,
    amount       NUMERIC(15, 4)  NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_items PRIMARY KEY (id),
    CONSTRAINT fk_di_document FOREIGN KEY (document_id) REFERENCES documents (id) ON DELETE CASCADE
);

CREATE INDEX idx_di_document ON document_items (document_id);

-- -----------------------------------------------------------------------------

CREATE TABLE document_activities
(
    id           VARCHAR(26)  NOT NULL,
    document_id  VARCHAR(26)  NOT NULL,
    action       VARCHAR(50)  NOT NULL,
    from_status  VARCHAR(30),
    to_status    VARCHAR(30),
    performed_by VARCHAR(26),
    note         TEXT         NOT NULL DEFAULT '',
    metadata     JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_activities PRIMARY KEY (id),
    CONSTRAINT fk_da_document FOREIGN KEY (document_id) REFERENCES documents (id) ON DELETE CASCADE,
    CONSTRAINT fk_da_user FOREIGN KEY (performed_by) REFERENCES users (id)
);

CREATE INDEX idx_da_document ON document_activities (document_id);

-- -----------------------------------------------------------------------------

CREATE TABLE document_sequences
(
    id              VARCHAR(26)  NOT NULL,
    tenant_id       VARCHAR(26)  NOT NULL,
    organization_id VARCHAR(26)  NOT NULL,
    type            VARCHAR(30)  NOT NULL,
    prefix          VARCHAR(20)  NOT NULL DEFAULT '',
    next_number     INTEGER      NOT NULL DEFAULT 1,
    format          VARCHAR(100) NOT NULL DEFAULT '{PREFIX}-{YEAR}-{SEQ:4}',
    last_reset_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_document_sequences PRIMARY KEY (id),
    CONSTRAINT fk_ds_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT fk_ds_org FOREIGN KEY (organization_id) REFERENCES organizations (id),
    CONSTRAINT ck_ds_type CHECK (type IN ('quotation', 'invoice', 'receipt', 'purchase_order', 'sales_order', 'debit_note')),
    CONSTRAINT uq_ds_org_type UNIQUE (organization_id, type)
);

CREATE INDEX idx_ds_tenant ON document_sequences (tenant_id);
CREATE INDEX idx_ds_org ON document_sequences (organization_id);

-- =============================================================================
-- TEMPLATE LAYER
-- =============================================================================

CREATE TABLE template_configs
(
    id         VARCHAR(26)  NOT NULL,
    tenant_id  VARCHAR(26)  NOT NULL,
    base_type  VARCHAR(30)  NOT NULL,
    status     VARCHAR(30)  NOT NULL DEFAULT 'draft',
    name       VARCHAR(255) NOT NULL DEFAULT '',
    config     JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_template_configs PRIMARY KEY (id),
    CONSTRAINT fk_tc_tenant FOREIGN KEY (tenant_id) REFERENCES tenants (id),
    CONSTRAINT ck_tc_base_type CHECK (base_type IN ('quotation', 'invoice', 'receipt', 'purchase_order', 'sales_order', 'debit_note'))
);

CREATE INDEX idx_tc_tenant ON template_configs (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tc_type ON template_configs (tenant_id, base_type) WHERE deleted_at IS NULL;
