# SEO Implementation Guide

This directory contains SEO components and utilities for FunctionFly, using the best SEO libraries for React 2026.

## Libraries Used

### 1. `react-meta-seo` - Meta Tags & Head Management

**Why chosen:** Zero runtime overhead, React 19 native support, no provider wrapper needed.

**Usage:**

```tsx
import { MetaTags } from '@/components/seo/MetaTags';

// In your component
<MetaTags
  title="Page Title"
  description="Page description"
  keywords={["keyword1", "keyword2"]}
/>
```

### 2. `react-schemaorg` - Structured Data (JSON-LD)

**Why chosen:** Full TypeScript support, Google-backed, type-safe JSON-LD validation.

**Usage:**

```tsx
import { LandingPageStructuredData } from '@/components/seo/StructuredData';

// Add to your page component
<LandingPageStructuredData />
```

### 3. `web-vitals` - Core Web Vitals Monitoring

**Why chosen:** Official Google library, measures all CWV metrics, attribution data included.

**Usage:**

```tsx
import { useWebVitals } from '@/hooks/useWebVitals';

function MyComponent() {
  useWebVitals((metrics) => {
    // Send to analytics
    console.log('CWV:', metrics);
  });

  return <div>...</div>;
}
```

### 4. `next-sitemap` - Sitemap Generation

**Why chosen:** Even for non-Next.js apps, it's the most robust sitemap generator.

**Configuration:** `next-sitemap.config.js`

**Usage:**

```bash
# Generate sitemap after build
npm run build  # Automatically runs next-sitemap via postbuild

# Or manually generate
npx next-sitemap
```

### 5. `generate-robotstxt` - Robots.txt Generation

**Why chosen:** Most popular and actively maintained robots.txt generator.

**Configuration:** `robots.config.js`

**Usage:**

```bash
# Generate robots.txt
npm run generate-robots

# Or both sitemap and robots.txt
npm run seo
```

## Implementation Checklist

### ✅ Completed

- [x] Meta tags component (`MetaTags.tsx`)
- [x] Structured data components (`StructuredData.tsx`)
- [x] Web Vitals monitoring hook (`useWebVitals.ts`)
- [x] Sitemap configuration (`next-sitemap.config.js`)
- [x] Robots.txt configuration (`robots.config.js`)
- [x] Landing page SEO integration
- [x] Build scripts updated

### 🔄 Next Steps

- [ ] Add SEO to other public pages (Pricing, Features, Team)
- [ ] Create Open Graph images
- [ ] Set up Google Analytics/Search Console
- [ ] Add breadcrumb structured data
- [ ] Implement FAQ structured data for FAQ section

## Performance Benefits

- **Zero runtime overhead** for meta tags (react-meta-seo)
- **Core Web Vitals monitoring** for SEO ranking signals
- **Structured data** for rich snippets in search results
- **Proper sitemap** for better crawling
- **Robots.txt** for controlled indexing

## Testing SEO

1. **Meta Tags:** Use browser dev tools → Elements → `<head>`
2. **Structured Data:** Use Google's [Rich Results Test](https://search.google.com/test/rich-results)
3. **Core Web Vitals:** Use Chrome DevTools → Lighthouse
4. **Sitemap:** Check `/sitemap.xml` URL
5. **Robots.txt:** Check `/robots.txt` URL

## Analytics Integration

Update the `useWebVitals` hook to send data to your analytics service:

```typescript
// In useWebVitals.ts
if (typeof window !== 'undefined' && (window as any).gtag) {
  (window as any).gtag('event', 'web_vitals', {
    event_category: 'Web Vitals',
    event_label: metric.name,
    value: Math.round(metric.value),
  });
}
```
