import { Injectable, NotFoundException } from '@nestjs/common';
import { and, eq, desc, like, or, inArray } from 'drizzle-orm';
import { DatabaseService } from '../../database/database.service';
import { blogPosts, authors, categories, NewBlogPost } from '../../db/schema/index';
import { CreateBlogPostDto, UpdateBlogPostDto, BlogPostQueryDto } from './dto/blog-post.dto';

@Injectable()
export class BlogPostsService {
  constructor(private databaseService: DatabaseService) {}

  async findAll(query: BlogPostQueryDto) {
    const { page = 1, limit = 10, status, category, author, search, tag } = query;
    const offset = (page - 1) * limit;

    // Build conditions
    const conditions = [];

    // Only show published posts for public API
    if (status) {
      conditions.push(eq(blogPosts.status, status as any));
    } else {
      // Default to published for public access
      conditions.push(or(
        eq(blogPosts.status, 'published' as any),
        eq(blogPosts.status, 'scheduled' as any)
      ));
    }

    // Category filter by slug
    if (category) {
      const categoryResult = await this.databaseService.getDatabase()
        .select({ id: categories.id })
        .from(categories)
        .where(eq(categories.slug, category))
        .limit(1);

      if (categoryResult.length > 0) {
        conditions.push(eq(blogPosts.categoryId, categoryResult[0].id));
      }
    }

    // Author filter by slug
    if (author) {
      const authorResult = await this.databaseService.getDatabase()
        .select({ id: authors.id })
        .from(authors)
        .where(eq(authors.slug, author))
        .limit(1);

      if (authorResult.length > 0) {
        conditions.push(eq(blogPosts.authorId, authorResult[0].id));
      }
    }

    // Tag filter
    if (tag) {
      conditions.push(
        like(blogPosts.tags, `%${tag}%`)
      );
    }

    // Search functionality
    if (search && search.trim()) {
      const searchTerm = `%${search.trim()}%`;
      conditions.push(
        or(
          like(blogPosts.title, searchTerm),
          like(blogPosts.description, searchTerm),
          like(blogPosts.body, searchTerm),
          like(blogPosts.tags, searchTerm)
        )
      );
    }

    // Get total count
    const totalResult = await this.databaseService.getDatabase()
      .select({ count: blogPosts.id })
      .from(blogPosts)
      .where(and(...conditions));

    const total = totalResult.length;

    // Get posts with pagination
    const posts = await this.databaseService.getDatabase()
      .select()
      .from(blogPosts)
      .where(and(...conditions))
      .orderBy(desc(blogPosts.publishedAt))
      .limit(limit)
      .offset(offset);

    // Fetch author and category for each post
    const postsWithRelations = await Promise.all(
      posts.map(async (post: any) => {
        let authorData = null;
        let categoryData = null;

        if (post.authorId) {
          const [author] = await this.databaseService.getDatabase()
            .select()
            .from(authors)
            .where(eq(authors.id, post.authorId))
            .limit(1);
          authorData = author;
        }

        if (post.categoryId) {
          const [category] = await this.databaseService.getDatabase()
            .select()
            .from(categories)
            .where(eq(categories.id, post.categoryId))
            .limit(1);
          categoryData = category;
        }

        return {
          ...post,
          author: authorData ? { name: authorData.name, slug: authorData.slug } : null,
          category: categoryData ? { title: categoryData.title, slug: categoryData.slug } : null,
        };
      })
    );

    return {
      data: postsWithRelations,
      meta: {
        total,
        page,
        limit,
        totalPages: Math.ceil(total / limit),
        ...(search && { search }),
      },
    };
  }

