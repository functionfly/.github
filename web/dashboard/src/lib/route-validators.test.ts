import { describe, expect, it } from 'vitest';
import {
  authorSchema,
  blogPostParamsSchema,
  emailSchema,
  functionNameSchema,
  functionParamsSchema,
  pageSchema,
  replayParamsSchema,
  safeValidateRouteParams,
  searchQuerySchema,
  slugSchema,
  usernameSchema,
  userProfileParamsSchema,
  uuidSchema,
  validateRouteParams,
  versionedFunctionParamsSchema,
  versionSchema,
} from './route-validators';

describe('authorSchema', () => {
  it('accepts valid author', () => {
    expect(authorSchema.parse('acme')).toBe('acme');
    expect(authorSchema.parse('user_name')).toBe('user_name');
    expect(authorSchema.parse('my-org')).toBe('my-org');
  });

  it('rejects empty', () => {
    expect(() => authorSchema.parse('')).toThrow();
  });

  it('rejects invalid characters', () => {
    expect(() => authorSchema.parse('user name')).toThrow();
    expect(() => authorSchema.parse('user@org')).toThrow();
  });

  it('rejects over 50 chars', () => {
    expect(() => authorSchema.parse('a'.repeat(51))).toThrow();
  });
});

describe('functionNameSchema', () => {
  it('accepts valid function name', () => {
    expect(functionNameSchema.parse('myFunc')).toBe('myFunc');
    expect(functionNameSchema.parse('hello-world')).toBe('hello-world');
  });

  it('rejects empty', () => {
    expect(() => functionNameSchema.parse('')).toThrow();
  });

  it('rejects over 100 chars', () => {
    expect(() => functionNameSchema.parse('a'.repeat(101))).toThrow();
  });
});

describe('uuidSchema', () => {
  it('accepts valid UUID', () => {
    const uuid = '550e8400-e29b-41d4-a716-446655440000';
    expect(uuidSchema.parse(uuid)).toBe(uuid);
  });

  it('rejects invalid UUID', () => {
    expect(() => uuidSchema.parse('not-a-uuid')).toThrow();
    expect(() => uuidSchema.parse('550e8400-e29b-41d4-a716')).toThrow();
  });
});

describe('slugSchema', () => {
  it('accepts valid slug', () => {
    expect(slugSchema.parse('my-post')).toBe('my-post');
    expect(slugSchema.parse('hello-world-123')).toBe('hello-world-123');
  });

  it('rejects uppercase', () => {
    expect(() => slugSchema.parse('My-Post')).toThrow();
  });

  it('rejects spaces', () => {
    expect(() => slugSchema.parse('my post')).toThrow();
  });
});

describe('usernameSchema', () => {
  it('accepts valid username', () => {
    expect(usernameSchema.parse('jane')).toBe('jane');
    expect(usernameSchema.parse('user_123')).toBe('user_123');
  });

  it('rejects less than 3 chars', () => {
    expect(() => usernameSchema.parse('ab')).toThrow();
  });

  it('rejects over 30 chars', () => {
    expect(() => usernameSchema.parse('a'.repeat(31))).toThrow();
  });

  it('rejects hyphens', () => {
    expect(() => usernameSchema.parse('user-name')).toThrow();
  });
});

describe('pageSchema', () => {
  it('defaults to 1', () => {
    expect(pageSchema.parse(undefined)).toBe(1);
  });

  it('accepts positive integer', () => {
    expect(pageSchema.parse(5)).toBe(5);
  });

  it('coerces string to number', () => {
    expect(pageSchema.parse('3')).toBe(3);
  });

  it('rejects zero or negative', () => {
    expect(() => pageSchema.parse(0)).toThrow();
    expect(() => pageSchema.parse(-1)).toThrow();
  });
});

describe('emailSchema', () => {
  it('accepts valid email', () => {
    expect(emailSchema.parse('a@b.co')).toBe('a@b.co');
  });

  it('rejects invalid email', () => {
    expect(() => emailSchema.parse('notanemail')).toThrow();
  });
});

describe('versionSchema', () => {
  it('accepts semver-like', () => {
    expect(versionSchema.parse('1.0.0')).toBe('1.0.0');
    expect(versionSchema.parse('2.1.3-beta')).toBe('2.1.3-beta');
  });

  it('rejects invalid version', () => {
    expect(() => versionSchema.parse('v1.0.0')).toThrow();
    expect(() => versionSchema.parse('1.0')).toThrow();
  });
});

describe('validateRouteParams', () => {
  it('returns parsed data when valid', () => {
    const result = validateRouteParams({ author: 'acme', name: 'fn' }, functionParamsSchema);
    expect(result).toEqual({ author: 'acme', name: 'fn' });
  });

  it('throws on invalid', () => {
    expect(() => validateRouteParams({ author: '', name: 'fn' }, functionParamsSchema)).toThrow();
  });
});

describe('safeValidateRouteParams', () => {
  it('returns success with data when valid', () => {
    const result = safeValidateRouteParams({ author: 'acme', name: 'fn' }, functionParamsSchema);
    expect(result.success).toBe(true);
    if (result.success) expect(result.data).toEqual({ author: 'acme', name: 'fn' });
  });

  it('returns success: false with error when invalid', () => {
    const result = safeValidateRouteParams({ author: '', name: 'fn' }, functionParamsSchema);
    expect(result.success).toBe(false);
    if (!result.success && 'error' in result) expect(result.error).toBeDefined();
  });
});

describe('functionParamsSchema', () => {
  it('parses author and name', () => {
    expect(functionParamsSchema.parse({ author: 'acme', name: 'my-fn' })).toEqual({
      author: 'acme',
      name: 'my-fn',
    });
  });
});

describe('blogPostParamsSchema', () => {
  it('parses slug', () => {
    expect(blogPostParamsSchema.parse({ slug: 'my-post' })).toEqual({ slug: 'my-post' });
  });
});

describe('userProfileParamsSchema', () => {
  it('parses username', () => {
    expect(userProfileParamsSchema.parse({ username: 'jane_doe' })).toEqual({
      username: 'jane_doe',
    });
  });
});

describe('replayParamsSchema', () => {
  it('parses execId UUID', () => {
    const uuid = '550e8400-e29b-41d4-a716-446655440000';
    expect(replayParamsSchema.parse({ execId: uuid })).toEqual({ execId: uuid });
  });
});

describe('versionedFunctionParamsSchema', () => {
  it('parses author, name, optional version', () => {
    expect(
      versionedFunctionParamsSchema.parse({ author: 'acme', name: 'fn', version: '1.0.0' })
    ).toEqual({ author: 'acme', name: 'fn', version: '1.0.0' });
    expect(versionedFunctionParamsSchema.parse({ author: 'acme', name: 'fn' })).toEqual({
      author: 'acme',
      name: 'fn',
    });
  });
});

describe('searchQuerySchema', () => {
  it('parses q, page, limit', () => {
    expect(searchQuerySchema.parse({ q: 'test', page: 2, limit: 10 })).toEqual({
      q: 'test',
      page: 2,
      limit: 10,
    });
  });

  it('rejects empty q', () => {
    expect(() => searchQuerySchema.parse({ q: '' })).toThrow();
  });
});
