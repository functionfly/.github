import type { Runtime } from '../types';
import { createLanguageService } from './LanguageService';
import type { LanguageDiagnostic, ILanguageService } from './LanguageService';

export interface ValidationIssue {
  type: 'error' | 'warning' | 'info';
  message: string;
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  source?: string;
  code?: string | number;
}

export interface ValidationResult {
  valid: boolean;
  issues: ValidationIssue[];
  hasRealValidation: boolean;
}

const languageServiceCache: Map<Runtime, ILanguageService> = new Map();

function getLanguageService(runtime: Runtime): ILanguageService {
  if (!languageServiceCache.has(runtime)) {
    languageServiceCache.set(runtime, createLanguageService(runtime));
  }
  return languageServiceCache.get(runtime)!;
}

export function invalidateLanguageService(runtime?: Runtime): void {
  if (runtime) {
    const service = languageServiceCache.get(runtime);
    if (service) {
      service.dispose();
      languageServiceCache.delete(runtime);
    }
  } else {
    languageServiceCache.forEach((service) => service.dispose());
    languageServiceCache.clear();
  }
}

export async function validateCodeWithLanguageService(
  code: string,
  runtime: Runtime
): Promise<{ diagnostics: ValidationIssue[]; hasRealValidation: boolean }> {
  if (!code.trim()) {
    return {
      diagnostics: [{ type: 'error', message: 'Code is empty' }],
      hasRealValidation: true,
    };
  }

  const service = getLanguageService(runtime);
  const diagnostics = await service.validate(code);

  return {
    diagnostics: diagnostics.map((d: LanguageDiagnostic) => ({
      type: d.type,
      message: d.message,
      line: d.line,
      column: d.column,
      endLine: d.endLine,
      endColumn: d.endColumn,
      source: d.source,
      code: typeof d.code === 'object' ? String((d.code as { value: string }).value) : d.code,
    })),
    hasRealValidation: true,
  };
}

export async function validateCodeAsync(code: string, runtime: Runtime): Promise<ValidationResult> {
  const { diagnostics, hasRealValidation } = await validateCodeWithLanguageService(code, runtime);
  const issues = [...diagnostics];

  const supplementaryIssues = getSupplementaryValidation(code, runtime);
  issues.push(...supplementaryIssues);

  const realErrors = issues.filter((i) => i.type === 'error' && i.source !== 'heuristic');
  const allErrors = issues.filter((i) => i.type === 'error');

  return {
    valid: realErrors.length === 0 && allErrors.length === 0,
    issues,
    hasRealValidation,
  };
}

export function validateCode(code: string, runtime: Runtime): ValidationResult {
  return validateCodeSync(code, runtime);
}

export function validateCodeSync(code: string, runtime: Runtime): ValidationResult {
  const issues: ValidationIssue[] = [];

  if (!code.trim()) {
    return {
      valid: false,
      issues: [{ type: 'error', message: 'Code is empty', source: 'heuristic' }],
      hasRealValidation: false,
    };
  }

  if (runtime === 'go' && !code.includes('package ')) {
    issues.push({
      type: 'error',
      message: 'Go code must have a package declaration (e.g., package main).',
      source: 'heuristic',
    });
  }

  issues.push(...getSupplementaryValidation(code, runtime));

  return {
    valid: !issues.some((i) => i.type === 'error'),
    issues,
    hasRealValidation: false,
  };
}

