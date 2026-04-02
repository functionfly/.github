#!/usr/bin/env bun
/**
 * translate-ui.ts
 *
 * Uses DeepL to auto-translate missing strings in src/i18n/ui.ts.
 *
 * Usage:
 *   DEEPL_API_KEY=... bun scripts/translate-ui.ts          # dry run
 *   DEEPL_API_KEY=... bun scripts/translate-ui.ts --write  # write translations
 */

import { readFileSync, writeFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const UI_TS_PATH = resolve(__dirname, "../src/i18n/ui.ts");
const DRY_RUN = !process.argv.includes("--write");
const DEEPL_API_KEY = process.env.DEEPL_API_KEY ?? "";

// ---------------------------------------------------------------------------
// Language config (matches src/lib/i18n/languages.ts)
// ---------------------------------------------------------------------------
const LANGUAGES = [
  { code: "en", deeplCode: "EN", skip: true },
  { code: "es", deeplCode: "ES", skip: false },
  { code: "fr", deeplCode: "FR", skip: false },
  { code: "de", deeplCode: "DE", skip: false },
  { code: "zh", deeplCode: "ZH", skip: false },
  { code: "ja", deeplCode: "JA", skip: false },
  { code: "ko", deeplCode: "KO", skip: false },
  { code: "pt", deeplCode: "PT-BR", skip: false },
  { code: "ar", deeplCode: "AR", skip: false },
  { code: "ru", deeplCode: "RU", skip: false },
  { code: "hi", deeplCode: "HI", skip: false },
  { code: "nl", deeplCode: "NL", skip: false },
  { code: "pl", deeplCode: "PL", skip: false },
  { code: "tr", deeplCode: "TR", skip: false },
  { code: "vi", deeplCode: "VI", skip: false },
];

// ---------------------------------------------------------------------------
// Parse ui.ts into: { enStrings: Map<key, value>, allLangs: {code, start, end}[] }
// ---------------------------------------------------------------------------
interface ParsedLang { code: string; lines: Array<{ indent: number; key: string; value: string }>; }

function parseUiTs(source: string): { enStrings: Map<string, string>; langs: ParsedLang[] } {
  const lines = source.split("\n");
  const enStrings = new Map<string, string>();
  const langs: ParsedLang[] = [];
  let currentLang: ParsedLang | null = null;
  let braceDepth = 0;

  for (const rawLine of lines) {
    const line = rawLine.replace(/\/\/.*$/, ""); // strip line comments
    const indent = rawLine.match(/^(\s*)/)?.[1].length ?? 0;

    // Detect language block start: "  xx: {"
    const langOpen = line.match(/^\s*["'](\w+)["']\s*:\s*\{/);
    if (langOpen && braceDepth === 0) {
      if (currentLang) langs.push(currentLang);
      currentLang = { code: langOpen[1], lines: [] };
      braceDepth = 1;
      continue;
    }

    if (braceDepth > 0 && currentLang) {
      if (line.includes("{")) braceDepth++;
      if (line.includes("}")) braceDepth--;
      if (braceDepth === 0) { langs.push(currentLang); currentLang = null; continue; }

      // Match string key: "key": "value"
      const kvMatch = rawLine.match(/^\s*["'](\S+?)["']\s*:\s*(?:"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)')/);
      if (kvMatch) {
        const key = kvMatch[1];
        const value = (kvMatch[2] ?? kvMatch[3] ?? "").replace(/\\"/g, '"').replace(/\\'/g, "'");
        if (currentLang.code === "en") enStrings.set(key, value);
        currentLang.lines.push({ indent, key, value });
      }
    }
  }

  return { enStrings, langs };
}

// ---------------------------------------------------------------------------
// DeepL
// ---------------------------------------------------------------------------
async function translateTexts(texts: string[], targetLang: string): Promise<string[]> {
  if (!DEEPL_API_KEY) throw new Error("DEEPL_API_KEY is not set");
  if (texts.length === 0) return [];

  const body = new URLSearchParams({
    auth_key: DEEPL_API_KEY,
    text: texts.join("\n"),
    target_lang: targetLang,
    tag_handling: "xml",
    ignore_tags: "code",
    preserve_formatting: "1",
  });

  const resp = await fetch("https://api-free.deepl.com/v2/translate", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });

  if (!resp.ok) throw new Error(`DeepL ${resp.status}: ${await resp.text()}`);
  const data = await resp.json() as { translations: Array<{ text: string }> };
  return data.translations.map((t) => t.text);
}

// ---------------------------------------------------------------------------
// Indent helper
// ---------------------------------------------------------------------------
function indent(n: number): string { return "  ".repeat(n); }

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------
async function main() {
  console.log(`\n🌐 DeepL Auto-Translator for ui.ts\n${"─".repeat(40)}`);

  if (!DEEPL_API_KEY) {
    console.error("❌  DEEPL_API_KEY is not set.\n");
    console.error("    Get one at https://deepl.com/pro-api (free: 500k chars/month)\n");
    console.error("    Then:  DEEPL_API_KEY=... bun scripts/translate-ui.ts --write\n");
    process.exit(1);
  }

  if (DRY_RUN) console.log("🔎 DRY RUN — no changes will be written\n");

  const source = readFileSync(UI_TS_PATH, "utf-8");
  const { enStrings, langs } = parseUiTs(source);

  console.log(`📖 Parsed ${enStrings.size} English strings across ${langs.length} languages`);

  const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
  const BATCH = 25;

  // For each language, find missing strings and translate them
  for (const lang of LANGUAGES) {
    if (lang.skip) continue;

    const langData = langs.find((l) => l.code === lang.code);
    if (!langData) { console.log(`⚠️  No '${lang.code}' block found in ui.ts — skipping`); continue; }

    const existingKeys = new Set(langData.lines.map((l) => l.key));

    // Strings that need translation (in en but missing/incomplete in lang)
    const toTranslate: Array<{ key: string; enValue: string }> = [];
    for (const [key, enValue] of enStrings) {
      const existing = langData.lines.find((l) => l.key === key);
      if (!existing || !existing.value || existing.value === key) {
        toTranslate.push({ key, enValue });
      }
    }

    if (toTranslate.length === 0) {
      console.log(`✅  ${lang.code} — complete (${existingKeys.size} strings)`);
      continue;
    }

    console.log(`\n📦 ${lang.code} — ${toTranslate.length} strings need translation`);

    const translatedMap = new Map<string, string>();

    for (let i = 0; i < toTranslate.length; i += BATCH) {
      const batch = toTranslate.slice(i, i + BATCH);
      const englishTexts = batch.map((b) => b.enValue);

      try {
        const translated = await translateTexts(englishTexts, lang.deeplCode);

        if (DRY_RUN) {
          batch.forEach((b, bi) => {
            console.log(`   [${b.key}]`);
            console.log(`     EN: "${b.enValue}"`);
            console.log(`     ${lang.code.toUpperCase()}: "${translated[bi]}"`);
          });
        }

        batch.forEach((b, bi) => translatedMap.set(b.key, translated[bi]));
      } catch (err) {
        console.error(`   ❌  Batch failed: ${err}`);
        continue;
      }

      if (i + BATCH < toTranslate.length) await sleep(300);
    }

    if (DRY_RUN) {
      console.log(`   → ${translatedMap.size} ready to write (dry run only)`);
      continue;
    }

    // Build new language block
    const existingLines = langData.lines;
    const existingKeysSet = new Set(existingLines.map((l) => l.key));
    const newLines: Array<{ indent: number; key: string; value: string }> = [...existingLines];

    // Add new translations
    for (const [key, value] of translatedMap) {
      if (!existingKeysSet.has(key)) {
        const enEntry = enStrings.get(key);
        // Find the indent of a sibling key (same section)
        const sibling = existingLines.find((l) => l.key.startsWith(key.split(".")[0]));
        const newIndent = sibling?.indent ?? 4;
        newLines.push({ indent: newIndent, key, value });
      }
    }

    // Sort new lines by key for consistency
    newLines.sort((a, b) => a.key.localeCompare(b.key));

    // Find brace depth lines
    const langStartIdx = source.split("\n").findIndex(
      (l) => l.match(new RegExp(`["']${lang.code}["']\\s*:\\s*\\{`))
    );
    const afterOpenBrace = langStartIdx + 1;

    // Reconstruct the block
    const blockLines = newLines.map((l) => `${indent(l.indent)}  "${l.key}": "${l.value.replace(/"/g, '\\"')}",`);
    const closingLineIdx = langStartIdx + 1 + newLines.length + 1;

    // Replace in source
    const srcLines = source.split("\n");
    // Remove old closing brace line
    const cleanLines = srcLines.filter((_, i) =>
      !(i === langStartIdx + 1 + existingLines.length && srcLines[i].trim() === "}")
    );
    // Insert translated lines
    cleanLines.splice(langStartIdx + 1, 0, ...blockLines, `${indent(3)}});

    const newSource = cleanLines.join("\n");
    writeFileSync(UI_TS_PATH, newSource, "utf-8");
    console.log(`   ✍️  Wrote ${translatedMap.size} translations for ${lang.code}`);
  }

  console.log(DRY_RUN ? "\n✅ Dry run complete. Run with --write to apply." : "\n✅ All done!");
}

main().catch((e) => { console.error("Fatal:", e); process.exit(1); });
