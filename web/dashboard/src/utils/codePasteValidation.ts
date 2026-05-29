export const MAX_CODE_BYTES = 102_400;
export const MAX_FUNCTIONS_PER_IMPORT = 50;

export const FUNCTION_NAME_REGEX = /^[a-zA-Z][a-zA-Z0-9_-]{0,99}$/;

export const ALLOWED_PROVIDERS = ['cloud', 'edge', 'local'] as const;

export const ALLOWED_REGIONS = [
  'us-east-1',
  'us-west-2',
  'eu-west-1',
  'eu-central-1',
  'ap-southeast-1',
  'ap-northeast-1',
] as const;

export function getCodeByteLength(code: string): number {
  return new TextEncoder().encode(code).length;
}

export function formatCodeSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  return `${(bytes / 1024).toFixed(1)} KB`;
}

export function validateCodeSize(code: string): string | null {
  if (!code.trim()) {
    return 'Please enter some code to parse';
  }
  if (getCodeByteLength(code) > MAX_CODE_BYTES) {
    return `Code exceeds maximum size of ${Math.floor(MAX_CODE_BYTES / 1024)}KB`;
  }
  return null;
}

export function validateFunctionName(name: string): string | null {
  const trimmed = name.trim();
  if (!trimmed) {
    return 'Function name is required';
  }
  if (trimmed.length > 100) {
    return 'Function name must be 100 characters or less';
  }
  if (!FUNCTION_NAME_REGEX.test(trimmed)) {
    return 'Function name must start with a letter and contain only letters, numbers, underscores, and hyphens';
  }
  return null;
}

export function validateImportConfig(config: {
  providers: string[];
  region: string;
}): string | null {
  if (config.providers.length === 0) {
    return 'Select at least one provider';
  }
  const invalidProvider = config.providers.find(
    (provider) => !ALLOWED_PROVIDERS.includes(provider as (typeof ALLOWED_PROVIDERS)[number])
  );
  if (invalidProvider) {
    return `Invalid provider: ${invalidProvider}`;
  }
  if (!ALLOWED_REGIONS.includes(config.region as (typeof ALLOWED_REGIONS)[number])) {
    return `Invalid region: ${config.region}`;
  }
  return null;
}