function getSupplementaryValidation(code: string, runtime: Runtime): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  if (code.length > 100000) {
    issues.push({
      type: 'warning',
      message: 'Code is very large (>100KB). Consider splitting into smaller functions.',
      source: 'heuristic',
    });
  }

  const hasExport = checkForExport(code, runtime);
  if (!hasExport) {
    issues.push({
      type: 'warning',
      message: 'No exported handler detected. Make sure you export a handler function.',
      source: 'heuristic',
    });
  }

  if (code.includes('eval(') || code.includes('eval (')) {
    issues.push({
      type: 'warning',
      message: 'Use of eval() detected. This can be a security risk.',
      source: 'heuristic',
    });
  }

  if (code.includes('Function(') && !code.includes('Function.prototype')) {
    issues.push({
      type: 'warning',
      message: 'Dynamic code execution with Function() constructor. Review for security.',
      source: 'heuristic',
    });
  }

  if (code.includes('innerHTML') && !code.includes('sanitize')) {
    issues.push({
      type: 'warning',
      message: 'Potential XSS vulnerability: innerHTML usage without sanitization.',
      source: 'heuristic',
    });
  }

  if (code.includes('dangerouslySetInnerHTML')) {
    issues.push({
      type: 'warning',
      message: 'Potential XSS: dangerouslySetInnerHTML usage detected.',
      source: 'heuristic',
    });
  }

  if (code.includes('atob(') || code.includes('btoa(')) {
    issues.push({
      type: 'info',
      message: 'Base64 encoding/decoding detected. Ensure data is properly validated.',
      source: 'heuristic',
    });
  }

  if (code.includes('setTimeout') && code.includes('setTimeout(') && /\d+\s*===\s*\d+/.test(code)) {
    issues.push({
      type: 'info',
      message: 'Timing attack patterns detected. Use constant-time comparisons.',
      source: 'heuristic',
    });
  }

  if (runtime === 'python' || runtime === 'python-light' || runtime === 'python-wasm') {
    if (code.includes('os.system') || code.includes('subprocess')) {
      issues.push({
        type: 'warning',
        message: 'Shell command execution detected. Ensure input is sanitized.',
        source: 'heuristic',
      });
    }

    if (code.includes('pickle.load') || code.includes('pickle.loads')) {
      issues.push({
        type: 'warning',
        message: 'Pickle deserialization can be dangerous with untrusted data.',
        source: 'heuristic',
      });
    }

    if (code.includes('eval(') || code.includes('exec(')) {
      issues.push({
        type: 'warning',
        message: 'Dynamic code execution (eval/exec) detected. This is a security risk.',
        source: 'heuristic',
      });
    }

    if (code.includes('__import__') || code.includes('importlib')) {
      issues.push({
        type: 'info',
        message: 'Dynamic import detected. Ensure the module path is validated.',
        source: 'heuristic',
      });
    }
  }

  if (runtime === 'go') {
    if (code.match(/exec\.Command/)) {
      issues.push({
        type: 'warning',
        message: 'Command execution detected. Sanitize all user inputs.',
        source: 'heuristic',
      });
    }

    if (code.match(/os\.Open\(/)) {
      issues.push({
        type: 'info',
        message: 'File operations detected. Validate file paths to prevent path traversal.',
        source: 'heuristic',
      });
    }
  }

  if (runtime === 'typescript' || runtime === 'javascript' || runtime === 'deno' || runtime === 'bun') {
    if (code.match(/\.innerHTML\s*=/)) {
      issues.push({
        type: 'warning',
        message: 'DOM XSS: innerHTML assignment detected. Consider using textContent or sanitized HTML.',
        source: 'heuristic',
      });
    }

    if (code.match(/new\s+Function/)) {
      issues.push({
        type: 'warning',
        message: 'Dynamic Function constructor detected. Consider alternatives.',
        source: 'heuristic',
      });
    }
  }

  if (runtime === 'go') {
    const imports = code.match(/import\s+\(\s*([^)]+)\)/);
    if (imports) {
      const importList = imports[1].split('\n').map((i) => i.trim().replace(/"/g, ''));
      for (const imp of importList) {
        const shortName = imp.split('/').pop() || imp;
        if (shortName && !new RegExp(`\\b${shortName}\\.`).test(code) && !code.includes('// ' + shortName)) {
          issues.push({
            type: 'info',
            message: `Import '${imp}' may be unused.`,
            source: 'heuristic',
          });
        }
      }
    }
  }

  return issues;
}

function checkForExport(code: string, runtime: Runtime): boolean {
  switch (runtime) {
    case 'typescript':
    case 'javascript':
    case 'deno':
    case 'bun':
      return (
        code.includes('export default') ||
        code.includes('export function') ||
        code.includes('export async function') ||
        code.includes('export const') ||
        code.includes('export class') ||
        code.includes('module.exports')
      );
    case 'python':
    case 'python-wasm':
      return code.includes('def handler') || code.includes('def handle') || code.includes('async def handler');
    case 'go':
      return code.includes('func Handler') || code.includes('func handler') || code.includes('func (');
    case 'rust-wasm':
    case 'browser-wasm':
      return code.includes('#[no_mangle]') || code.includes('pub extern "C"');
    default:
      return true;
  }
}

export function getQuickFixes(issue: ValidationIssue, code: string, runtime: Runtime): string | null {
  if (issue.source !== 'heuristic' && issue.source) {
    return null;
  }

  if (issue.message.includes('async but has no await')) {
    return 'Consider removing async keyword or adding an await statement.';
  }

  if (issue.message.includes('missing return statement')) {
    return 'Add return Response.json({ ... }) at the end of your handler.';
  }

  if (issue.message.includes('No exported handler')) {
    switch (runtime) {
      case 'typescript':
      case 'javascript':
      case 'deno':
      case 'bun':
        return 'Export a handler using: export default { async fetch(request, env, ctx) { ... } }';
      case 'python':
        return 'Define a handler: def handler(request, env, ctx): ...';
      case 'go':
        return 'Define a handler: func Handler(w http.ResponseWriter, r *http.Request) { ... }';
      case 'rust-wasm':
      case 'browser-wasm':
        return 'Use #[no_mangle] and extern "C" fn handle(req: Request, env: Env) -> Response';
    }
  }

  return null;
}

export function getMonacoLanguage(runtime: Runtime): string {
  switch (runtime) {
    case 'typescript':
    case 'deno':
    case 'bun':
      return 'typescript';
    case 'javascript':
      return 'javascript';
    case 'python':
    case 'python-wasm':
      return 'python';
    case 'go':
      return 'go';
    case 'rust-wasm':
    case 'browser-wasm':
      return 'rust';
    default:
      return 'plaintext';
  }
}
