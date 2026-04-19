import type { Runtime } from '../types';

export interface ValidationIssue {
  type: 'error' | 'warning' | 'info';
  message: string;
  line?: number;
  column?: number;
}

export interface ValidationResult {
  valid: boolean;
  issues: ValidationIssue[];
}

export function validateCode(code: string, runtime: Runtime): ValidationResult {
  const issues: ValidationIssue[] = [];

  // Basic syntax checks
  if (!code.trim()) {
    return { valid: false, issues: [{ type: 'error', message: 'Code is empty' }] };
  }

  // Check for common issues by runtime
  switch (runtime) {
    case 'typescript':
    case 'javascript':
    case 'deno':
    case 'bun':
      issues.push(...validateJavaScript(code));
      break;
    case 'python':
    case 'python-wasm':
      issues.push(...validatePython(code));
      break;
    case 'go':
      issues.push(...validateGo(code));
      break;
    case 'rust-wasm':
    case 'browser-wasm':
      issues.push(...validateRust(code));
      break;
  }

  // Check for common patterns across all runtimes
  if (code.length > 100000) {
    issues.push({
      type: 'warning',
      message: 'Code is very large (>100KB). Consider splitting into smaller functions.',
    });
  }

  // Check for handler/export patterns
  const hasExport =
    code.includes('export default') ||
    code.includes('export function') ||
    code.includes('export async function') ||
    code.includes('module.exports') ||
    code.includes('def handler') ||
    code.includes('func Handler') ||
    code.includes('#[no_mangle]');

  if (!hasExport) {
    issues.push({
      type: 'warning',
      message: 'No exported handler detected. Make sure you export a handler function.',
    });
  }

  // Check for potential security issues
  if (code.includes('eval(')) {
    issues.push({
      type: 'warning',
      message: 'Use of eval() detected. This can be a security risk.',
    });
  }

  if (code.includes('Function(') && !code.includes('Function.prototype')) {
    issues.push({
      type: 'warning',
      message: 'Dynamic code execution with Function() constructor. Review for security.',
    });
  }

  return {
    valid: !issues.some((i) => i.type === 'error'),
    issues,
  };
}

function validateJavaScript(code: string): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  // Check for TypeScript-specific issues
  if (code.includes('any') && !code.includes('unknown')) {
    const anyCount = (code.match(/:\s*any\b/g) || []).length;
    if (anyCount > 3) {
      issues.push({
        type: 'info',
        message: `Consider using more specific types instead of 'any' (${anyCount} occurrences)`,
      });
    }
  }

  // Check for async/await issues
  const asyncMatches = code.match(/async\s+\w+/g) || [];
  const awaitMatches = code.match(/await\s+/g) || [];
  if (asyncMatches.length > 0 && awaitMatches.length === 0) {
    issues.push({
      type: 'warning',
      message: 'Function is declared async but has no await statements.',
    });
  }

  // Check for missing return statements in handler
  if (code.includes('export default') && !code.includes('return')) {
    issues.push({
      type: 'info',
      message: 'Handler may be missing return statement. Ensure you return a Response.',
    });
  }

  // Check for proper fetch usage
  if (code.includes('fetch(') && !code.includes('await fetch')) {
    issues.push({
      type: 'warning',
      message: 'fetch() call should probably be awaited.',
    });
  }

  return issues;
}

function validatePython(code: string): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  // Check for proper indentation (basic check)
  const lines = code.split('\n');
  let hasIndentation = false;
  for (const line of lines) {
    if (line.startsWith('  ') || line.startsWith('\t')) {
      hasIndentation = true;
      break;
    }
  }
  if (!hasIndentation && lines.length > 5) {
    issues.push({
      type: 'warning',
      message: 'Python code may have indentation issues.',
    });
  }

  // Check for import statements
  if (!code.includes('import') && code.includes('json') && !code.includes('json.')) {
    issues.push({
      type: 'warning',
      message: "Using 'json' but no import statement found.",
    });
  }

  return issues;
}

function validateGo(code: string): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  // Check for package declaration
  if (!code.includes('package ')) {
    issues.push({
      type: 'error',
      message: 'Go code must have a package declaration (e.g., package main).',
    });
  }

  // Check for function signature
  if (!code.includes('func ')) {
    issues.push({
      type: 'warning',
      message: 'No function declarations found.',
    });
  }

  // Check for imported but potentially unused packages
  const imports = code.match(/import\s+\(\s*([^)]+)\)/);
  if (imports) {
    const importList = imports[1].split('\n').map((i) => i.trim().replace(/"/g, ''));
    for (const imp of importList) {
      const shortName = imp.split('/').pop() || imp;
      const usage = new RegExp(`\\b${shortName}\\.`).test(code);
      if (!usage && shortName && !code.includes('// ' + shortName)) {
        issues.push({
          type: 'info',
          message: `Import '${imp}' may be unused.`,
        });
      }
    }
  }

  return issues;
}

function validateRust(code: string): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  // Check for no_mangle
  if (!code.includes('#[no_mangle]')) {
    issues.push({
      type: 'warning',
      message: "WASM functions should use #[no_mangle] to preserve function names.",
    });
  }

  // Check for extern C
  if (!code.includes('extern "C"')) {
    issues.push({
      type: 'warning',
      message: 'WASM export functions should use extern "C" for C ABI compatibility.',
    });
  }

  return issues;
}

export function getQuickFixes(issue: ValidationIssue, code: string, runtime: Runtime): string | null {
  // Return quick fix suggestions for common issues
  if (issue.message.includes('async but has no await')) {
    return 'Consider removing async keyword or adding an await statement.';
  }

  if (issue.message.includes('missing return statement')) {
    return 'Add return Response.json({ ... }) at the end of your handler.';
  }

  return null;
}
