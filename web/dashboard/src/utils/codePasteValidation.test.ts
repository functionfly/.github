import { describe, expect, it } from 'vitest';
import {
  ALLOWED_PROVIDERS,
  ALLOWED_REGIONS,
  formatCodeSize,
  getCodeByteLength,
  MAX_CODE_BYTES,
  validateCodeSize,
  validateFunctionName,
  validateImportConfig,
} from './codePasteValidation';

describe('codePasteValidation', () => {
  it('validates function names', () => {
    expect(validateFunctionName('helloWorld')).toBeNull();
    expect(validateFunctionName('1bad')).toMatch(/letter/);
    expect(validateFunctionName('')).toMatch(/required/);
  });

  it('validates code size', () => {
    expect(validateCodeSize('print("hi")')).toBeNull();
    expect(validateCodeSize('')).toMatch(/enter some code/);
    expect(validateCodeSize('a'.repeat(MAX_CODE_BYTES + 1))).toMatch(/maximum size/);
  });

  it('formats code size labels', () => {
    expect(formatCodeSize(512)).toBe('512 B');
    expect(formatCodeSize(2048)).toBe('2.0 KB');
  });

  it('counts utf-8 bytes', () => {
    expect(getCodeByteLength('hello')).toBe(5);
    expect(getCodeByteLength('🙂')).toBe(4);
  });

  it('validates import config', () => {
    expect(
      validateImportConfig({
        providers: [...ALLOWED_PROVIDERS],
        region: ALLOWED_REGIONS[0],
      })
    ).toBeNull();
    expect(
      validateImportConfig({
        providers: [],
        region: ALLOWED_REGIONS[0],
      })
    ).toMatch(/provider/);
    expect(
      validateImportConfig({
        providers: ['cloud'],
        region: 'invalid',
      })
    ).toMatch(/region/);
  });
});
