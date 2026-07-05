import type { Runtime } from '../types';

export interface LanguageServiceOptions {
  runtime: Runtime;
  code: string;
  onDiagnostics?: (diagnostics: LanguageDiagnostic[]) => void;
}

export interface LanguageDiagnostic {
  type: 'error' | 'warning' | 'info';
  message: string;
  line?: number;
  column?: number;
  endLine?: number;
  endColumn?: number;
  code?: string | number | { value: string; target: unknown };
  source?: string;
}

export interface ILanguageService {
  validate: (code: string) => Promise<LanguageDiagnostic[]>;
  dispose: () => void;
}

const FUNCTIONFLY_RUNTIME_TYPES = `
declare const Request: {
  new (input: RequestInfo | URL, init?: RequestInit): Request;
  prototype: Request;
};

declare const Response: {
  new (body?: BodyInit | null, init?: ResponseInit): Response;
  prototype: Response;
  json(data: unknown, init?: ResponseInit): Response;
  redirect(url: string, status?: number): Response;
  error(): Response;
};

declare const Headers: {
  prototype: Headers;
  new (init?: HeadersInit): Headers;
};

declare const URL: {
  prototype: URL;
  new (url: string | URL, base?: string | URL): URL;
};

declare const URLSearchParams: {
  prototype: URLSearchParams;
  new (init?: string | URLSearchParams | Record<string, string>): URLSearchParams;
};

interface ExecutionContext {
  waitUntil(promise: Promise<unknown>): void;
  passThroughOnException(): void;
}

interface Env {
  [key: string]: string | undefined;
}

declare const env: Env;
declare const ctx: ExecutionContext;
`;

const PYTHON_RUNTIME_TYPES = `
class Request:
    method: str
    url: str
    headers: dict
    body: str | None
    
    def json(self) -> dict: ...
    def text(self) -> str: ...

class Response:
    status: int
    headers: dict
    body: str
    
    @staticmethod
    def json(data: dict) -> Response: ...

class Env:
    def __getitem__(self, key: str) -> str: ...
    def get(self, key: str, default: str = "") -> str: ...

class Context:
    wait_until(promise: Any): ...
`;

export function createTypeScriptLanguageService(): ILanguageService {
  let disposed = false;

  return {
    async validate(code: string): Promise<LanguageDiagnostic[]> {
      if (disposed) return [];

      return new Promise((resolve) => {
        if (typeof window === 'undefined' || disposed) {
          resolve([]);
          return;
        }

        const diagnostics: LanguageDiagnostic[] = [];

        try {
          if (typeof window.monaco !== 'undefined') {
            const monaco = window.monaco;
            const languageId = 'typescript';

            const model = monaco.editor.createModel(code, languageId);
            // @ts-expect-error - Monaco typescript API not properly typed
            const worker = monaco.languages.typescript.getTypeScriptWorker();

            worker(model.uri).then((tsWorker) => {
              if (disposed) {
                model.dispose();
                resolve([]);
                return;
              }

              Promise.all([
                tsWorker.getEmitOutput(model.uri),
                tsWorker.getSemanticDiagnostics(model.uri.toString()),
                tsWorker.getSyntacticDiagnostics(model.uri.toString()),
              ])
                .then(([emitOutput, semanticDiagnostics, syntacticDiagnostics]) => {
                  if (disposed) {
                    model.dispose();
                    resolve([]);
                    return;
                  }

                  if (emitOutput.outputFiles.length === 0 && semanticDiagnostics.length === 0 && syntacticDiagnostics.length === 0) {
                    diagnostics.push({
                      type: 'info',
                      message: 'No errors found',
                      line: 1,
                    });
                  }

                  syntacticDiagnostics.forEach((d) => {
                    if (d.start !== undefined && d.length !== undefined) {
                      const start = model.getPositionAt(d.start);
                      diagnostics.push({
                        type: 'error',
                        message: d.messageText,
                        line: start.lineNumber,
                        column: start.column,
                        endLine: start.lineNumber,
                        endColumn: start.column + d.length,
                        source: 'ts',
                        code: d.code,
                      });
                    }
                  });

                  semanticDiagnostics.forEach((d) => {
                    if (d.start !== undefined && d.length !== undefined) {
                      const start = model.getPositionAt(d.start);
                      diagnostics.push({
                        type: 'error',
                        message: d.messageText,
                        line: start.lineNumber,
                        column: start.column,
                        endLine: start.lineNumber,
                        endColumn: start.column + d.length,
                        source: 'ts',
                        code: d.code,
                      });
                    }
                  });

                  model.dispose();
                  resolve(diagnostics);
                })
                .catch(() => {
                  model.dispose();
                  resolve(diagnostics);
                });
            });
          } else {
            resolve(diagnostics);
          }
        } catch {
          resolve(diagnostics);
        }
      });
    },

    dispose() {
      disposed = true;
    },
  };
}

