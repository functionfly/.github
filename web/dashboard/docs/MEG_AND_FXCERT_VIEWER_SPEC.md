# MEG Graph & FXCERT Viewer – Component Spec

Users **should** be able to view the Merkle Execution Graph (MEG) and FXCERTs. The backend already exposes the data; the dashboard has DRE components but is missing certificate API wiring and a few UI pieces.

---

## Current state

| Area | Backend | Dashboard |
|------|--------|-----------|
| **MEG (executions)** | `GET /registry/{author}/{name}/executions` (list), `GET .../executions/{id}` (detail with `component_hashes`) | ✅ ExecutionExplorerPage, `MerkleExecutionTree`, `dreApi.listExecutions` / `getExecution` |
| **FXCERT list** | `GET /registry/{author}/{name}/certs` (paginated) | ❌ No API client, no list UI |
| **FXCERT detail** | `GET /registry/{author}/{name}/cert/{cert_id}` (full FXCERT JSON in `cert`) | ❌ No API client; execution detail only has cert summary, not full `cert` |
| **Execution ↔ cert link** | Execution detail returns `certificate: { certificate_id, ... }` when found | ⚠️ Backend links cert to execution by `ExecutionID`; ensure this is correct so “View FXCERT” works |

---

## Custom React components we need

### 1. **API client (DRE)** – `src/api/dre.ts`

- **`listCertificates(author, name, params?: { limit?, offset? })`**  
  - `GET /v1/registry/{author}/{name}/certs`  
  - Returns `{ function, certs, limit, offset }`; each cert: `certificate_id`, `cert_level`, `execution_root_hash`, `certificate_hash`, `anchored`, `created_at`.

- **`getCertificate(author, name, certId)`**  
  - `GET /v1/registry/{author}/{name}/cert/{cert_id}`  
  - Returns full response: `certificate_id`, `cert_level`, `execution_root_hash`, `certificate_hash`, `created_at`, `anchored`, **`cert`** (full FXCERT: Execution, Capsule, Integrity, Trust, Signatures, Anchoring).

- **Types**: Add `CertificateListItem`, `CertificateDetailResponse`, and a type for the backend `cert` object (or reuse a single FXCERT type that matches backend FXCert).

### 2. **CertificateList** (new) – `src/components/dre/certificate/CertificateList.tsx`

- Paginated list of FXCERTs for a function.
- Uses `dreApi.listCertificates(author, name, { limit, offset })`.
- Renders a row/card per cert: `certificate_id`, `created_at`, `cert_level`, truncated `execution_root_hash`, anchored badge, “View” action.
- “View” opens full FXCERT (see below).

### 3. **CertificateCard** (new) – `src/components/dre/certificate/CertificateCard.tsx`

- Single card for one certificate in the list.
- Props: `certificate_id`, `created_at`, `cert_level`, `execution_root_hash`, `anchored`, `onView()`.
- Optional: small `HashBlock` for root hash.

### 4. **FXCertViewer** (existing – enhance)

- **Current**: Expects `FXCertData` (flat: level, hashes, signatures, anchor). Execution detail only has a cert summary, so we build a partial `FXCertData` (with placeholder signatures/expiry).
- **Enhancement**: Support full backend FXCERT when available:
  - Either add a **mapper** `mapBackendCertToFXCertData(apiResponse)` that maps `cert` (ExecutionSection, IntegritySection, etc.) into existing `FXCertData`, or
  - Add an optional “raw” mode that renders the full `cert` (Execution, Capsule, Integrity, Trust, Signatures, Anchoring) in collapsible sections.
- **Download**: Implement real download (JSON/CBOR/PDF) using the full cert from `getCertificate`.

### 5. **Full FXCERT view (modal or page)**

- When user clicks “View” on a cert (from CertificateList or from execution detail “View full FXCERT”):
  - Call `dreApi.getCertificate(author, name, certificate_id)`.
  - Render `FXCertViewer` with mapped (or raw) cert data.
- Can be:
  - **Modal/drawer** on Execution Explorer, or
  - **Route** e.g. `/registry/:author/:name/certs/:certId` that shows only the viewer (good for sharing links).

### 6. **Execution Explorer enhancements** – `ExecutionExplorerPage`

- **Tabs**: Add a **“Certificates”** tab next to the current executions list. Content: `<CertificateList author={author} name={name} />`.
- **Execution detail**: Keep existing MEG + cert summary. Add a **“View full FXCERT”** button when `execution.certificate` exists; on click, fetch `getCertificate(author, name, execution.certificate.certificate_id)` and open the full FXCERT in a modal or navigate to cert route.

### 7. **MEG graph (optional)**

- **MerkleExecutionTree** already shows the MEG as a tree of component hashes (input, output, environment, dependency, trace, resource, metadata). No change required for “viewing the MEG.”
- **Optional – MEGGraph**: If you want a visual Merkle *graph* (nodes = hashes, edges = parent/child), add a component that takes `execution_root_hash` + `component_hashes` and draws a small tree (e.g. root at top, children below, using react-flow or SVG). Not required for “viewing” the MEG; only for a more diagram-like view.

### 8. **Navigation**

- Function page already links to “Executions” (Execution Explorer). Add a link to **“Certificates”** (same explorer with Certificates tab active, or a dedicated `/registry/:author/:name/certs` route).
- Ensure both “Executions” and “Certificates” are reachable from the function/playground header or tabs.

---

## Summary checklist

| # | Item | Type |
|---|------|------|
| 1 | `listCertificates` + `getCertificate` + types in `api/dre.ts` | API |
| 2 | `CertificateList` (paginated list of certs) | New component |
| 3 | `CertificateCard` (one cert card) | New component |
| 4 | Map backend `cert` → `FXCertData` and/or raw FXCERT view in `FXCertViewer` | Enhance |
| 5 | Full FXCERT view (modal or route) using `getCertificate` | New view |
| 6 | “Certificates” tab + “View full FXCERT” on Execution Explorer | Page enhance |
| 7 | (Optional) MEGGraph – visual Merkle tree diagram | Optional component |
| 8 | Nav link to Certificates from function page | Nav |

---

## Backend note

- **Execution ↔ certificate link**: `HandleGetExecution` currently finds a cert by `cert.ExecutionID == executionID`. In `buildAndStoreMEG`, the stored certificate uses `ExecutionID: megRecord.ID`. So the execution ID in the list/detail is `rec.ExecutionID` (the execution UUID). If the certificate stores a different ID (e.g. MEG record ID), the execution detail may not show the certificate. Verify and fix backend linking so that execution detail returns the correct `certificate_id` when a cert exists for that execution (e.g. by linking cert to execution ID or MEG record ID consistently).
