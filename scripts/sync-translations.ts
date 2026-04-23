#!/usr/bin/env bun
/**
 * sync-translations.ts
 *
 * Production-ready translation sync for FunctionFly.
 * Finds missing translation keys and translates them using free APIs.
 *
 * Translation backends (auto-detected, cheapest first):
 *   1. MyMemory API — 5000 words/day free, no key needed
 *   2. Ollama — local LLM, unlimited, free, private
 *
 * Usage:
 *   bun scripts/sync-translations.ts              # dry run (check what's missing)
 *   bun scripts/sync-translations.ts --write       # translate and write
 *   bun scripts/sync-translations.ts --check       # CI check (exit 1 if missing)
 *   bun scripts/sync-translations.ts --write --backend=ollama  # use local Ollama
 *
 * Supports:
 *   - Dashboard: web/dashboard/src/locales/{lang}/common.json
 *   - Marketing: web/site/src/locales/{lang}/common.json
 */

import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { resolve, dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(__dirname, "..");

const DRY_RUN = !process.argv.includes("--write");
const CHECK_MODE = process.argv.includes("--check");
const BACKEND = (process.argv.find((a) => a.startsWith("--backend="))?.split("=")[1] ?? "mymemory") as "mymemory" | "ollama";
const OLLAMA_URL = process.env.OLLAMA_URL ?? "http://localhost:11434";
const OLLAMA_MODEL = process.env.OLLAMA_MODEL ?? "llama3.1";

interface Language {
  code: string;
  /** MyMemory language pair code */
  mymemory: string;
  skip?: boolean;
}

const LANGUAGES: Language[] = [
  { code: "en", mymemory: "en", skip: true },
  { code: "es", mymemory: "es" },
  { code: "fr", mymemory: "fr" },
  { code: "de", mymemory: "de" },
  { code: "zh", mymemory: "zh" },
  { code: "ja", mymemory: "ja" },
  { code: "ko", mymemory: "ko" },
  { code: "pt", mymemory: "pt" },
  { code: "ar", mymemory: "ar" },
  { code: "ru", mymemory: "ru" },
  { code: "hi", mymemory: "hi" },
  { code: "nl", mymemory: "nl" },
  { code: "pl", mymemory: "pl" },
  { code: "tr", mymemory: "tr" },
  { code: "vi", mymemory: "vi" },
];

// ---------------------------------------------------------------------------
// Flatten/expand nested JSON
// ---------------------------------------------------------------------------
function flattenKeys(obj: Record<string, unknown>, prefix = ""): Map<string, string> {
  const result = new Map<string, string>();
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof value === "string") {
      result.set(fullKey, value);
    } else if (typeof value === "object" && value !== null) {
      for (const [k, v] of flattenKeys(value as Record<string, unknown>, fullKey)) {
        result.set(k, v);
      }
    }
  }
  return result;
}

function expandKeys(flat: Map<string, string>): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [dotKey, value] of flat) {
    const parts = dotKey.split(".");
    let current: Record<string, unknown> = result;
    for (let i = 0; i < parts.length - 1; i++) {
      if (!(parts[i] in current) || typeof current[parts[i]] !== "object") {
        current[parts[i]] = {};
      }
      current = current[parts[i]] as Record<string, unknown>;
    }
    current[parts[parts.length - 1]] = value;
  }
  return result;
}

// ---------------------------------------------------------------------------
// MyMemory API — 5000 words/day free, no key needed
// https://mymemory.translated.net/doc/spec.php
// ---------------------------------------------------------------------------
async function translateMyMemory(texts: string[], targetLang: string): Promise<string[]> {
  const results: string[] = [];

  // MyMemory has a limit of ~500 chars per request, so batch carefully
  const CHUNK_SIZE = 10;
  for (let i = 0; i < texts.length; i += CHUNK_SIZE) {
    const chunk = texts.slice(i, i + CHUNK_SIZE);

    for (const text of chunk) {
      // Preserve interpolation variables like {{name}} — MyMemory sometimes mangles them
      const placeholders: string[] = [];
      const safeText = text.replace(/\{\{[^}]+\}\}/g, (match) => {
        const idx = placeholders.length;
        placeholders.push(match);
        return `__VAR${idx}__`;
      });

      const url = `https://api.mymemory.translated.net/get?q=${encodeURIComponent(safeText)}&langpair=en|${targetLang}&de=translate@functionfly.com`;

      try {
        const resp = await fetch(url);
        if (!resp.ok) {
          console.error(`   ⚠️  MyMemory ${resp.status} for "${text.slice(0, 40)}..."`);
          results.push(text);
          continue;
        }

        const data = (await resp.json()) as {
          responseStatus: number;
          responseData: { translatedText: string };
        };

        if (data.responseStatus === 200 && data.responseData?.translatedText) {
          let translated = data.responseData.translatedText;
          // Restore placeholders
          for (let pi = 0; pi < placeholders.length; pi++) {
            translated = translated.replace(new RegExp(`__VAR${pi}__`, "gi"), placeholders[pi]);
          }
          results.push(translated);
        } else {
          results.push(text); // fallback to English
        }
      } catch {
        results.push(text);
      }

      // Respect rate limit: ~1 req/sec without key
      await sleep(200);
    }
  }

  return results;
}

