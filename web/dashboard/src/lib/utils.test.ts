import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  cn,
  formatDate,
  formatDateTime,
  formatNumber,
  formatBytes,
  formatDuration,
  truncate,
  capitalize,
  sleep,
  generateId,
} from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('a', 'b')).toBe('a b')
  })

  it('handles conditional classes', () => {
    expect(cn('a', false && 'b', 'c')).toBe('a c')
  })

  it('merges tailwind classes and dedupes conflicting ones', () => {
    expect(cn('px-2 py-1', 'px-4')).toBe('py-1 px-4')
  })

  it('handles undefined and null', () => {
    expect(cn('a', undefined, null, 'b')).toBe('a b')
  })
})

describe('formatDate', () => {
  it('formats ISO date string', () => {
    const result = formatDate('2024-06-15T12:00:00Z')
    expect(result).toMatch(/Jun|June/)
    expect(result).toContain('15')
    expect(result).toContain('2024')
  })

  it('formats Date instance', () => {
    const result = formatDate(new Date('2024-06-15'))
    expect(result).toMatch(/Jun|June/)
    expect(result).toContain('2024')
    expect(result).toMatch(/\d{1,2}/) // day number
  })
})

describe('formatDateTime', () => {
  it('includes time for ISO string', () => {
    const result = formatDateTime('2024-06-15T14:30:00Z')
    expect(result).toMatch(/Jun|June/)
    expect(result).toContain('15')
    expect(result).toContain('2024')
    expect(result).toMatch(/\d|:/)
  })
})

describe('formatNumber', () => {
  it('formats integer with en-US locale', () => {
    expect(formatNumber(1234567)).toBe('1,234,567')
  })

  it('formats decimal', () => {
    expect(formatNumber(1234.56)).toBe('1,234.56')
  })
})

describe('formatBytes', () => {
  it('returns "0 B" for zero', () => {
    expect(formatBytes(0)).toBe('0 B')
  })

  it('formats bytes', () => {
    expect(formatBytes(500)).toBe('500 B')
  })

  it('formats kilobytes', () => {
    expect(formatBytes(1024)).toBe('1 KB')
    expect(formatBytes(1536)).toBe('1.5 KB')
  })

  it('formats megabytes', () => {
    expect(formatBytes(1024 * 1024)).toBe('1 MB')
  })

  it('formats gigabytes', () => {
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1 GB')
  })
})

describe('formatDuration', () => {
  it('formats milliseconds', () => {
    expect(formatDuration(500)).toBe('500ms')
  })

  it('formats seconds', () => {
    expect(formatDuration(2500)).toBe('2.5s')
  })

  it('formats minutes', () => {
    expect(formatDuration(90000)).toBe('1.5m')
  })
})

describe('truncate', () => {
  it('returns string as-is when within length', () => {
    expect(truncate('hello', 10)).toBe('hello')
  })

  it('truncates and appends ellipsis when over length', () => {
    expect(truncate('hello world', 5)).toBe('hello...')
  })

  it('handles exact length', () => {
    expect(truncate('hello', 5)).toBe('hello')
  })
})

describe('capitalize', () => {
  it('capitalizes first character', () => {
    expect(capitalize('hello')).toBe('Hello')
  })

  it('leaves rest unchanged', () => {
    expect(capitalize('hELLO')).toBe('HELLO')
  })

  it('handles single character', () => {
    expect(capitalize('a')).toBe('A')
  })
})

describe('sleep', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('resolves after given ms', async () => {
    const p = sleep(100)
    vi.advanceTimersByTime(100)
    await expect(p).resolves.toBeUndefined()
  })
})

describe('generateId', () => {
  it('returns a string', () => {
    expect(typeof generateId()).toBe('string')
  })

  it('returns alphanumeric substring of length 7', () => {
    const id = generateId()
    expect(id).toHaveLength(7)
    expect(id).toMatch(/^[a-z0-9]+$/)
  })

  it('returns different values on each call', () => {
    const a = generateId()
    const b = generateId()
    expect(a).not.toBe(b)
  })
})
