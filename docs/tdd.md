# Technical Design Document: CozyBox Architecture Updates

This document details the engineering specifications for the two primary architectural updates: the Token-Based Sequence Generator and the Client-Side PDF Rendering Pipeline.

## 1. Token-Based Sequence Generator

Allowing custom document numbers introduces severe concurrency risks. To mitigate duplicate numbers and database fragmentation, sequence generation is decoupled from the string representation using a lexical parser and isolated state tables.

### 1.1 Database Schema
We isolate the configuration of the pattern from the mathematical state of the sequence.

-- Stores the user's custom format preference
CREATE TABLE tenant_sequence_settings (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
document_type VARCHAR(30) NOT NULL,
pattern_template VARCHAR(64) NOT NULL, -- e.g., "{{COMPANY_ID}}-{{YYYY}}-{{SEQ:5}}"
reset_cycle VARCHAR(20) NOT NULL, -- 'monthly', 'yearly', 'never'
UNIQUE(tenant_id, document_type)
);

-- Tracks the actual numeric increment strictly by period
CREATE TABLE tenant_sequence_state (
tenant_id UUID NOT NULL,
document_type VARCHAR(30) NOT NULL,
current_period VARCHAR(7) NOT NULL, -- '2026-06', '2026', or 'global' based on reset_cycle
last_value INT NOT NULL DEFAULT 0,
PRIMARY KEY (tenant_id, document_type, current_period)
);

### 1.2 Supported Lexical Tokens
The Go backend will parse the `pattern_template` string and replace the following exact token matches:
* {{YYYY}} / {{YY}}: Current Year (UTC or Tenant Timezone).
* {{MM}}: Current Month (01-12).
* {{COMPANY_ID}}: The unique internal identifier for the tenant's company.
* {{TEXT:value}}: Static text provided by the user (must be sanitized for URL-safe characters).
* {{SEQ:X}}: The auto-incrementing integer. `X` defines the zero-padding length (1-9).

### 1.3 The Transactional Boundary (Concurrency Control)
To guarantee sequence integrity when multiple users click "Publish" simultaneously, the sequence generation must occur within a strict Postgres transaction using pessimistic locking.

Go Backend Execution Flow:
1.  Begin Transaction (tx).
2.  Determine Period: Based on the tenant's `reset_cycle`, calculate the `current_period` string (e.g., "2026-06").
3.  Acquire Lock:
    SELECT last_value FROM tenant_sequence_state
    WHERE tenant_id = $1 AND document_type = $2 AND current_period = $3
    FOR UPDATE; -- Blocks concurrent requests until tx commits
4.  Increment: If no row exists, `last_value = 0`. Calculate `next_value = last_value + 1`. Update or insert the state row.
5.  Compile String: Pass the `pattern_template`, `next_value`, and current dates through the Go tokenizer.
    * Input: INV-{{YYYY}}-{{MM}}-{{SEQ:4}}
    * Output: INV-2026-06-0001
6.  Insert Document: Insert the main document record using the compiled string as the `document_number`.
7.  Commit Transaction. (Releases the lock.)

---

## 2. API Contract

All protected routes require a valid `Authorization: Bearer <access_token>` header. The `tenant_id` is extracted from the JWT claims — it is **never** accepted as a request body field. All routes are prefixed with `/api`.

> **Legend**
> - `public` — no token required
> - `auth` — valid JWT required (tenant context injected from token)

---

### 2.1 Auth

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `POST` | `/auth/register` | public | Create user account + personal tenant. Sends OTP email. Rate-limited (3/60s per IP). |
| `POST` | `/auth/verify-email` | public | Confirm email using OTP + reference ID. |
| `POST` | `/auth/login` | public | Authenticate user. Returns `access_token`, `refresh_token`, workspace list. Rate-limited (5/60s per IP). |
| `POST` | `/auth/refresh-token` | public | Exchange refresh token + session ID for a new token pair. |
| `POST` | `/auth/forgot-password` | public | Request OTP to reset password. Always returns 200 (prevents email enumeration). Rate-limited (3/60s per IP). |
| `POST` | `/auth/reset-password` | public | Reset password using OTP + reference ID. |
| `POST` | `/auth/logout` | auth | Invalidate the current session in Redis. |

