-- =============================================================================
-- ROLLBACK — Drop in reverse dependency order
-- =============================================================================

-- TEMPLATE LAYER
DROP INDEX IF EXISTS idx_tc_type;
DROP INDEX IF EXISTS idx_tc_tenant;
DROP TABLE IF EXISTS template_configs;

-- DOCUMENT LAYER (leaf tables first)
DROP INDEX IF EXISTS idx_ds_org;
DROP INDEX IF EXISTS idx_ds_tenant;
DROP TABLE IF EXISTS document_sequences;

DROP INDEX IF EXISTS idx_da_document;
DROP TABLE IF EXISTS document_activities;

DROP INDEX IF EXISTS idx_di_document;
DROP TABLE IF EXISTS document_items;

DROP INDEX IF EXISTS idx_docs_creator;
DROP INDEX IF EXISTS idx_docs_ref;
DROP INDEX IF EXISTS idx_docs_state;
DROP INDEX IF EXISTS idx_docs_type;
DROP INDEX IF EXISTS idx_docs_parent;
DROP INDEX IF EXISTS idx_docs_contact;
DROP INDEX IF EXISTS idx_docs_org;
DROP INDEX IF EXISTS idx_docs_tenant;
DROP TABLE IF EXISTS documents;

DROP INDEX IF EXISTS idx_contacts_org;
DROP INDEX IF EXISTS idx_contacts_tenant;
DROP TABLE IF EXISTS contacts;

-- TENANT & ORGANIZATION LAYER
DROP INDEX IF EXISTS idx_org_tenant;
DROP TABLE IF EXISTS organizations;

DROP INDEX IF EXISTS idx_tm_user;
DROP INDEX IF EXISTS idx_tm_tenant;
DROP TABLE IF EXISTS tenant_members;

DROP INDEX IF EXISTS idx_tenants_type;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_owner;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP TABLE IF EXISTS tenants;

-- IDENTITY LAYER
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
