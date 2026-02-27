import { Module } from '@nestjs/common';
import { BlogController } from './blog.controller';
import { BlogService } from './blog.service';
import { BlogPostsService } from './blog-posts.service';
import { CategoriesService } from './categories.service';
import { AuthorsService } from './authors.service';

@Module({
  controllers: [BlogController],
  providers: [BlogService, BlogPostsService, CategoriesService, AuthorsService],
  exports: [BlogService],
})
export class BlogModule {}
