import { describe, expect, it } from 'vitest';
import { validateCode } from './codeValidation';

describe('validateCode', () => {
  const validTS = `export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    return Response.json({ ok: true });
  },
};`;

  const validJS = `export default {
  async fetch(request, env, ctx) {
    return new Response(JSON.stringify({ ok: true }), { status: 200 });
  },
};`;

  const validPython = `import json

def handler(request, env, ctx):
    return {
        "status": 200,
        "headers": {"Content-Type": "application/json"},
        "body": json.dumps({"ok": True}),
    }`;

  const validGo = `package main

import "encoding/json"

func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}`;

  const validRust = `use functionfly_sdk::{Request, Response};

#[no_mangle]
pub extern "C" fn handle(req: Request, env: Env) -> Response {
    Response::json(serde_json::json!({"ok": true}))
}`;

  it('returns valid for correct typescript code', () => {
    const result = validateCode(validTS, 'typescript');
    expect(result.valid).toBe(true);
    expect(result.issues.filter(i => i.type === 'error').length).toBe(0);
  });

  it('returns valid for correct javascript code', () => {
    const result = validateCode(validJS, 'javascript');
    expect(result.valid).toBe(true);
    expect(result.issues.filter(i => i.type === 'error').length).toBe(0);
  });

  it('returns valid for correct python code', () => {
    const result = validateCode(validPython, 'python');
    expect(result.valid).toBe(true);
    expect(result.issues.filter(i => i.type === 'error').length).toBe(0);
  });

  it('returns valid for correct go code', () => {
    const result = validateCode(validGo, 'go');
    expect(result.valid).toBe(true);
    expect(result.issues.filter(i => i.type === 'error').length).toBe(0);
  });

  it('returns valid for correct rust/wasm code', () => {
    const result = validateCode(validRust, 'rust-wasm');
    expect(result.valid).toBe(true);
    expect(result.issues.filter(i => i.type === 'error').length).toBe(0);
  });

  it('returns error for empty code', () => {
    const result = validateCode('', 'typescript');
    expect(result.valid).toBe(false);
    expect(result.issues.some(i => i.message === 'Code is empty')).toBe(true);
  });

  it('returns error for go code without package declaration', () => {
    const result = validateCode(`func Handler(w http.ResponseWriter, r *http.Request) {}`, 'go');
    expect(result.issues.some(i => i.message.includes('package declaration'))).toBe(true);
  });

  it('returns warning for code over 100KB', () => {
    const large = 'export default {};\n' + 'a'.repeat(100_001);
    const result = validateCode(large, 'typescript');
    expect(result.issues.some(i => i.message.includes('>100KB'))).toBe(true);
  });

  it('returns warning for missing export', () => {
    const result = validateCode('const x = 1;', 'typescript');
    expect(result.issues.some(i => i.message.includes('No exported handler'))).toBe(true);
  });

  it('returns warning for eval usage', () => {
    const result = validateCode('eval("1+1");', 'typescript');
    expect(result.issues.some(i => i.message.includes('eval()'))).toBe(true);
  });
});
