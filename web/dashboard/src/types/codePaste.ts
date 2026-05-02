export interface Parameter {
  name: string;
  type?: string;
  has_default: boolean;
  default_value?: string;
}

export interface ParsedFunction {
  id: string;
  name: string;
  language: string;
  signature: string;
  parameters: Parameter[];
  return_type?: string;
  docstring?: string;
  code: string;
  start_line: number;
  end_line: number;
}

export interface ParseResult {
  language: string;
  confidence: number;
  functions: ParsedFunction[];
  raw_code_length: number;
}

export interface ParseCodeRequest {
  code: string;
  force_language?: string;
}

export interface CreateFunctionInput {
  name: string;
  code: string;
  language: string;
}

export interface CreateFunctionsRequest {
  functions: CreateFunctionInput[];
  visibility: 'private' | 'public';
  providers?: string[];
  region?: string;
  author?: string;
  changelog?: string;
}

export interface CreatedFunction {
  id: string;
  name: string;
  status: string;
}

export interface FailedFunction {
  name: string;
  error: string;
}

export interface CreateFromCodeResponse {
  created: CreatedFunction[];
  failed?: FailedFunction[];
}

export interface ImportConfig {
  visibility: 'private' | 'public';
  providers: string[];
  region: string;
  author?: string;
  changelog?: string;
}

export type AnalyzerStatus =
  | 'idle'
  | 'typing'
  | 'parsing'
  | 'parsed'
  | 'importing'
  | 'success'
  | 'error';

export interface CodePasteState {
  code: string;
  language: string | null;
  confidence: number;
  functions: ParsedFunction[];
  selectedIds: Set<string>;
  importConfig: ImportConfig;
  status: AnalyzerStatus;
  error: string | null;
}

export const SUPPORTED_LANGUAGES = {
  python: 'Python',
  javascript: 'JavaScript',
  typescript: 'TypeScript',
  go: 'Go',
  rust: 'Rust',
  ruby: 'Ruby',
  java: 'Java',
  kotlin: 'Kotlin',
  swift: 'Swift',
  cpp: 'C++',
  c: 'C',
} as const;

export type SupportedLanguage = keyof typeof SUPPORTED_LANGUAGES;