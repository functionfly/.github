import { createClient, type SanityClient } from "@sanity/client";

let _client: SanityClient | null = null;

function getClient(): SanityClient | null {
  const projectId = import.meta.env.PUBLIC_SANITY_PROJECT_ID;
  if (!projectId) return null;
  if (!_client) {
    _client = createClient({
      projectId,
      dataset: import.meta.env.PUBLIC_SANITY_DATASET || "production",
      apiVersion: "2024-01-01",
      useCdn: true,
    });
  }
  return _client;
}

export interface ReportSummary {
  _id: string;
  title: string;
  slug: string;
  periodStart: string;
  periodEnd: string;
  generatedAt: string;
  tldrMetros: number;
  tldrUniversities: number;
  tldrActiveBuilders: number;
  tldrNewBuilders: number;
  tldrAmbassadors: number;
}

export interface BlogPost {
  _id: string;
  title: string;
  slug: string;
  author: { name: string } | null;
  category: { title: string } | null;
  publishedAt: string;
  description: string;
  body: string;
  tags: string[];
}

export interface Report extends ReportSummary {
  tldrDeployments: number;
  tldrExecutions: number;
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
  newMetros: any[];
  newAmbassadors: any[];
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

export async function getReports(): Promise<ReportSummary[]> {
  const client = getClient();
  if (!client) {
    console.warn(
      "[sanity] PUBLIC_SANITY_PROJECT_ID not set, returning empty array",
    );
    return [];
  }
  return client.fetch<ReportSummary[]>(`
    *[_type == "report"] | order(generatedAt desc) {
      _id,
      title,
      "slug": slug.current,
      periodStart,
      periodEnd,
      generatedAt,
      tldrMetros,
      tldrUniversities,
      tldrActiveBuilders,
      tldrNewBuilders,
      tldrAmbassadors
    }
  `);
}

export async function getReport(slug: string): Promise<Report | null> {
  const client = getClient();
  if (!client) {
    console.warn("[sanity] PUBLIC_SANITY_PROJECT_ID not set");
    return null;
  }
  return client.fetch<Report | null>(
    `
    *[_type == "report" && slug.current == $slug][0] {
      _id,
      title,
      "slug": slug.current,
      periodStart,
      periodEnd,
      generatedAt,
      tldrMetros,
      tldrUniversities,
      tldrActiveBuilders,
      tldrNewBuilders,
      tldrDeployments,
      tldrExecutions,
      tldrAmbassadors,
      topMetros,
      topUniversities,
      newMetros,
      newAmbassadors,
      biggestGainers,
      biggestLosers,
      body
    }
  `,
    { slug },
  );
}