---

### 2.2 Workspaces

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/workspaces` | auth | List all tenants the authenticated user is a member of. |
| `POST` | `/workspaces/switch` | auth | Re-issue JWT scoped to a different tenant. Returns new `access_token`. |

---

### 2.3 User (Profile)

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/user/profile` | auth | Get the authenticated user's profile. |
| `PUT` | `/user/profile` | auth | Update name and/or username. |
| `PUT` | `/user/password` | auth | Change password (requires current password verification). |

---

### 2.4 Onboarding

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `POST` | `/onboarding` | auth | Complete organization profile setup for the active tenant. Accepts `multipart/form-data` (fields + optional logo file). Sets `onboarding_completed = true` on the user. Idempotent — can be re-submitted if skipped. |

> **Note:** The dashboard wizard enforces onboarding completion before allowing document creation. This check is handled client-side by reading `onboarding_completed` from the login/profile response, not a separate API gate.

---

### 2.5 Tenant Settings

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/tenant/organization` | auth | Get the active tenant's organization profile. |
| `PUT` | `/tenant/organization` | auth | Update organization fields and/or logo. Accepts `multipart/form-data`. |
| `GET` | `/tenant/template-configs` | auth | List all 6 template configs for the active tenant (uses system defaults if tenant hasn't customized). |
| `PUT` | `/tenant/template-configs/:id` | auth | Update a specific template config (brand color, watermark, header layout). Only `published` configs are used at generation time. |

---

### 2.6 Contacts

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/contacts` | auth | List contacts for the active tenant. Supports query params: `?type=client\|supplier\|dual`, `?search=<name>`, `?page=`, `?limit=`. |
| `POST` | `/contacts` | auth | Create a new contact scoped to the active tenant. |
| `GET` | `/contacts/:id` | auth | Get a single contact by ID (must belong to active tenant). |
| `PUT` | `/contacts/:id` | auth | Update a contact. |
| `DELETE` | `/contacts/:id` | auth | Soft-delete a contact. Contacts referenced by published documents are not hard-deleted. |

---

### 2.7 Documents

