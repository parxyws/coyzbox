# PRD: CozyBox Document Engine (v1.0)

## 1. Summary / Purpose
**CozyBox** is a **multi-tenant, desktop-optimized document engine** that eliminates "folder chaos" and "blank page syndrome" for SMEs, freelancers, and micro-landlords. It provides a unified service to generate compliant business documents (PDFs) and manage them in a centralized, searchable repository.

In v1.0, CozyBox focuses on **six standard financial document types**: Quotation, Invoice, Receipt, Purchase Order, Sales Order, and Debit Note. Each tenant operates in isolation — documents are never shared across tenants — while default templates are globally available and customizable per tenant (e.g., logo, brand color).

---

## 2. Target User Personas

| Persona | Pain Point | Need |
| :--- | :--- | :--- |
| **SME Owner** | High legal/admin costs; manual errors in contracts and invoices. | Professional templates with auto-fill capabilities. |
| **Freelancer** | Disorganized files across email and local drives. | A "Single Source of Truth" for all business docs. |
| **Property Admin** | Repetitive data entry for tenant renewals. | Hybrid data pulling (System data + Manual entry). |

---

## 3. User Stories

*   **US.1:** As a user, I want to pull data from my existing **Organization Profile** so that I don't have to re-type my business address or tax ID.
*   **US.2:** As a user, I want to **preview the document structure** on my desktop before generating the final PDF so that I can verify data correctness.
*   **US.3:** As a user, I want the system to **auto-assign a unique document reference number** (e.g., `INV/2026/05/0001`) so that every document is traceable.
*   **US.4:** As a tenant owner, I want my documents **isolated from other tenants** so that confidential business data is never exposed.

*Note: DIY Template Editor is deferred to v1.1. Tags and folders are deferred to a future version.*

---

## 4. Functional Requirements

### 4.1 Template Engine
*   **Standard Library (v1.0):** Six pre-loaded financial document templates:
    *   Quotation
    *   Invoice
    *   Receipt
    *   Purchase Order
    *   Sales Order
    *   Debit Note
*   **Shared Defaults, Per-Tenant Customization:** Default templates are available to all tenants. Each tenant can customize:
    *   Primary brand color
    *   Watermark toggle (show/hide CozyBox watermark)
    *   Header layout (left-aligned or centered)
    *   Organization logo (uploaded during onboarding or settings)
*   **Version Control:** Templates support "Draft" and "Published" states. Only published templates are available for document generation. Draft changes do not affect existing documents.

### 4.2 Generation Wizard
*   **Hybrid Data Sourcing:** 
    *   **Auto-Fill:** Fetches data from the tenant's Organization Profile (business name, address, tax ID, logo) and Contact records (client name, address).
    *   **Manual Entry:** Provides input fields for dynamic data (e.g., "Total Amount," "Expiry Date," line items).
*   **PDF Export:** High-quality PDF generation via Maroto v2 engine. Generation is an explicit user action ("Generate PDF" button), not live/real-time. PDF is stored in MinIO (S3-compatible) with the S3 key persisted in the database.
*   **Snapshot Integrity:** The PDF is a frozen snapshot at the time of generation. Subsequent edits to the document record do not alter the generated PDF.

### 4.3 Document Manager
*   **CRUD Operations:** Create, read, update, and soft-delete documents. Updates are only allowed while the document is in `draft` status.
*   **Document Lifecycle:** Documents follow a defined status state machine (see Section 9).
*   **Search & Listing:** Filterable by document type, status, contact, and date range. Paginated.
*   **Document Sequences:** Auto-incrementing reference numbers per document type, per tenant, with monthly reset (format: `{PREFIX}/{YYYY}/{MM}/{SEQ:4}`).
*   **Document Conversion:** Ability to convert a Quotation into an Invoice (child references parent via `parent_id`).

### 4.4 Contacts
*   **Contact Types:** Client, Supplier, or Dual.
*   **CRUD Operations:** Full management of contacts scoped to the tenant.
*   **Contact Picker:** Reusable component for associating contacts with documents during generation.

### 4.5 Organization Profile
*   **Tenant-level settings:** Business name, email, phone, address, tax ID, logo, website, timezone, default currency.
*   **Onboarding:** Organization profile is created during onboarding. If onboarding is skipped, the user is **forced to complete it** before accessing the document generation wizard on the Dashboard.