export function createJavaScriptLanguageService(): ILanguageService {
  let disposed = false;

  return {
    async validate(code: string): Promise<LanguageDiagnostic[]> {
      if (disposed) return [];

      return new Promise((resolve) => {
        if (typeof window === 'undefined' || disposed) {
          resolve([]);
          return;
        }

        const diagnostics: LanguageDiagnostic[] = [];

        try {
          if (typeof window.monaco !== 'undefined') {
            const monaco = window.monaco;
            const languageId = 'javascript';

            const model = monaco.editor.createModel(code, languageId);

            // @ts-expect-error - Monaco javascript API not properly typed
            monaco.languages.typescript.getJavaScriptWorker().then((jsWorker) => {
              if (disposed) {
                model.dispose();
                resolve([]);
                return;
              }

              jsWorker.getEmitOutput(model.uri).then(() => {
                jsWorker.getSyntacticDiagnostics(model.uri.toString()).then((syntacticDiagnostics) => {
                  if (disposed) {
                    model.dispose();
                    resolve([]);
                    return;
                  }

                  syntacticDiagnostics.forEach((d) => {
                    if (d.start !== undefined && d.length !== undefined) {
                      const start = model.getPositionAt(d.start);
                      diagnostics.push({
                        type: 'error',
                        message: d.messageText,
                        line: start.lineNumber,
                        column: start.column,
                        endLine: start.lineNumber,
                        endColumn: start.column + d.length,
                        source: 'js',
                        code: d.code,
                      });
                    }
                  });

                  model.dispose();
                  resolve(diagnostics);
                }).catch(() => {
                  model.dispose();
                  resolve(diagnostics);
                });
              }).catch(() => {
                model.dispose();
                resolve(diagnostics);
              });
            });
          } else {
            resolve(diagnostics);
          }
        } catch {
          resolve(diagnostics);
        }
      });
    },

    dispose() {
      disposed = true;
    },
  };
}

