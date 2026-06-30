/**
 * Blog API Client — reads from Sanity CMS via GROQ
 */

import { getClient } from "./sanity";

// Types matching what the frontend components expect
export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  body: unknown;
  authorId?: string;
  categoryId?: string;
  tags?: string[];
  heroImage?: { url: string; alt: string; caption?: string } | null;
  status: string;
  publishedAt?: string | null;
  scheduledAt?: string | null;
  seoTitle?: string | null;
  seoDescription?: string | null;
  keywords?: string[];
  canonicalUrl?: string | null;
  ogImage?: { url: string; alt: string } | null;
  campaign?: string;
  ownerId?: string;
  createdAt: string;
  updatedAt: string;
  author?: {
    name: string;
    slug: string;
    email?: string;
    photo?: { url: string } | null;
    bio?: string;
    role?: string;
  } | null;
  category?: {
    title: string;
    slug: string;
  } | null;
}

export interface Author {
  id: string;
  name: string;
  slug: string;
  bio?: string;
  photo?: unknown;
  email?: string;
  website?: string;
  socialLinks?: { platform: string; url: string }[];
  role?: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Category {
  id: string;
  title: string;
  slug: string;
  description?: string;
  color?: string;
  icon?: string;
  order: number;
  postCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

export interface PostsQuery {
  page?: number;
  limit?: number;
  category?: string;
  author?: string;
  tag?: string;
  search?: string;
  status?: string;
}

const PUBLISHED_FILTER = "defined(publishedAt) && publishedAt <= now()";

const POST_FIELDS = `
  _id,
  title,
  "slug": slug.current,
  description,
  body,
  tags,
  "heroImage": heroImage {
    "url": asset->url,
    "alt": coalesce(alt, title)
  },
  publishedAt,
  seoTitle,
  seoDescription,
  _createdAt,
  _updatedAt,
  "author": author-> {
    name,
    "slug": slug.current,
    bio,
    role,
    "photo": { "url": photo.asset->url }
  },
  "category": category-> {
    title,
    "slug": slug.current
  }
`;

function mapPost(raw: any): BlogPost {
  return {
    id: raw._id,
    title: raw.title ?? "",
    slug: raw.slug ?? "",
    description: raw.description ?? "",
    body: raw.body ?? null,
    tags: raw.tags ?? [],
    heroImage: raw.heroImage ?? null,
    status: raw.publishedAt ? "published" : "draft",
    publishedAt: raw.publishedAt ?? null,
    seoTitle: raw.seoTitle ?? null,
    seoDescription: raw.seoDescription ?? null,
    createdAt: raw._createdAt ?? "",
    updatedAt: raw._updatedAt ?? "",
    author: raw.author
      ? {
          name: raw.author.name ?? "",
          slug: raw.author.slug ?? "",
          bio: raw.author.bio ?? undefined,
          role: raw.author.role ?? undefined,
          photo: raw.author.photo ?? null,
        }
      : null,
    category: raw.category
      ? {
          title: raw.category.title ?? "",
          slug: raw.category.slug ?? "",
        }
      : null,
  };
}

export async function getPosts(
  query?: PostsQuery,
): Promise<PaginatedResponse<BlogPost>> {
  const client = getClient();
  if (!client) {
    return { data: [], meta: { total: 0, page: 1, limit: 10, totalPages: 0 } };
  }

  const page = query?.page && query.page > 0 ? query.page : 1;
  const limit =
    query?.limit && query.limit > 0 && query.limit <= 50 ? query.limit : 10;
  const offset = (page - 1) * limit;

  const conditions: string[] = [`_type == "blogPost"`];

  if (query?.status === "published" || !query?.status) {
    conditions.push(PUBLISHED_FILTER);
  }

  if (query?.category) {
    conditions.push(`category->slug.current == "${query.category}"`);
  }

  if (query?.author) {
    conditions.push(`author->slug.current == "${query.author}"`);
  }

  if (query?.tag) {
    conditions.push(`"${query.tag}" in tags`);
  }

  const where = conditions.join(" && ");

  try {
    if (query?.search) {
      const searchFilter = `${where} && (title match $search || description match $search || body match $search)`;
      const [total, data] = await Promise.all([
        client.fetch<number>(
          `count(*[${searchFilter}])`,
          { search: `${query.search}*` },
        ),
        client.fetch<any[]>(
          `*[${searchFilter}] | order(publishedAt desc) [${offset}...${offset + limit}] {${POST_FIELDS}}`,
          { search: `${query.search}*` },
        ),
      ]);
      const totalPages = Math.ceil(total / limit);
      return { data: (data ?? []).map(mapPost), meta: { total, page, limit, totalPages } };
    }

    const [total, data] = await Promise.all([
      client.fetch<number>(`count(*[${where}])`),
      client.fetch<any[]>(
        `*[${where}] | order(publishedAt desc) [${offset}...${offset + limit}] {${POST_FIELDS}}`,
      ),
    ]);

    const totalPages = Math.ceil(total / limit);
    return {
      data: (data ?? []).map(mapPost),
      meta: { total, page, limit, totalPages },
    };
  } catch (error) {
    console.error("Failed to fetch blog posts from Sanity:", error);
    return { data: [], meta: { total: 0, page: 1, limit, totalPages: 0 } };
  }
}

export async function getPostBySlug(slug: string): Promise<BlogPost | null> {
  const client = getClient();
  if (!client) return null;

  try {
    const raw = await client.fetch<any>(
      `*[_type == "blogPost" && slug.current == $slug && ${PUBLISHED_FILTER}][0] {${POST_FIELDS}}`,
      { slug },
    );
    return raw ? mapPost(raw) : null;
  } catch (error) {
    console.error("Failed to fetch blog post by slug:", error);
    return null;
  }
}

export async function getCategories(): Promise<Category[]> {
  const client = getClient();
  if (!client) return [];

  try {
    const data = await client.fetch<any[]>(
      `*[_type == "category"] | order(title asc) {
        _id,
        title,
        "slug": slug.current,
        description,
        color,
        _createdAt,
        _updatedAt,
        "postCount": count(*[_type == "blogPost" && references(^._id) && ${PUBLISHED_FILTER}])
      }`,
    );

    return (data ?? []).map((c: any) => ({
      id: c._id,
      title: c.title ?? "",
      slug: c.slug ?? "",
      description: c.description ?? "",
      color: c.color ?? "",
      icon: "",
      order: 0,
      postCount: c.postCount ?? 0,
      createdAt: c._createdAt ?? "",
      updatedAt: c._updatedAt ?? "",
    }));
  } catch (error) {
    console.error("Failed to fetch categories from Sanity:", error);
    return [];
  }
}

export async function getAuthors(): Promise<Author[]> {
  const client = getClient();
  if (!client) return [];

  try {
    const data = await client.fetch<any[]>(
      `*[_type == "author"] | order(name asc) {
        _id,
        name,
        "slug": slug.current,
        bio,
        role,
        twitter,
        github,
        "photo": { "url": photo.asset->url },
        _createdAt,
        _updatedAt
      }`,
    );

    return (data ?? []).map((a: any) => ({
      id: a._id,
      name: a.name ?? "",
      slug: a.slug ?? "",
      bio: a.bio ?? undefined,
      role: a.role ?? undefined,
      photo: a.photo ?? undefined,
      active: true,
      createdAt: a._createdAt ?? "",
      updatedAt: a._updatedAt ?? "",
    }));
  } catch (error) {
    console.error("Failed to fetch authors from Sanity:", error);
    return [];
  }
}

export async function getPostsByCategory(
  categorySlug: string,
  params?: Omit<PostsQuery, "category">,
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, category: categorySlug });
}