---

## 5. User Experience & Flow (Desktop-First)

1.  **Authentication:** User registers → verifies email → logs in → completes onboarding (organization profile + logo).
2.  **Dashboard:** Central hub showing document statistics and quick-access to document generation.
3.  **The Generation Wizard:** 
    *   **Left Side:** Input form with auto-filled organization fields, contact picker, and manual entry fields (line items, dates, amounts).
    *   **Right Side:** Document structure summary / placeholder. *(Live HTML preview deferred to v1.1.)*
4.  **Finalization:** User clicks "Generate PDF." The system creates the document record, assigns a reference number, renders the PDF via Maroto, uploads to storage, and downloads locally.
5.  **Post-Generation:** User can update document status (e.g., mark as paid) and view the activity timeline.

---

## 6. Acceptance Criteria (Gherkin)

### Scenario: Document Generation with Hybrid Data
*   **Given** I am on the Dashboard with a completed Organization Profile
*   **When** I select the "Standard Invoice" template, pick a contact, and manually enter line items totaling $1,000
*   **And** I click "Generate PDF"
*   **Then** the system should create a document record with a unique reference number (e.g., `INV/2026/05/0001`)
*   **And** generate a PDF containing the organization logo, contact details, line items, and correct totals
*   **And** the PDF should be downloadable.

### Scenario: Overdue Document Detection
*   **Given** an Invoice document with status `published` and `due_date` in the past
*   **When** the scheduled cron job runs
*   **Then** the document status should transition to `overdue`.

### Scenario: Forced Onboarding Before Document Generation
*   **Given** a user who skipped onboarding (no Organization Profile exists)
*   **When** they navigate to the Dashboard and attempt to create a new document
*   **Then** the system should redirect them to complete the Organization Profile before proceeding.

### Scenario: Multi-Tenant Isolation
*   **Given** two tenants, Tenant A and Tenant B
*   **When** Tenant A creates a document and Tenant B queries their document list
*   **Then** Tenant B must not see Tenant A's document.

---

## 7. Success Metrics (KPIs)

*   **Time-to-PDF:** Average time from template selection to download (Target: < 120 seconds).
*   **Documents per Tenant:** Number of documents generated per active tenant per month.
*   **Onboarding Completion:** % of users who complete organization profile setup.
*   **Document Lifecycle Completion:** % of documents that reach a terminal status (paid, canceled, rejected).

---

## 8. Technical Architecture (High-Level)

*   **Structure:** Monolith (Single Binary).
*   **Backend:** **Golang (Gin Framework)** for routing and business logic.
*   **Frontend:** **React (Vite)** embedded into the Go binary via `//go:embed`.
*   **Database:** **PostgreSQL** (master + read replicas via GORM dbresolver).
*   **PDF Engine:** **Maroto v2** — native Go PDF generation. Templates accept per-tenant style configuration (primary color, watermark toggle, header layout).
*   **Storage:** **MinIO** (S3-compatible) for PDF files, organization logos, and uploaded assets.
*   **Caching & Sessions:** **Redis** (3 instances: auth, cache, rate limiter).
*   **Email:** **GoMail** for OTP, invitations, and notifications.
*   **Deployment:** Single Docker image (multi-stage build: Node for UI → Go binary → Alpine runtime).

*Note: Client-side HTML preview is deferred to v1.1. See Section 12 for the planned approach.*

---

## 9. Document Status State Machine

### Status Definitions

| Status | Description | Allowed Transitions |
| :--- | :--- | :--- |
| `draft` | Document is being edited. Not yet published. | `published`, `cancelled` |
| `published` | Document is finalized and reference number assigned. Awaiting recipient action. | `accepted`, `rejected`, `expired`, `overdue`, `paid`, `partially_paid`, `cancelled` |
| `accepted` | Recipient has accepted the document (e.g., quotation accepted). | `paid`, `partially_paid`, `overdue`, `cancelled` |
| `rejected` | Recipient has rejected the document. Terminal. | *(none)* |
| `expired` | `valid_until` date has passed while document was still `published`. No longer valid. | `cancelled` |
| `overdue` | `due_date` has passed without payment/response. Requires follow-up. | `paid`, `partially_paid`, `cancelled` |
| `paid` | Full payment received (`amount_paid >= total`). Terminal. | *(none)* |
| `partially_paid` | Partial payment received (`0 < amount_paid < total`). | `paid`, `overdue`, `cancelled` |
| `cancelled` | Document voided. Terminal. | *(none)* |

