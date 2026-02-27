import { Injectable, NotFoundException } from '@nestjs/common';
import { BlogPostsService } from './blog-posts.service';
import { CategoriesService } from './categories.service';
import { AuthorsService } from './authors.service';
import { CreateBlogPostDto, UpdateBlogPostDto, BlogPostQueryDto } from './dto/blog-post.dto';
import { CreateCategoryDto, UpdateCategoryDto } from './dto/category.dto';
import { CreateAuthorDto, UpdateAuthorDto } from './dto/author.dto';

@Injectable()
export class BlogService {
  constructor(
    private readonly blogPostsService: BlogPostsService,
    private readonly categoriesService: CategoriesService,
    private readonly authorsService: AuthorsService,
  ) {}

  // Blog Posts
  async getPosts(query: BlogPostQueryDto) {
    return this.blogPostsService.findAll(query);
  }

  async getPostBySlug(slug: string) {
    return this.blogPostsService.findBySlug(slug);
  }

  async createPost(dto: CreateBlogPostDto) {
    return this.blogPostsService.create(dto);
  }

  async updatePost(id: string, dto: UpdateBlogPostDto) {
    return this.blogPostsService.update(id, dto);
  }

  async deletePost(id: string) {
    return this.blogPostsService.delete(id);
  }

  // Categories
  async getCategories() {
    return this.categoriesService.findAll();
  }

  async createCategory(dto: CreateCategoryDto) {
    return this.categoriesService.create(dto);
  }

  async updateCategory(id: string, dto: UpdateCategoryDto) {
    return this.categoriesService.update(id, dto);
  }

  async deleteCategory(id: string) {
    return this.categoriesService.delete(id);
  }

  // Authors
  async getAuthors() {
    return this.authorsService.findAll();
  }

  async createAuthor(dto: CreateAuthorDto) {
    return this.authorsService.create(dto);
  }

  async updateAuthor(id: string, dto: UpdateAuthorDto) {
    return this.authorsService.update(id, dto);
  }

  async deleteAuthor(id: string) {
    return this.authorsService.delete(id);
  }
}