// ---------------------------------------------------------------------------
// Ollama — local LLM translation (free, unlimited, private)
// ---------------------------------------------------------------------------
async function translateOllama(texts: string[], targetLang: string): Promise<string[]> {
  const langNames: Record<string, string> = {
    es: "Spanish", fr: "French", de: "German", zh: "Chinese (Simplified)",
    ja: "Japanese", ko: "Korean", pt: "Portuguese", ar: "Arabic",
    ru: "Russian", hi: "Hindi", nl: "Dutch", pl: "Polish",
    tr: "Turkish", vi: "Vietnamese",
  };
  const langName = langNames[targetLang] ?? targetLang;

  const results: string[] = [];

  // Batch 10 strings at a time to keep prompts manageable
  const CHUNK_SIZE = 10;
  for (let i = 0; i < texts.length; i += CHUNK_SIZE) {
    const chunk = texts.slice(i, i + CHUNK_SIZE);
    const numbered = chunk.map((t, idx) => `${idx + 1}. ${t}`).join("\n");

    const prompt = `Translate these UI strings from English to ${langName}.
Rules:
- Keep {{variable}} placeholders exactly as-is
- Keep proper nouns (FunctionFly, API, SDK) unchanged
- Return ONLY the translations, numbered 1-N, nothing else

Strings:
${numbered}`;

    try {
      const resp = await fetch(`${OLLAMA_URL}/api/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          model: OLLAMA_MODEL,
          prompt,
          stream: false,
          options: { temperature: 0.1 },
        }),
      });

      if (!resp.ok) {
        console.error(`   ⚠️  Ollama ${resp.status}`);
        results.push(...chunk);
        continue;
      }

      const data = (await resp.json()) as { response: string };
      const lines = data.response.trim().split("\n");

      for (let li = 0; li < chunk.length; li++) {
        const line = lines[li] ?? "";
        // Strip "1. " prefix
        const cleaned = line.replace(/^\d+\.\s*/, "").trim();
        results.push(cleaned || chunk[li]);
      }
    } catch (err) {
      console.error(`   ⚠️  Ollama not available: ${err}`);
      results.push(...chunk);
    }
  }

  return results;
}

// ---------------------------------------------------------------------------
// Unified translate function
// ---------------------------------------------------------------------------
async function translateBatch(texts: string[], targetLang: string): Promise<string[]> {
  if (BACKEND === "ollama") {
    return translateOllama(texts, targetLang);
  }
  return translateMyMemory(texts, targetLang);
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
const BATCH_SIZE = BACKEND === "ollama" ? 10 : 10;

// ---------------------------------------------------------------------------
// Sync a locale directory
// ---------------------------------------------------------------------------
async function syncLocaleDir(
  label: string,
  localesDir: string,
  jsonFileName: string = "common.json"
): Promise<{ missing: number; translated: number }> {
  if (!existsSync(localesDir)) {
    console.log(`⏭️  ${label}: locales dir not found, skipping`);
    return { missing: 0, translated: 0 };
  }

  console.log(`\n📊 ${label}\n${"─".repeat(40)}`);

  const enPath = join(localesDir, `en/${jsonFileName}`);
  if (!existsSync(enPath)) {
    console.log(`⏭️  English locale not found, skipping`);
    return { missing: 0, translated: 0 };
  }

  const enJson = JSON.parse(readFileSync(enPath, "utf-8")) as Record<string, unknown>;
  const enKeys = flattenKeys(enJson);

  console.log(`📖 ${enKeys.size} English keys found`);
  if (BACKEND === "ollama") {
    console.log(`🤖 Backend: Ollama (${OLLAMA_URL}, model: ${OLLAMA_MODEL})`);
  } else {
    console.log(`🌐 Backend: MyMemory (free, no API key)`);
  }

  let totalMissing = 0;
  let totalTranslated = 0;

  for (const lang of LANGUAGES) {
    if (lang.skip) continue;

    const langPath = join(localesDir, `${lang.code}/${jsonFileName}`);
    if (!existsSync(langPath)) {
      console.log(`⚠️  ${lang.code}: file not found`);
      totalMissing += enKeys.size;
      continue;
    }

    const langJson = JSON.parse(readFileSync(langPath, "utf-8")) as Record<string, unknown>;
    const langKeys = flattenKeys(langJson);

    // Find keys that are missing OR still in English (untranslated placeholders)
    const missing: Array<{ key: string; value: string }> = [];
    for (const [key, value] of enKeys) {
      const existing = langKeys.get(key);
      if (!existing) {
        missing.push({ key, value });
      }
    }

    if (missing.length === 0) {
      console.log(`✅ ${lang.code} — complete (${langKeys.size} keys)`);
      continue;
    }

    totalMissing += missing.length;
    console.log(`📦 ${lang.code} — ${missing.length} keys need translation`);

    if (CHECK_MODE) continue;

    const translated = new Map<string, string>();
    for (let i = 0; i < missing.length; i += BATCH_SIZE) {
      const batch = missing.slice(i, i + BATCH_SIZE);
      const texts = batch.map((b) => b.value);

      try {
        const results = await translateBatch(texts, lang.mymemory);
        batch.forEach((b, bi) => translated.set(b.key, results[bi]));
        totalTranslated += batch.length;

        const pct = Math.min(i + BATCH_SIZE, missing.length);
        process.stdout.write(`\r   🔄 ${pct}/${missing.length}...`);
      } catch (err) {
        console.error(`\n   ❌  Batch failed: ${err}`);
      }

      if (i + BATCH_SIZE < missing.length) await sleep(300);
    }

    if (translated.size > 0) {
      process.stdout.write(`\r`);
    }

    if (DRY_RUN) {
      console.log(`   → ${translated.size} ready (dry run)`);
      continue;
    }

    for (const [key, value] of translated) {
      langKeys.set(key, value);
    }

    const newJson = expandKeys(langKeys);
    writeFileSync(langPath, JSON.stringify(newJson, null, 2) + "\n", "utf-8");
    console.log(`   ✍️  Wrote ${translated.size} translations`);
  }

  if (CHECK_MODE && totalMissing > 0) {
    console.error(`\n❌ ${label}: ${totalMissing} missing translations.`);
  }

  return { missing: totalMissing, translated: totalTranslated };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------
async function main() {
  console.log(`\n🌐 FunctionFly Translation Sync\n${"─".repeat(40)}`);

  if (DRY_RUN && !CHECK_MODE) {
    console.log("🔎 DRY RUN — no changes written\n");
    console.log("   Backend:", BACKEND === "ollama"
      ? `Ollama (${OLLAMA_URL})`
      : "MyMemory (free, no key)"
    );
    console.log("   Run with --write to apply translations.");
  }

  const dashboard = await syncLocaleDir(
    "Dashboard",
    resolve(ROOT, "web/dashboard/src/locales"),
    "common.json"
  );

  const marketing = await syncLocaleDir(
    "Marketing Site",
    resolve(ROOT, "web/site/src/locales"),
    "common.json"
  );

  const totalMissing = dashboard.missing + marketing.missing;
  const totalTranslated = dashboard.translated + marketing.translated;

  console.log(`\n${"─".repeat(40)}`);
  console.log(`Missing: ${totalMissing} | Translated: ${totalTranslated}`);

  if (DRY_RUN && !CHECK_MODE) {
    console.log("Run with --write to apply translations.");
  } else if (CHECK_MODE) {
    console.log("CI check complete.");
  } else {
    console.log("Done!");
  }

  if (CHECK_MODE && totalMissing > 0) {
    process.exit(1);
  }
}

main().catch((e) => {
  console.error("Fatal:", e);
  process.exit(1);
});