#### 2.7.1 Core CRUD

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/documents` | auth | List documents for the active tenant. Supports: `?type=`, `?status=`, `?search=<ref\|contact>`, `?page=`, `?limit=`, `?sort=created_at\|due_date`. |
| `POST` | `/documents` | auth | **Step 1 of the publish flow.** Commit document data (line items, contact, amounts) as a frozen record. Executes the sequence generation transaction. Returns `document_id` and `document_number`. |
| `GET` | `/documents/:id` | auth | Get a single document with items and activity timeline. |
| `PUT` | `/documents/:id` | auth | Update document fields. Only allowed when `status = draft`. Returns `400` if document is not in draft. |
| `DELETE` | `/documents/:id` | auth | Soft-delete a document. Only allowed when `status = draft` or terminal. |

#### 2.7.2 PDF Artifact

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `POST` | `/documents/:id/upload` | auth | **Step 2 of the publish flow.** Accept the client-generated PDF blob. Validates MIME type (`application/pdf`) and size (`< 5MB`). Streams to MinIO under `/{tenant_id}/{doc_type}/{doc_id}.pdf`. Updates `pdf_s3_key`, `pdf_size_bytes`, `pdf_generated_at`. |
| `GET` | `/documents/:id/download` | auth | Return a pre-signed MinIO download URL (short TTL, e.g. 15 min). Does not stream the file through the Go server. |

#### 2.7.3 Status Transitions

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `POST` | `/documents/:id/publish` | auth | Transition `draft → published`. Assigns `document_ref`. Triggers PDF upload flow on the client side. Logs activity. |
| `POST` | `/documents/:id/status` | auth | Manual status transition (e.g., `published → accepted`, `accepted → paid`). Body: `{ "status": "<target>", "note": "" }`. Validated against `AllowedTransitions()`. Logs activity. |
| `POST` | `/documents/:id/payment` | auth | Record a payment. Body: `{ "amount": <decimal>, "note": "" }`. Automatically transitions to `paid` if `amount_paid >= total`, else `partially_paid`. Logs activity. |

---

### 2.8 Document Sequence Settings

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `GET` | `/tenant/sequences` | auth | List sequence settings for all 6 document types for the active tenant. |
| `PUT` | `/tenant/sequences/:type` | auth | Create or update the sequence pattern for a specific document type (e.g., `invoice`). Body: `{ "pattern_template": "INV-{{YYYY}}-{{MM}}-{{SEQ:4}}", "reset_cycle": "monthly" }`. Validates token syntax. `type` must be one of the 6 valid document types. |

---

### 2.9 Internal / Cron

| Method | Path | Auth | Description |
| :----- | :--- | :--: | :---------- |
| `POST` | `/internal/cron/document-transitions` | internal | Triggered by the cron worker. Runs batch status transitions: `published → expired` (past `valid_until`) and `published\|accepted → overdue` (past `due_date`). Responds with a count of affected documents. Not exposed to the public router. |

---

### 2.10 Route Summary

| Domain | Routes | Status |
| :----- | :----: | :----- |
| Auth | 7 | Done |
| Workspaces | 2 | Done |
| User | 3 | Done |
| Onboarding | 1 | Pending |
| Tenant Settings | 4 | Done |
| Contacts | 5 | Pending |
| Documents (CRUD) | 5 | Pending |
| Documents (PDF) | 2 | Pending |
| Documents (Status) | 3 | Pending |
| Sequence Settings | 2 | Pending |
| Internal / Cron | 1 | Pending |
| **Total** | **35** | |

---

## 3. Client-Side PDF Rendering Pipeline

Moving PDF rendering to the client offloads compute from the Go server but requires a strict verification flow to ensure the database remains the ultimate source of truth.

### 3.1 The @react-pdf/renderer Implementation
We will use @react-pdf/renderer instead of standard HTML-to-Canvas libraries. It uses the Yoga layout engine, which interprets a React-like DOM and renders it directly to a PDF binary, bypassing browser-specific CSS inconsistencies.

* Font Embedding: The frontend bundle must dynamically load required standard fonts (e.g., Roboto) via URL before rendering to ensure standard financial formatting.
* Background Generation: The generation process happens in-memory and does not require the PDF to be visible on the DOM.

### 3.2 The Two-Step "Publish" Flow
Because the client generates the file, but the server must securely store it, the "Generate PDF" action is a synchronized, two-step pipeline.

Step 1: Data Commitment (The Source of Truth)
1.  The React app sends a POST /api/documents request with the JSON payload containing line items, totals, and contact IDs.
2.  The Go backend executes the Sequence Generation transaction (Section 1.3).
3.  The backend saves the data as a frozen JSONB snapshot in the database.
4.  The backend responds with 201 Created and returns the `document_id` and the generated `document_number`.

Step 2: Artifact Generation & Upload
1.  The React app receives the `document_number` and immediately injects it into the @react-pdf/renderer document state.
2.  React-PDF generates the PDF Blob in the browser memory.
3.  The React app automatically triggers a POST /api/documents/:id/upload request containing the binary Blob.
4.  The Go backend streams this Blob directly to MinIO and updates the document record with the `s3_key`.
5.  The React app provides the user with the download link/button.

### 3.3 Security & Integrity Rules
* The Database is King: The uploaded PDF is treated as a visual artifact, not the data truth. If a malicious user intercepts Step 2 and uploads a tampered PDF, the financial math in the PostgreSQL JSONB snapshot remains unaltered. Any audit or API read will rely on the database, not OCR on the PDF.
* File Validation: The /upload endpoint must strictly validate the MIME type (application/pdf) and file size limit (e.g., < 5MB) before streaming to MinIO to prevent arbitrary file upload attacks.
* Storage Prefixing: All MinIO objects must be stored using a strict tenant-isolated path structure: /{tenant_id}/{document_type}/{document_id}.pdf.