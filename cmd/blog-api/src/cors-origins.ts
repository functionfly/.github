/** Default browser origins allowed for the blog API (merged with CORS_ORIGIN env). */
export const DEFAULT_CORS_ORIGINS = [
  'http://localhost:5173',
  'http://localhost:3000',
  'http://localhost:4321',
  'https://functionfly.com',
  'https://www.functionfly.com',
];

export function mergeCorsOrigins(envCsv: string | undefined): string[] {
  const extra =
    envCsv
      ?.split(',')
      .map((o) => o.trim())
      .filter(Boolean) ?? [];
  return [...new Set([...DEFAULT_CORS_ORIGINS, ...extra])];
}