export function createPythonValidationService(): ILanguageService {
  let disposed = false;
  let pyodideWorker: Worker | null = null;
  let workerPromise: Promise<Worker> | null = null;

  const initWorker = (): Promise<Worker> => {
    if (pyodideWorker) return Promise.resolve(pyodideWorker);
    if (workerPromise) return workerPromise;

    workerPromise = new Promise((resolve, reject) => {
      const workerCode = `
        let pyodide = null;
        let initPromise = null;

        const initPyodide = async () => {
          if (pyodide) return pyodide;
          if (initPromise) return initPromise;

          initPromise = new Promise(async (resolve, reject) => {
            try {
              importScripts('https://cdn.jsdelivr.net/pyodide/v0.25.1/full/pyodide.js');
              pyodide = await loadPyodide({
                indexURL: 'https://cdn.jsdelivr.net/pyodide/v0.25.1/full/',
              });
              resolve(pyodide);
            } catch (err) {
              reject(err);
            }
          });

          return initPromise;
        };

        self.onmessage = async function(e) {
          const { code, id } = e.data;

          try {
            await initPyodide();

            if (!pyodide) {
              self.postMessage({ id, error: 'Pyodide not loaded', diagnostics: [] });
              return;
            }

            const diagnostics = [];

            try {
              pyodide.runPython(\`
import ast
import sys
from io import StringIO

class SyntaxErrorCollector(ast.NodeVisitor):
    def __init__(self):
        self.errors = []

    def visit_Import(self, node):
        for alias in node.names:
            pass
        self.generic_visit(node)

    def visit_ImportFrom(self, node):
        for alias in node.names:
            pass
        self.generic_visit(node)

    def visit_FunctionDef(self, node):
        for arg in node.args.args:
            pass
        self.generic_visit(node)

    def visit_AsyncFunctionDef(self, node):
        for arg in node.args.args:
            pass
        self.generic_visit(node)
\`);

              const result = pyodide.runPython(\`
import ast

errors = []

try:
    tree = ast.parse(\` + '"""' + code + '"""' + \`)
    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef):
            if node.decorator_list:
                for dec in node.decorator_list:
                    pass
        elif isinstance(node, ast.Import):
            pass
        elif isinstance(node, ast.ImportFrom):
            pass
except SyntaxError as e:
    errors.append({
        'line': e.lineno or 1,
        'column': e.offset or 1,
        'message': str(e)
    })
except Exception as e:
    errors.append({
        'line': 1,
        'column': 1,
        'message': f'Parse error: {str(e)}'
    })

errors
\`);

              if (result && result.length !== undefined) {
                for (let i = 0; i < result.length; i++) {
                  const err = result.get(i);
                  if (err) {
                    diagnostics.push({
                      type: 'error',
                      line: err.line || 1,
                      column: err.column || 1,
                      message: err.message || String(err),
                    });
                  }
                }
              }

              pyodide.runPython('result = None');

            } catch (parseErr) {
              const errMsg = parseErr.message || String(parseErr);
              diagnostics.push({
                type: 'error',
                line: 1,
                column: 1,
                message: 'Python syntax error: ' + errMsg,
              });
            }

            self.postMessage({ id, diagnostics, error: null });
          } catch (err) {
            self.postMessage({
              id,
              error: String(err),
              diagnostics: [{
                type: 'error',
                line: 1,
                column: 1,
                message: 'Python validation unavailable: ' + String(err),
              }]
            });
          }
        };
      `;

      const blob = new Blob([workerCode], { type: 'application/javascript' });
      const worker = new Worker(URL.createObjectURL(blob));

      worker.onerror = (e) => {
        console.error('Python worker error:', e);
        reject(e);
      };

      pyodideWorker = worker;
      resolve(worker);
    });

    return workerPromise;
  };

  let requestId = 0;

  return {
    async validate(code: string): Promise<LanguageDiagnostic[]> {
      if (disposed) return [];

      return new Promise((resolve) => {
        const id = ++requestId;

        initWorker()
          .then((worker) => {
            const timeout = setTimeout(() => {
              resolve([]);
            }, 10000);

            const handler = (e: MessageEvent) => {
              if (e.data.id === id) {
                clearTimeout(timeout);
                worker.removeEventListener('message', handler);
                resolve(e.data.diagnostics || []);
              }
            };

            worker.addEventListener('message', handler);
            worker.postMessage({ code, id });
          })
          .catch(() => {
            resolve([]);
          });
      });
    },

    dispose() {
      disposed = true;
      if (pyodideWorker) {
        pyodideWorker.terminate();
        pyodideWorker = null;
      }
      workerPromise = null;
    },
  };
}

export function createGoValidationService(): ILanguageService {
  let disposed = false;

  return {
    async validate(code: string): Promise<LanguageDiagnostic[]> {
      if (disposed) return [];

      return new Promise((resolve) => {
        const diagnostics: LanguageDiagnostic[] = [];

        const lines = code.split('\n');
        let inString = false;
        let stringChar = '';
        let parenDepth = 0;
        let braceDepth = 0;
        let bracketDepth = 0;

        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];
          const trimmed = line.trim();

          if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
            continue;
          }

          for (let j = 0; j < line.length; j++) {
            const char = line[j];
            const prevChar = j > 0 ? line[j - 1] : '';

            if (prevChar !== '\\' && (char === '"' || char === '\'' || char === '`')) {
              if (!inString) {
                inString = true;
                stringChar = char;
              } else if (char === stringChar) {
                inString = false;
              }
            } else if (!inString) {
              if (char === '(') parenDepth++;
              else if (char === ')') parenDepth--;
              else if (char === '{') braceDepth++;
              else if (char === '}') braceDepth--;
              else if (char === '[') bracketDepth++;
              else if (char === ']') bracketDepth--;
            }
          }

          if (inString && i === lines.length - 1) {
            diagnostics.push({
              type: 'error',
              message: 'unterminated string literal',
              line: i + 1,
              source: 'go',
            });
          }
        }

        if (parenDepth !== 0) {
          diagnostics.push({
            type: 'error',
            message: parenDepth > 0 ? 'unclosed parenthesis' : 'unexpected closing parenthesis',
            line: lines.length,
            source: 'go',
          });
        }

        if (braceDepth !== 0) {
          diagnostics.push({
            type: 'error',
            message: braceDepth > 0 ? 'unclosed brace' : 'unexpected closing brace',
            line: lines.length,
            source: 'go',
          });
        }

        if (bracketDepth !== 0) {
          diagnostics.push({
            type: 'error',
            message: bracketDepth > 0 ? 'unclosed bracket' : 'unexpected closing bracket',
            line: lines.length,
            source: 'go',
          });
        }

        if (!code.includes('package ')) {
          diagnostics.push({
            type: 'error',
            message: 'expected package declaration',
            line: 1,
            source: 'go',
          });
        }

        resolve(diagnostics);
      });
    },

    dispose() {
      disposed = true;
    },
  };
}

