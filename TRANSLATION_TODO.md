# Translation TODO

Remaining translation work for the FunctionFly dashboard i18n implementation.

## Status

| Metric | Value |
|--------|-------|
| Total keys | 1,406 |
| Translated per locale | ~800 (57%) |
| English fallback per locale | ~600 (43%) |
| Missing keys | 0 |

## Sections Needing Translation

Translate these sections in **all 14 locale files** (`es`, `fr`, `de`, `zh`, `ja`, `ko`, `pt`, `ar`, `ru`, `hi`, `nl`, `pl`, `tr`, `vi`).

Run `bun scripts/sync-translations.ts --write` to auto-translate via MyMemory API (5000 words/day free).

| Section | Keys | Description |
|---------|------|-------------|
| `funcEditor` | 237 | Function editor page — all form labels, sections, dialogs, templates, shortcuts |
| `playground` | 106 | API playground — header, toolbar, input/output panels, sidebar, status bar, examples, schema, timeline, diff viewer |
| `appsPage` | 89 | Create app + apps list — form labels, visibility, tags, environments, sorting, empty states |
| `agentMarket` | 43 | Agent marketplace component — filters, cards, categories, ratings, install flow |
| `notifCenter` | 33 | Notification center — tabs, filters, empty states, actions, priority labels |
| `pricing` | 29 | Pricing page — plan names, features, checkout dialog, value strip |
| `cookieConsent` | 17 | Cookie consent modal — categories, descriptions, buttons |
| `embedCode` | 16 | Embed code generator — platform options, copy buttons, instructions |
| `realtimeNotif` | 14 | Realtime notification center — connection status, actions, empty state |
| `achievement` | 6 | Achievement badge — rarity labels, tooltip text |
| `passwordStrength` | 5 | Password strength indicator — strength levels |

## Acceptable English Fallbacks

These keys intentionally stay in English across all locales:

- `nav.aiComposer` — product name
- `nav.sdk` — universal acronym
- `nav.stateFabric` — product feature name
- `nav.state` — short label
- `common.error` — universal word
- `functionDetail.runtime` — technical term
- `usermenu.personal` — short label
- `profile.companyPlaceholder` — example company name ("Acme Inc")

## Quick Translate Command

```bash
# Dry run — check what's missing
bun scripts/sync-translations.ts

# Translate and write
bun scripts/sync-translations.ts --write

# CI check (exits 1 if missing)
bun scripts/sync-translations.ts --check
```

## File Locations

```
web/dashboard/src/locales/
├── en/common.json      (source of truth — 1,406 keys, 40 sections)
├── ar/common.json
├── de/common.json
├── es/common.json
├── fr/common.json
├── hi/common.json
├── ja/common.json
├── ko/common.json
├── nl/common.json
├── pl/common.json
├── pt/common.json
├── ru/common.json
├── tr/common.json
├── vi/common.json
└── zh/common.json
```
