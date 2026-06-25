/**
 * Blog utilities tests
 *
 * Tests the centralized blog helpers that are used throughout the
 * admin blog page. These cover slug generation, tag parsing, date
 * formatting, and the form data conversions.
 */

import { describe, expect, it } from 'vitest';
import {
  formatPostDate,
  formDataToPayload,
  getAuthorName,
  keywordsToString,
  postToFormData,
  slugify,
  stringToKeywords,
  stringToTags,
  tagsToString,
} from './blog';
import type { BlogPost } from '@/types/blog';

describe('slugify', () => {
  it('converts title to lowercase with hyphens', () => {
    expect(slugify('Hello World')).toBe('hello-world');
  });

  it('replaces spaces with hyphens', () => {
    expect(slugify('Foo  Bar  Baz')).toBe('foo-bar-baz');
  });

  it('collapses multiple hyphens', () => {
    expect(slugify('foo---bar')).toBe('foo-bar');
  });

  it('trims leading and trailing hyphens', () => {
    expect(slugify('!!!hello!!!')).toBe('hello');
  });

  it('strips non-alphanumeric characters except hyphens', () => {
    expect(slugify('Hello, World! 123')).toBe('hello-world-123');
  });

  it('handles empty string', () => {
    expect(slugify('')).toBe('');
  });
});

describe('getAuthorName', () => {
  it('returns string author as-is', () => {
    expect(getAuthorName({ author: 'Jane Doe' } as any)).toBe('Jane Doe');
  });

  it('returns name from object author', () => {
    expect(getAuthorName({ author: { name: 'Jane', slug: 'jane' } } as any)).toBe('Jane');
  });

  it('returns empty string for null/undefined', () => {
    expect(getAuthorName(null)).toBe('');
    expect(getAuthorName(undefined)).toBe('');
  });

  it('returns empty string for missing author', () => {
    expect(getAuthorName({} as any)).toBe('');
  });
});

describe('tagsToString / stringToTags', () => {
  it('converts array to comma-separated string', () => {
    expect(tagsToString(['a', 'b', 'c'])).toBe('a, b, c');
  });

  it('handles null/undefined', () => {
    expect(tagsToString(null)).toBe('');
    expect(tagsToString(undefined)).toBe('');
  });

  it('converts comma-separated string to array, trimming whitespace', () => {
    expect(stringToTags('a, b ,c')).toEqual(['a', 'b', 'c']);
  });

  it('filters empty strings', () => {
    expect(stringToTags('a,, b, ,c,')).toEqual(['a', 'b', 'c']);
  });

  it('round-trip preserves data', () => {
    const original = ['foo', 'bar', 'baz'];
    expect(stringToTags(tagsToString(original))).toEqual(original);
  });
});

describe('keywordsToString / stringToKeywords', () => {
  it('round-trip preserves data', () => {
    const original = ['ai', 'serverless', 'agents'];
    expect(stringToKeywords(keywordsToString(original))).toEqual(original);
  });

  it('handles empty input', () => {
    expect(stringToKeywords('')).toEqual([]);
    expect(keywordsToString([])).toBe('');
  });
});

describe('formatPostDate', () => {
  it('formats valid ISO string', () => {
    const result = formatPostDate('2024-01-15T10:00:00Z');
    expect(result).not.toBe('—');
    expect(result).toContain('2024');
  });

  it('returns em-dash for null/undefined', () => {
    expect(formatPostDate(null)).toBe('—');
    expect(formatPostDate(undefined)).toBe('—');
  });

  it('returns em-dash for invalid string', () => {
    expect(formatPostDate('not a date')).toBe('—');
  });

  it('returns em-dash for empty string', () => {
    expect(formatPostDate('')).toBe('—');
  });
});

