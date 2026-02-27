import { Controller, Get, Post, Put, Delete, Body, Param, Query, UseGuards } from '@nestjs/common';
import { ApiTags, ApiBearerAuth, ApiOperation, ApiQuery } from '@nestjs/swagger';
import { Throttle } from '@nestjs/throttler';
import { BlogService } from './blog.service';
import { CreateBlogPostDto, UpdateBlogPostDto, BlogPostQueryDto } from './dto/blog-post.dto';
import { CreateCategoryDto, UpdateCategoryDto } from './dto/category.dto';
import { CreateAuthorDto, UpdateAuthorDto } from './dto/author.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { RolesGuard } from '../auth/guards/roles.guard';
import { Roles } from '../../common/decorators/roles.decorator';

@ApiTags('Blog')
@Controller('blog')
export class BlogController {
  constructor(private readonly blogService: BlogService) {}

  // ============ Blog Posts ============

  @Get('posts')
  @ApiOperation({ summary: 'Get all published blog posts' })
  @Throttle({ short: { limit: 30, ttl: 1000 } })
  async getPosts(@Query() query: BlogPostQueryDto) {
    return this.blogService.getPosts(query);
  }

  @Get('posts/:slug')
  @ApiOperation({ summary: 'Get blog post by slug' })
  @Throttle({ short: { limit: 30, ttl: 1000 } })
  async getPostBySlug(@Param('slug') slug: string) {
    return this.blogService.getPostBySlug(slug);
  }

  @Post('posts')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Create a new blog post (admin)' })
  async createPost(@Body() dto: CreateBlogPostDto) {
    return this.blogService.createPost(dto);
  }

  @Put('posts/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Update a blog post (admin)' })
  async updatePost(@Param('id') id: string, @Body() dto: UpdateBlogPostDto) {
    return this.blogService.updatePost(id, dto);
  }

  @Delete('posts/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Delete a blog post (admin)' })
  async deletePost(@Param('id') id: string) {
    return this.blogService.deletePost(id);
  }

  // ============ Categories ============

  @Get('categories')
  @ApiOperation({ summary: 'Get all categories' })
  @Throttle({ short: { limit: 30, ttl: 1000 } })
  async getCategories() {
    return this.blogService.getCategories();
  }

  @Post('categories')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Create a category (admin)' })
  async createCategory(@Body() dto: CreateCategoryDto) {
    return this.blogService.createCategory(dto);
  }

  @Put('categories/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Update a category (admin)' })
  async updateCategory(@Param('id') id: string, @Body() dto: UpdateCategoryDto) {
    return this.blogService.updateCategory(id, dto);
  }

  @Delete('categories/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Delete a category (admin)' })
  async deleteCategory(@Param('id') id: string) {
    return this.blogService.deleteCategory(id);
  }

  // ============ Authors ============

  @Get('authors')
  @ApiOperation({ summary: 'Get all authors' })
  @Throttle({ short: { limit: 30, ttl: 1000 } })
  async getAuthors() {
    return this.blogService.getAuthors();
  }

  @Post('authors')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Create an author (admin)' })
  async createAuthor(@Body() dto: CreateAuthorDto) {
    return this.blogService.createAuthor(dto);
  }

  @Put('authors/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin', 'editor')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Update an author (admin)' })
  async updateAuthor(@Param('id') id: string, @Body() dto: UpdateAuthorDto) {
    return this.blogService.updateAuthor(id, dto);
  }

  @Delete('authors/:id')
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles('admin')
  @ApiBearerAuth()
  @ApiOperation({ summary: 'Delete an author (admin)' })
  async deleteAuthor(@Param('id') id: string) {
    return this.blogService.deleteAuthor(id);
  }
}