  async findBySlug(slug: string) {
    const [post] = await this.databaseService.getDatabase()
      .select()
      .from(blogPosts)
      .where(and(
        eq(blogPosts.slug, slug),
        or(
          eq(blogPosts.status, 'published' as any),
          eq(blogPosts.status, 'scheduled' as any)
        )
      ))
      .limit(1);

    if (!post) {
      throw new NotFoundException(`Post with slug "${slug}" not found`);
    }

    // Fetch author
    let authorData = null;
    if (post.authorId) {
      const [author] = await this.databaseService.getDatabase()
        .select()
        .from(authors)
        .where(eq(authors.id, post.authorId))
        .limit(1);
      authorData = author;
    }

    // Fetch category
    let categoryData = null;
    if (post.categoryId) {
      const [category] = await this.databaseService.getDatabase()
        .select()
        .from(categories)
        .where(eq(categories.id, post.categoryId))
        .limit(1);
      categoryData = category;
    }

    return {
      ...post,
      author: authorData ? {
        name: authorData.name,
        slug: authorData.slug,
        photo: authorData.photo,
        bio: authorData.bio,
        socialLinks: authorData.socialLinks,
      } : null,
      category: categoryData ? {
        title: categoryData.title,
        slug: categoryData.slug,
      } : null,
    };
  }

  async create(dto: CreateBlogPostDto) {
    const now = new Date();

    const insertData: NewBlogPost = {
      title: dto.title,
      slug: dto.slug,
      description: dto.description || '',
      body: dto.body,
      authorId: dto.authorId,
      categoryId: dto.categoryId,
      tags: dto.tags || [],
      heroImage: null,
      status: (dto.status as any) || 'draft',
      publishedAt: dto.publishedAt ? new Date(dto.publishedAt) : null,
      scheduledAt: dto.scheduledAt ? new Date(dto.scheduledAt) : null,
      updatedAt: now,
      createdAt: now,
      seoTitle: dto.seoTitle,
      seoDescription: dto.seoDescription,
      keywords: dto.keywords,
      canonicalUrl: dto.canonicalUrl,
      ogImage: null,
      campaign: dto.campaign,
      ownerId: dto.ownerId,
    };

    const [post] = await this.databaseService.getDatabase()
      .insert(blogPosts)
      .values(insertData)
      .returning();

    return post;
  }

  async update(id: string, dto: UpdateBlogPostDto) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(blogPosts)
      .where(eq(blogPosts.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Post with ID "${id}" not found`);
    }

    const [updated] = await this.databaseService.getDatabase()
      .update(blogPosts)
      .set({
        ...(dto.title && { title: dto.title }),
        ...(dto.slug && { slug: dto.slug }),
        ...(dto.description && { description: dto.description }),
        ...(dto.body && { body: dto.body }),
        ...(dto.authorId !== undefined && { authorId: dto.authorId }),
        ...(dto.categoryId !== undefined && { categoryId: dto.categoryId }),
        ...(dto.tags && { tags: dto.tags }),
        ...(dto.heroImage && { heroImage: dto.heroImage }),
        ...(dto.status && { status: dto.status }),
        ...(dto.publishedAt !== undefined && { publishedAt: dto.publishedAt }),
        ...(dto.scheduledAt !== undefined && { scheduledAt: dto.scheduledAt }),
        ...(dto.seoTitle !== undefined && { seoTitle: dto.seoTitle }),
        ...(dto.seoDescription !== undefined && { seoDescription: dto.seoDescription }),
        ...(dto.keywords && { keywords: dto.keywords }),
        ...(dto.canonicalUrl !== undefined && { canonicalUrl: dto.canonicalUrl }),
        ...(dto.ogImage && { ogImage: dto.ogImage }),
        ...(dto.campaign !== undefined && { campaign: dto.campaign }),
        ...(dto.ownerId !== undefined && { ownerId: dto.ownerId }),
        updatedAt: new Date(),
      })
      .where(eq(blogPosts.id, id))
      .returning();

    return updated;
  }

  async delete(id: string) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(blogPosts)
      .where(eq(blogPosts.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Post with ID "${id}" not found`);
    }

    await this.databaseService.getDatabase()
      .delete(blogPosts)
      .where(eq(blogPosts.id, id));

    return { success: true, message: 'Post deleted successfully' };
  }
}

