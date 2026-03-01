import { describe, it, expect } from 'vitest'
import {
  parseQueryParams,
  buildQueryString,
  parseUrl,
  createSlug,
  isValidUrl,
  sanitizeUrl,
  isValidEmail,
  isValidUuid,
  buildPath,
  joinPaths,
  encodeParam,
  decodeParam,
} from './url-utils'

describe('parseQueryParams', () => {
  it('parses search string to record', () => {
    expect(parseQueryParams('?a=1&b=2')).toEqual({ a: '1', b: '2' })
  })

  it('handles empty search', () => {
    expect(parseQueryParams('')).toEqual({})
  })
})

describe('buildQueryString', () => {
  it('builds query string from object', () => {
    const s = buildQueryString({ a: 1, b: 'hello' })
    expect(s).toContain('a=1')
    expect(s).toContain('b=hello')
  })
})

describe('parseUrl', () => {
  it('parses url and query', () => {
    const parsed = parseUrl('https://example.com/path?a=1')
    expect(parsed.url).toContain('example.com')
    expect(parsed.query).toEqual({ a: '1' })
  })
})

describe('createSlug', () => {
  it('lowercases and replaces spaces with separator', () => {
    expect(createSlug('Hello World')).toBe('hello-world')
  })

  it('respects maxLength', () => {
    const long = 'a'.repeat(150)
    expect(createSlug(long, { maxLength: 10 })).toHaveLength(10)
  })

  it('uses custom separator when provided', () => {
    expect(createSlug('hello world', { separator: '_' })).toMatch(/hello.world/)
  })
})

describe('isValidUrl', () => {
  it('accepts https URL', () => {
    expect(isValidUrl('https://example.com')).toBe(true)
  })

  it('accepts http URL', () => {
    expect(isValidUrl('http://example.com')).toBe(true)
  })

  it('rejects URL without protocol when require_protocol is true', () => {
    expect(isValidUrl('example.com')).toBe(false)
  })

  it('rejects invalid URL', () => {
    expect(isValidUrl('not a url')).toBe(false)
  })
})

describe('sanitizeUrl', () => {
  it('trims and escapes', () => {
    const result = sanitizeUrl('  <script>  ')
    expect(result).not.toContain('<')
    expect(result).not.toContain('>')
  })
})

describe('isValidEmail', () => {
  it('accepts valid email', () => {
    expect(isValidEmail('user@example.com')).toBe(true)
  })

  it('rejects invalid email', () => {
    expect(isValidEmail('invalid')).toBe(false)
  })
})

describe('isValidUuid', () => {
  it('accepts valid UUID', () => {
    expect(isValidUuid('550e8400-e29b-41d4-a716-446655440000')).toBe(true)
  })

  it('rejects invalid UUID', () => {
    expect(isValidUuid('not-a-uuid')).toBe(false)
  })
})

describe('buildPath', () => {
  it('joins base and parts with single leading slash', () => {
    expect(buildPath('api', 'v1', 'users')).toBe('/api/v1/users')
  })

  it('strips leading/trailing slashes from parts (base without extra slashes)', () => {
    expect(buildPath('api', 'v1', 'users')).toBe('/api/v1/users')
  })

  it('filters empty parts', () => {
    expect(buildPath('api', '', 'users')).toBe('/api/users')
  })
})

describe('joinPaths', () => {
  it('joins path segments', () => {
    expect(joinPaths('a', 'b', 'c')).toBe('a/b/c')
  })

  it('strips slashes and filters empty', () => {
    expect(joinPaths('/a/', '', '/b/')).toBe('a/b')
  })
})

describe('encodeParam', () => {
  it('encodes special characters', () => {
    expect(encodeParam('a b')).toBe('a%20b')
  })
})

describe('decodeParam', () => {
  it('decodes encoded value', () => {
    expect(decodeParam('a%20b')).toBe('a b')
  })
})