export async function getPostsByAuthor(
  authorSlug: string,
  params?: Omit<PostsQuery, "author">,
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, author: authorSlug });
}

export async function searchPosts(
  query: string,
  params?: Omit<PostsQuery, "search">,
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, search: query });
}

export async function getRelatedPosts(
  currentPostId: string,
  categorySlug?: string | null,
  limit: number = 3,
): Promise<BlogPost[]> {
  const client = getClient();
  if (!client) return [];

  try {
    if (categorySlug) {
      const data = await client.fetch<any[]>(
        `*[_type == "blogPost" && _id != $id && category->slug.current == $catSlug && ${PUBLISHED_FILTER}] | order(publishedAt desc) [0...${limit}] {${POST_FIELDS}}`,
        { id: currentPostId, catSlug: categorySlug },
      );
      if (data && data.length >= limit) {
        return data.map(mapPost);
      }
    }

    const data = await client.fetch<any[]>(
      `*[_type == "blogPost" && _id != $id && ${PUBLISHED_FILTER}] | order(publishedAt desc) [0...${limit + 5}] {${POST_FIELDS}}`,
      { id: currentPostId },
    );
    return (data ?? []).slice(0, limit).map(mapPost);
  } catch (error) {
    console.error("Failed to fetch related posts:", error);
    return [];
  }
}

export async function getAllPosts(limit: number = 1000): Promise<BlogPost[]> {
  const client = getClient();
  if (!client) return [];

  try {
    const data = await client.fetch<any[]>(
      `*[_type == "blogPost" && ${PUBLISHED_FILTER}] | order(publishedAt desc) [0...${limit}] {${POST_FIELDS}}`,
    );
    return (data ?? []).map(mapPost);
  } catch (error) {
    console.error("Failed to fetch all posts:", error);
    return [];
  }
}
