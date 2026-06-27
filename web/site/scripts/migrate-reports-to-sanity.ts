/**
 * Migrate existing .md report files to Sanity CMS.
 * Usage: SANITY_STUDIO_PROJECT_ID=xxx SANITY_STUDIO_DATASET=production SANITY_API_TOKEN=xxx bun run scripts/migrate-reports-to-sanity.ts
 */
import { createClient } from "@sanity/client";
import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";

const projectId = process.env.SANITY_STUDIO_PROJECT_ID;
const dataset = process.env.SANITY_STUDIO_DATASET || "production";
const token = process.env.SANITY_API_TOKEN;

if (!projectId) {
  console.error("Missing SANITY_STUDIO_PROJECT_ID");
  process.exit(1);
}

const sanity = createClient({
  projectId,
  dataset,
  apiVersion: "2024-01-01",
  token,
  useCdn: false,
});

interface ReportDoc {
  _type: "report";
  _id: string;
  title: string;
  slug: { _type: "slug"; current: string };
  periodStart: string;
  periodEnd: string;
  generatedAt: string;
  tldrMetros: number;
  tldrUniversities: number;
  tldrActiveBuilders: number;
  tldrNewBuilders: number;
  tldrDeployments: number;
  tldrExecutions: number;
  tldrAmbassadors: number;
  topMetros: Array<{
    rank: number;
    city: string;
    country: string;
    population: number;
    perCapita: number;
    builders: number;
  }>;
  topUniversities: Array<{
    rank: number;
    university: string;
    country: string;
    students: number;
    perCapita: number;
    builders: number;
  }>;
  biggestGainers: Array<{
    delta: string;
    city: string;
    newRank: string;
    perCapita: number;
  }>;
  biggestLosers: Array<{
    delta: string;
    city: string;
    newRank: string;
    perCapita: number;
  }>;
  body: string;
}

function parseReport(raw: string, slug: string): ReportDoc {
  const titleMatch = raw.match(/^# (.+)$/m);
  const periodMatch = raw.match(
    /_Period: ([\d-]+) → ([\d-]+) · Generated (.+?)_/m,
  );
  const tldrMatch = raw.match(
    /- \*\*(\d+) metros\*\* ranked, \*\*(\d+) universities\*\* ranked\.\n\n- \*\*(\d+)\*\* total active builders in the last 30 days\.\n- \*\*(\d+)\*\* new builders joined in the last 30 days\.\n- \*\*(\d+)\*\* deployments and \*\*(\d+)\*\* function executions\.\n- \*\*(\d+) new ambassadors\*\* promoted\./,
  );

  // Parse top metros table
  const metrosSection = raw.match(
    /## Top 10 metros\n\n\| # \| City \| Country \| Pop\. \| Per-capita \| Builders \|\n\|---\|---\|---\|-----:\|-----------:\|---------:\|\n([\s\S]*?)(?=## Top 10 universities|$)/,
  );
  const topMetros: ReportDoc["topMetros"] = [];
  if (metrosSection) {
    for (const line of metrosSection[1].split("\n")) {
      const m = line.match(
        /\| (\d+) \| (.+?), (.+?) \| (.+?) \| ([\d.]+) \| (\d+) \|/,
      );
      if (m) {
        topMetros.push({
          rank: Number(m[1]),
          city: m[2],
          country: m[3],
          population: Number(m[4].replace(/,/g, "")),
          perCapita: Number(m[5]),
          builders: Number(m[6]),
        });
      }
    }
  }

  // Parse top universities table
  const unisSection = raw.match(
    /## Top 10 universities\n\n\| # \| University \| Country \| Students \| Per-capita \| Builders \|\n\|---\|---\|---\|---------:\|-----------:\|---------:\|\n([\s\S]*?)(?=## New metros|$)/,
  );
  const topUniversities: ReportDoc["topUniversities"] = [];
  if (unisSection) {
    for (const line of unisSection[1].split("\n")) {
      const m = line.match(
        /\| (\d+) \| (.+?) \| (.+?) \| ([\d,]+) \| ([\d.]+) \| (\d+) \|/,
      );
      if (m) {
        topUniversities.push({
          rank: Number(m[1]),
          university: m[2],
          country: m[3],
          students: Number(m[4].replace(/,/g, "")),
          perCapita: Number(m[5]),
          builders: Number(m[6]),
        });
      }
    }
  }

  // Parse biggest gainers/losers
  function parseCityDeltas(
    section: string | undefined,
  ): Array<{
    delta: string;
    city: string;
    newRank: string;
    perCapita: number;
  }> {
    const result: Array<{
      delta: string;
      city: string;
      newRank: string;
      perCapita: number;
    }> = [];
    if (!section) return result;
    for (const line of section.split("\n")) {
      const m = line.match(
        /\| ([\+\-±]\d+) \| (.+?), (.+?) \(#\d+\) \| (\d+) \| ([\d.]+) \|/,
      );
      if (m) {
        result.push({
          delta: m[1],
          city: `${m[2]}, ${m[3]}`,
          newRank: m[4],
          perCapita: Number(m[5]),
        });
      }
    }
    return result;
  }

  const gainersSection = raw.match(
    /## Biggest gainers\n\n([\s\S]*?)(?=## Biggest losers|$)/,
  );
  const losersSection = raw.match(/## Biggest losers\n\n([\s\S]*?)(?=##|$)/);

  return {
    _type: "report",
    _id: `report-${slug}`,
    title: titleMatch?.[1] ?? slug,
    slug: { _type: "slug", current: slug },
    periodStart: periodMatch?.[1] ?? "",
    periodEnd: periodMatch?.[2] ?? "",
    generatedAt: periodMatch?.[3] ?? "",
    tldrMetros: Number(tldrMatch?.[1] ?? 0),
    tldrUniversities: Number(tldrMatch?.[2] ?? 0),
    tldrActiveBuilders: Number(tldrMatch?.[3] ?? 0),
    tldrNewBuilders: Number(tldrMatch?.[4] ?? 0),
    tldrDeployments: Number(tldrMatch?.[5] ?? 0),
    tldrExecutions: Number(tldrMatch?.[6] ?? 0),
    tldrAmbassadors: Number(tldrMatch?.[7] ?? 0),
    topMetros,
    topUniversities,
    biggestGainers: parseCityDeltas(gainersSection?.[1]),
    biggestLosers: parseCityDeltas(losersSection?.[1]),
    body: raw,
  };
}

async function main() {
  const reportsDir = join(process.cwd(), "src/content/reports");
  let files: string[] = [];
  try {
    files = (await readdir(reportsDir))
      .filter((f) => f.endsWith(".md"))
      .sort()
      .reverse();
  } catch (err) {
    console.error(`Failed to read reports directory: ${err}`);
    process.exit(1);
  }

  console.log(`Found ${files.length} report files to migrate`);

  for (const file of files) {
    const slug = file.replace(/\.md$/, "");
    const path = join(reportsDir, file);
    const raw = await readFile(path, "utf-8");
    const doc = parseReport(raw, slug);

    try {
      await sanity.createOrReplace(doc);
      console.log(`✅ Migrated: ${slug}`);
    } catch (err) {
      console.error(`❌ Failed to migrate ${slug}: ${err}`);
    }
  }

  console.log(
    "\nMigration complete! Visit sanity.io/manage to review your content.",
  );
}

main();
