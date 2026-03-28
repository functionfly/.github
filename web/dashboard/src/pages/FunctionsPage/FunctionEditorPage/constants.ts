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
  javascript: {
    label: 'JavaScript',
    color: '#f7df1e',
    description: 'Pure JavaScript with modern ES2022 features',
    monacoLang: 'javascript',
  },
  python: {
    label: 'Python',
    color: '#3572A5',
    description: 'CPython 3.12, batteries included',
    monacoLang: 'python',
  },
  'python-wasm': {
    label: 'MicroPython',
    color: '#2b5b84',
    description: 'Lightweight Python for embedded and edge devices',
    monacoLang: 'python',
  },
  'rust-wasm': {
    label: 'Rust / WASM',
    color: '#dea584',
    description: 'Compile to WebAssembly for maximum performance',
    monacoLang: 'rust',
  },
  go: {
    label: 'Go',
    color: '#00add8',
    description: 'Fast, reliable, and efficient cloud-native runtime',
    monacoLang: 'go',
  },
  deno: {
    label: 'Deno',
    color: '#000000',
    description: 'Secure runtime for JavaScript and TypeScript',
    monacoLang: 'typescript',
  },
  bun: {
    label: 'Bun',
    color: '#f9f1e5',
    description: 'Fast all-in-one JavaScript runtime',
    monacoLang: 'typescript',
  },
  'browser-wasm': {
    label: 'Browser WASM',
    color: '#6b46c1',
    description: 'Zero cold start WebAssembly for browser edge',
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
  javascript: `export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
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
  'python-wasm': `import json

def handler(request, env, ctx):
    """FunctionFly MicroPython handler - optimized for edge."""
    return {
        "status": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({
            "message": "Hello from FunctionFly (MicroPython)!",
            "method": request.get("method", "GET"),
            "optimized_for": "edge",
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
  go: `package main

import (
	"encoding/json"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Hello from FunctionFly (Go)!",
		"path":    r.URL.Path,
		"method":  r.Method,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}`,
  deno: `export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly (Deno)!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  },
};`,
  bun: `export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    return new Response(
      JSON.stringify({
        message: 'Hello from FunctionFly (Bun)!',
        path: url.pathname,
        method: request.method,
        timestamp: new Date().toISOString(),
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  },
};`,
  'browser-wasm': `use functionfly_sdk::{Request, Response, Env};

#[no_mangle]
pub extern "C" fn handle(req: Request, env: Env) -> Response {
    Response::json(serde_json::json!({
        "message": "Hello from FunctionFly (Browser WASM)!",
        "method": req.method(),
        "cold_start_ms": 0,
    }))
}`,
};

export const RUNTIME_VERSIONS: Record<Runtime, string[]> = {
  typescript: ['ES2022', 'ES2021', 'ES2020'],
  javascript: ['ES2022', 'ES2021', 'ES2020'],
  python: ['3.12', '3.11', '3.10'],
  'python-wasm': ['1.23', '1.22', '1.21'],
  'rust-wasm': ['1.75', '1.74', '1.73'],
  go: ['1.23', '1.22', '1.21'],
  deno: ['1.46', '1.45', '1.44'],
  bun: ['1.1', '1.0'],
  'browser-wasm': ['2.0', '1.0'],
};

export const HTTP_METHODS: HttpMethod[] = ['ANY', 'GET', 'POST', 'PUT', 'PATCH', 'DELETE'];
export const MEMORY_OPTIONS = [64, 128, 256, 512, 1024, 2048];
export const TIMEOUT_OPTIONS = [1000, 5000, 10000, 30000, 60000, 300000];
export const DRAFT_KEY = 'functionfly:new-function-draft';
