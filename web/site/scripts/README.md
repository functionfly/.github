# Translation Guide

This folder contains the DeepL-powered auto-translation script for `src/i18n/ui.ts`.

## Prerequisites

1. Get a **DeepL API key** at https://deepl.com/pro-api (free tier: 500k chars/month)
2. Set it in your `.env` file:

```env
DEEPL_API_KEY=your-key-here
```

## Usage

```bash
# Preview what needs translating (dry run — no writes)
bun scripts/translate-ui.ts

# Actually write the translations to ui.ts
DEEPL_API_KEY=... bun scripts/translate-ui.ts --write
```

## How it works

1. Reads `src/i18n/ui.ts` and finds the English block (`en:`)
2. Identifies all translatable strings that are missing or empty in target languages
3. Sends batches of 30 strings at a time to DeepL
4. Writes translated strings back into the `ui.ts` file under the correct language key

## Adding new strings

After adding new English strings to `ui.ts`, just re-run:

```bash
bun scripts/translate-ui.ts --write
```

Only strings that are missing or empty will be re-translated.

## Supported languages

`en` (English), `es` (Spanish), `fr` (French), `de` (German), `zh` (Chinese),
`ja` (Japanese), `ko` (Korean), `pt` (Portuguese), `ar` (Arabic), `ru` (Russian),
`hi` (Hindi), `nl` (Dutch), `pl` (Polish), `tr` (Turkish), `vi` (Vietnamese)
