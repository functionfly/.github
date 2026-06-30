import { createClient, type SanityClient } from "@sanity/client";

let _client: SanityClient | null = null;

export function getClient(): SanityClient | null {
  const projectId = import.meta.env.PUBLIC_SANITY_PROJECT_ID;
  if (!projectId) return null;
  if (!_client) {
    _client = createClient({
      projectId,
      dataset: import.meta.env.PUBLIC_SANITY_DATASET || "production",
      apiVersion: "2024-01-01",
      useCdn: false,
    });
  }
  return _client;
}