describe('postToFormData', () => {
  const fullPost: BlogPost = {
    id: '123e4567-e89b-12d3-a456-426614174000',
    title: 'Test Post',
    slug: 'test-post',
    body: '# Hello\n\nThis is markdown.',
    description: 'A test post',
    tags: ['tag1', 'tag2'],
    heroImage: { url: 'https://example.com/img.jpg' },
    status: 'published',
    publishedAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-16T10:00:00Z',
    createdAt: '2024-01-14T10:00:00Z',
    seoTitle: 'SEO Title',
    seoDescription: 'SEO Description',
    keywords: ['kw1', 'kw2'],
    canonicalUrl: 'https://example.com/canonical',
    ogImage: { url: 'https://example.com/og.jpg', alt: 'OG Alt' },
    isPublished: true,
    author: { name: 'Jane Doe', slug: 'jane' },
  };

  it('converts all fields to form strings', () => {
    const form = postToFormData(fullPost);
    expect(form.title).toBe('Test Post');
    expect(form.slug).toBe('test-post');
    expect(form.body).toBe('# Hello\n\nThis is markdown.');
    expect(form.excerpt).toBe('A test post');
    expect(form.author).toBe('Jane Doe');
    expect(form.tags).toBe('tag1, tag2');
    expect(form.isPublished).toBe(true);
    expect(form.featuredImage).toBe('https://example.com/img.jpg');
    expect(form.seoTitle).toBe('SEO Title');
    expect(form.seoDescription).toBe('SEO Description');
    expect(form.keywords).toBe('kw1, kw2');
    expect(form.canonicalUrl).toBe('https://example.com/canonical');
    expect(form.ogImageUrl).toBe('https://example.com/og.jpg');
    expect(form.ogImageAlt).toBe('OG Alt');
  });

  it('handles missing optional fields', () => {
    const minimal: BlogPost = {
      ...fullPost,
      description: undefined,
      tags: [],
      heroImage: undefined,
      seoTitle: undefined,
      seoDescription: undefined,
      keywords: [],
      canonicalUrl: undefined,
      ogImage: undefined,
    };
    const form = postToFormData(minimal);
    expect(form.excerpt).toBe('');
    expect(form.tags).toBe('');
    expect(form.featuredImage).toBe('');
    expect(form.seoTitle).toBe('');
  });

  it('derives isPublished from status when not explicit', () => {
    const post: BlogPost = { ...fullPost, isPublished: undefined, status: 'published' };
    expect(postToFormData(post).isPublished).toBe(true);
  });
});

describe('formDataToPayload', () => {
  it('builds a payload with markdown body (not JSON)', () => {
    const form = {
      title: 'Test',
      slug: 'test',
      body: '# Hello\n\nWorld',
      excerpt: 'desc',
      author: 'Jane',
      tags: 'a, b',
      isPublished: true,
      featuredImage: 'https://example.com/img.jpg',
      seoTitle: 'SEO',
      seoDescription: 'SEO desc',
      keywords: 'kw1, kw2',
      canonicalUrl: 'https://example.com/canon',
      ogImageUrl: 'https://example.com/og.jpg',
      ogImageAlt: 'OG Alt',
    };

    const payload = formDataToPayload(form);
    // Body should be raw markdown, not JSON
    expect(payload.body).toBe('# Hello\n\nWorld');
    expect(payload.status).toBe('published');
    expect(payload.tags).toEqual(['a', 'b']);
    expect(payload.keywords).toEqual(['kw1', 'kw2']);
    expect(payload.heroImage).toEqual({ url: 'https://example.com/img.jpg' });
    expect(payload.ogImage).toEqual({ url: 'https://example.com/og.jpg', alt: 'OG Alt' });
  });

  it('generates slug from title if missing', () => {
    const payload = formDataToPayload({
      title: 'My Awesome Post',
      slug: '',
      body: '',
      excerpt: '',
      author: '',
      tags: '',
      isPublished: false,
      featuredImage: '',
      seoTitle: '',
      seoDescription: '',
      keywords: '',
      canonicalUrl: '',
      ogImageUrl: '',
      ogImageAlt: '',
    });
    expect(payload.slug).toBe('my-awesome-post');
  });

  it('sets status to draft when not published', () => {
    const payload = formDataToPayload({
      title: 'Test',
      slug: 'test',
      body: '',
      excerpt: '',
      author: '',
      tags: '',
      isPublished: false,
      featuredImage: '',
      seoTitle: '',
      seoDescription: '',
      keywords: '',
      canonicalUrl: '',
      ogImageUrl: '',
      ogImageAlt: '',
    });
    expect(payload.status).toBe('draft');
  });

  it('omits empty optional fields', () => {
    const payload = formDataToPayload({
      title: 'Test',
      slug: 'test',
      body: '',
      excerpt: '',
      author: '',
      tags: '',
      isPublished: false,
      featuredImage: '',
      seoTitle: '',
      seoDescription: '',
      keywords: '',
      canonicalUrl: '',
      ogImageUrl: '',
      ogImageAlt: '',
    });
    expect(payload.featuredImage).toBeUndefined();
    expect(payload.seoTitle).toBeUndefined();
    expect(payload.keywords).toBeUndefined();
    expect(payload.ogImage).toBeUndefined();
  });
});