export function createRustValidationService(): ILanguageService {
  let disposed = false;

  return {
    async validate(code: string): Promise<LanguageDiagnostic[]> {
      if (disposed) return [];

      return new Promise((resolve) => {
        const diagnostics: LanguageDiagnostic[] = [];

        const lines = code.split('\n');
        let inComment = false;
        let inString = false;
        let stringChar = '';

        for (let i = 0; i < lines.length; i++) {
          const line = lines[i];

          for (let j = 0; j < line.length; j++) {
            const char = line[j];
            const prevChar = j > 0 ? line[j - 1] : '';
            const prevPrevChar = j > 1 ? line[j - 2] : '';

            if (!inString) {
              if (prevChar === '/' && char === '/' && prevPrevChar !== '/') {
                break;
              }

              if (prevChar === '/' && char === '*' && prevPrevChar !== '*') {
                inComment = true;
                continue;
              }

              if (prevChar === '*' && char === '/' && prevPrevChar !== '/') {
                inComment = false;
                continue;
              }

              if (inComment) continue;
            }

            if (!inComment && (char === '"' || char === '\'')) {
              if (!inString) {
                inString = true;
                stringChar = char;
              } else if (char === stringChar && prevChar !== '\\') {
                inString = false;
              }
            }
          }

          if (inString && i === lines.length - 1) {
            diagnostics.push({
              type: 'error',
              message: 'unterminated string literal',
              line: i + 1,
              source: 'rust',
            });
          }
        }

        const fnMatch = code.match(/\bfn\s+\w+/);
        if (!fnMatch) {
          diagnostics.push({
            type: 'warning',
            message: 'no function declarations found (use fn keyword)',
            line: 1,
            source: 'rust',
          });
        }

        resolve(diagnostics);
      });
    },

    dispose() {
      disposed = true;
    },
  };
}

export function createLanguageService(runtime: Runtime): ILanguageService {
  switch (runtime) {
    case 'typescript':
    case 'deno':
    case 'bun':
      return createTypeScriptLanguageService();
    case 'javascript':
      return createJavaScriptLanguageService();
    case 'python':
    case 'python-wasm':
      return createPythonValidationService();
    case 'go':
      return createGoValidationService();
    case 'rust-wasm':
    case 'browser-wasm':
      return createRustValidationService();
    default:
      return {
        async validate() {
          return [];
        },
        dispose() {},
      };
  }
}

export function getMonacoDiagnostics(
  monaco: typeof import('monaco-editor'),
  code: string,
  language: string
): LanguageDiagnostic[] {
  const diagnostics: LanguageDiagnostic[] = [];

  try {
    const model = monaco.editor.createModel(code, language);

    // @ts-expect-error - Monaco MarkerSeverity not properly recognized
    const getDiagnostics = (severity: monaco.MarkerSeverity) => {
      const markers = monaco.editor.getModelMarkers({ resource: model.uri });
      return markers.filter((m) => m.severity === severity);
    };

    getDiagnostics(monaco.MarkerSeverity.Error).forEach((marker) => {
      diagnostics.push({
        type: 'error',
        message: marker.message,
        line: marker.startLineNumber,
        column: marker.startColumn,
        endLine: marker.endLineNumber,
        endColumn: marker.endColumn,
        source: marker.source,
        code: typeof marker.code === 'object' ? String((marker.code as { value: string }).value) : marker.code,
      });
    });

    getDiagnostics(monaco.MarkerSeverity.Warning).forEach((marker) => {
      diagnostics.push({
        type: 'warning',
        message: marker.message,
        line: marker.startLineNumber,
        column: marker.startColumn,
        endLine: marker.endLineNumber,
        endColumn: marker.endColumn,
        source: marker.source,
        code: typeof marker.code === 'object' ? String((marker.code as { value: string }).value) : marker.code,
      });
    });

    getDiagnostics(monaco.MarkerSeverity.Info).forEach((marker) => {
      diagnostics.push({
        type: 'info',
        message: marker.message,
        line: marker.startLineNumber,
        column: marker.startColumn,
        endLine: marker.endLineNumber,
        endColumn: marker.endColumn,
        source: marker.source,
        code: typeof marker.code === 'object' ? String((marker.code as { value: string }).value) : marker.code,
      });
    });

    model.dispose();
  } catch {
  }

  return diagnostics;
}

declare global {
  interface Window {
    monaco: typeof import('monaco-editor');
  }
}
