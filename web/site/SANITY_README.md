# FunctionFly Sanity CMS Integration

This document outlines the Sanity CMS integration for the FunctionFly website, providing a comprehensive content management system for blog posts, documentation, case studies, tools, benchmarks, and more.

## Overview

The Sanity integration includes:

- **Blog Posts**: Technical articles, tutorials, and insights
- **Documentation**: API docs, guides, and reference materials
- **Case Studies**: Success stories and client implementations
- **Tools & Templates**: Free downloadable resources
- **Benchmarks**: Performance comparisons and research
- **Authors**: Content creator profiles
- **Categories**: Content organization and taxonomy

## Setup Instructions

### 1. Create a Sanity Project

1. Go to [sanity.io](https://sanity.io) and create an account
2. Create a new project called "FunctionFly Content Studio"
3. Choose the default dataset name: `production`
4. Copy your project ID from the project settings

### 2. Environment Variables

Create a `.env` file in the `web/site` directory:

```env
# Sanity Configuration
SANITY_PROJECT_ID=your_project_id_here
SANITY_DATASET=production
SANITY_API_VERSION=2024-01-01

# Sanity Studio Configuration (for studio deployment)
SANITY_STUDIO_PROJECT_ID=your_project_id_here
SANITY_STUDIO_DATASET=production
```

### 3. Install Dependencies

Run the following command in the `web/site` directory:

```bash
npm install
```

This will install the Sanity packages that were added to `package.json`.

### 4. Initialize Sanity Studio

```bash
cd studio
npm install
npx sanity init
```

When prompted:
- Select "Clean project with no predefined schemas"
- Choose the project you created
- Select the dataset (production)

### 5. Deploy Sanity Studio (Optional)

To deploy the Sanity Studio to sanity.io:

```bash
cd studio
npm run build
npx sanity deploy
```

This will give you a URL like `https://functionfly-content-studio.sanity.studio`

### 6. Create Initial Content

1. Access your Sanity Studio
2. Create categories first (they're referenced by other content types)
3. Create author profiles
4. Start adding content using the predefined schemas

## Content Types

### Blog Posts
- Title, description, and body content
- Author and category references
- SEO fields (title, description, keywords)
- Hero image with alt text
- Tags and publication dates
- Draft/publish workflow

### Documentation
- Hierarchical organization with categories
- Order field for custom sorting
- Version tracking with updated dates
- SEO optimization
- Cross-linking support

### Case Studies
- Company information and logos
- Challenge/solution/results format
- Featured case studies
- Industry categorization
- Performance metrics

### Tools & Templates
- Multiple types (templates, tools, checklists, guides, calculators)
- Download and demo links
- Category organization
- Featured resources

### Benchmarks
- Performance comparison data
- Methodology documentation
- Key findings and insights
- Raw data file attachments
- Period tracking (Q1 2024, monthly, etc.)

### Authors
- Bio, photos, and social links
- Role/title information
- Active/inactive status

### Categories
- Hierarchical organization
- Custom colors and icons
- Display order
- SEO metadata

## File Structure

```
web/site/
├── sanity/
│   ├── schemas/
│   │   ├── index.ts          # Schema exports
│   │   ├── blogPost.ts       # Blog post schema
│   │   ├── doc.ts           # Documentation schema
│   │   ├── caseStudy.ts     # Case study schema
│   │   ├── tool.ts          # Tool/template schema
│   │   ├── benchmark.ts     # Benchmark schema
│   │   ├── author.ts        # Author schema
│   │   └── category.ts      # Category schema
│   ├── config.ts            # Sanity client config
│   └── cli.ts              # CLI configuration
├── studio/
│   ├── package.json        # Studio dependencies
│   └── sanity.config.ts    # Studio configuration
├── src/
│   ├── lib/
│   │   └── sanity.ts       # Client and queries
│   ├── components/
│   │   └── PortableText.astro  # Content renderer
│   └── pages/
│       ├── blog/
│       │   ├── index.astro     # Blog listing
│       │   └── [slug].astro    # Individual posts
│       ├── docs/
│       │   └── [...slug].astro # Documentation pages
│       ├── case-studies/
│       │   ├── index.astro     # Case studies listing
│       │   └── [slug].astro    # Individual case studies
│       └── tools/
│           ├── index.astro     # Tools listing
│           └── [slug].astro    # Individual tools
└── .env.example             # Environment variables template
```

## GROQ Queries

The integration uses optimized GROQ queries for fetching content:

- `blogPosts`: All published blog posts with author and category data
- `blogPostBySlug`: Individual blog post by slug
- `docs`: All documentation pages
- `docBySlug`: Individual documentation page
- `caseStudies`: All case studies
- `caseStudyBySlug`: Individual case study
- `tools`: All tools and templates
- `toolBySlug`: Individual tool
- `benchmarks`: All benchmarks
- `benchmarkBySlug`: Individual benchmark

## SEO & Performance

- Structured data (JSON-LD) for rich snippets
- Meta descriptions and titles
- Open Graph images
- Sitemap integration
- Optimized images with Sanity's image processing
- CDN delivery for media assets

## Development Workflow

1. **Content Creation**: Use Sanity Studio to create and edit content
2. **Preview**: Content is automatically available via the Astro build
3. **SEO**: Each content type includes SEO fields
4. **Publishing**: Use draft/publish workflow in Sanity
5. **Deployment**: Content changes deploy automatically with your site

## Customization

### Adding New Content Types

1. Create a new schema file in `sanity/schemas/`
2. Export it from `sanity/schemas/index.ts`
3. Add GROQ queries in `src/lib/sanity.ts`
4. Create corresponding Astro pages
5. Update the Sanity Studio desk structure

### Modifying Existing Schemas

Edit the schema files in `sanity/schemas/` and redeploy the studio:

```bash
cd studio
npm run build
npx sanity deploy
```

## Best Practices

- Always use descriptive alt text for images
- Fill out SEO fields for better search visibility
- Use categories consistently for navigation
- Keep content updated with publication dates
- Use the draft workflow for work-in-progress content
- Leverage Sanity's real-time collaboration features

## Support

For Sanity-specific issues:
- [Sanity Documentation](https://www.sanity.io/docs)
- [Sanity Community](https://slack.sanity.io/)

For FunctionFly integration issues, check the existing codebase and schemas.