### Automatic Transitions (Cron Job)
- **`published` → `expired`:** When `valid_until < NOW()`. Applies to Quotations, Sales Orders, Purchase Orders.
- **`published` or `accepted` → `overdue`:** When `due_date < NOW()` AND `amount_paid < total`. Applies to Invoices, Debit Notes.

### Manual Transitions (User Action)
- **`draft` → `published`:** User clicks "Publish." System assigns reference number.
- **`published` → `accepted` / `rejected`:** User records the recipient's response.
- **Any non-terminal → `cancelled`:** User voids the document.
- **`published` / `accepted` / `overdue` / `partially_paid` → `paid` / `partially_paid`:** User records a payment.

### Rules
- Updates (edit fields, modify items) are only allowed in `draft` status.
- Once `published`, a document is immutable except for status transitions and payment recording.
- Every status change logs an entry in `document_activities`.

---

## 10. Multi-Tenancy Model

*   **Tenant Isolation:** Every document, contact, and organization profile is scoped to a single tenant via `tenant_id`. Database queries enforce this at the service layer through middleware-injected context.
*   **Shared Templates:** Default template configurations are seeded once and available to all tenants. Each tenant can customize their copy (brand color, watermark, logo) without affecting other tenants.
*   **Template Customization per Tenant:** When a document is generated, the system loads the tenant's `template_configs` row for that document type. If the tenant has not customized it, system defaults are used. The tenant's organization logo is injected into the PDF at generation time.
*   **Tenant Membership:** A tenant has one owner and can have multiple members. All members of a tenant can access that tenant's documents (future: role-based access control).

---

## 11. Out of Scope for v1.0

The following features are explicitly **not included** in v1.0:

| Feature | Target Version | Notes |
| :--- | :--- | :--- |
| DIY Template Editor | v1.1 | Custom placeholders, user-created template bodies |
| Client-side HTML Preview | v1.1 | Right-panel live document preview; v1.0 uses a placeholder summary |
| Tags & Folder Organization | Future | Document categorization beyond status filtering |
| E-Signature Integration | Future | Digital signing workflow |
| Document Sharing (External Links) | Future | Shareable URLs for non-tenant users |
| Mobile / Tablet Support | Future | Desktop-first (≥1024px) is the sole target for v1.0 |
| Bulk Document Operations | Future | Multi-select publish, export, or status change |
| Custom Fonts in Templates | v1.1 | Template font selection; v1.0 uses default embedded font |
| Role-Based Access Control | Future | All tenant members have full access in v1.0 |

---

## 12. Future: HTML Preview Implementation Plan (v1.1)

*Context: v1.0 generates PDFs as an explicit user action only. v1.1 will introduce a client-side HTML preview in the right panel of the generation wizard.*

### Approach
1.  **Shared Config:** Both the React preview component and the Go Maroto renderer will consume the same `template_configs.config` JSON. This ensures consistent colors, watermark state, and header layout.
2.  **React Preview Components:** One component per document type. Each mirrors the layout structure of its Maroto counterpart using CSS (flexbox/grid) and Tailwind utilities.
3.  **No Server Calls:** The preview renders entirely client-side from form state. No API requests during preview.
4.  **Preview Modes:** Users can toggle between "Preview" (HTML approximation) and "Summary" (structured data view).
5.  **Known Fidelity Gaps:**
    *   Page breaks: HTML cannot perfectly replicate Maroto's pagination. A visual page-break indicator will be shown.
    *   Fonts: The browser's font may differ from the PDF's embedded font. A disclaimer will note this.
    *   Positioning: Exact coordinate positioning differs between CSS and Maroto's grid system. Layout will be "structurally equivalent" rather than pixel-perfect.
6.  **Preview Banner:** A persistent banner at the top of the preview panel: *"This is an approximate preview. Final PDF layout may differ slightly."*

### Delivery
- Backend: No changes needed (template_configs schema already supports it).
- Frontend: Build preview components as part of v1.1 milestone; swap into the split-view's right panel replacing v1.0's placeholder.
