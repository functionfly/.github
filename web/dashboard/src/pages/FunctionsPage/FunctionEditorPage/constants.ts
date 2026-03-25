import type { HttpMethod, Runtime } from './types';

export const RUNTIME_META: Record<
  Runtime,
  { label: string; color: string; description: string; monacoLang: string }
> = {
  typescript: {
    label: 'TypeScript',
    color: '#3178c6',
    description: 'V8-powered, edge-native, full TypeScript support',
    monacoLang: 'typescript',
  },
  python: {
    label: 'Python',
    color: '#3572A5',
    description: 'CPython 3.12, batteries included',
    monacoLang: 'python',
  },
  'rust-wasm': {
    label: 'Rust / WASM',
    color: '#dea584',
    description: 'Compile to WebAssembly for maximum performance',
    monacoLang: 'rust',
  },
};

export const CODE_TEMPLATES: Record<Runtime, string> = {
  typescript: `export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);

    return Response.json({
      message: 'Hello from FunctionFly!',
      path: url.pathname,
      method: request.method,
      timestamp: new Date().toISOString(),
    });
  },
};`,
  python: `import json

def handler(request, env, ctx):
    """FunctionFly Python handler."""
    return {
        "status": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({
            "message": "Hello from FunctionFly!",
            "method": request.get("method", "GET"),
        }),
    }`,
  'rust-wasm': `use functionfly_sdk::{Request, Response, Env};

#[no_mangle]
pub extern "C" fn handle(req: Request, env: Env) -> Response {
    Response::json(serde_json::json!({
        "message": "Hello from FunctionFly (Rust/WASM)!",
        "method": req.method(),
    }))
}`,
};

export const RUNTIME_VERSIONS: Record<Runtime, string[]> = {
  typescript: ['ES2022', 'ES2021', 'ES2020'],
  python: ['3.12', '3.11', '3.10'],
  'rust-wasm': ['1.75', '1.74', '1.73'],
};

export const HTTP_METHODS: HttpMethod[] = ['ANY', 'GET', 'POST', 'PUT', 'PATCH', 'DELETE'];
export const MEMORY_OPTIONS = [64, 128, 256, 512, 1024, 2048];
export const TIMEOUT_OPTIONS = [1000, 5000, 10000, 30000, 60000, 300000];
export const DRAFT_KEY = 'functionfly:new-function-draft';
