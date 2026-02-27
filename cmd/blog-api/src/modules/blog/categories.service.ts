import { Injectable, NotFoundException } from '@nestjs/common';
import { eq, asc } from 'drizzle-orm';
import { DatabaseService } from '../../database/database.service';
import { categories } from '../../db/schema/index';
import { CreateCategoryDto, UpdateCategoryDto } from './dto/category.dto';

@Injectable()
export class CategoriesService {
  constructor(private databaseService: DatabaseService) {}

  async findAll() {
    return this.databaseService.getDatabase()
      .select()
      .from(categories)
      .orderBy(asc(categories.order), asc(categories.title));
  }

  async findById(id: string) {
    const [category] = await this.databaseService.getDatabase()
      .select()
      .from(categories)
      .where(eq(categories.id, id))
      .limit(1);

    if (!category) {
      throw new NotFoundException(`Category with ID "${id}" not found`);
    }

    return category;
  }

  async findBySlug(slug: string) {
    const [category] = await this.databaseService.getDatabase()
      .select()
      .from(categories)
      .where(eq(categories.slug, slug))
      .limit(1);

    if (!category) {
      throw new NotFoundException(`Category with slug "${slug}" not found`);
    }

    return category;
  }

  async create(dto: CreateCategoryDto) {
    const now = new Date();

    const [category] = await this.databaseService.getDatabase()
      .insert(categories)
      .values({
        title: dto.title,
        slug: dto.slug,
        description: dto.description,
        color: dto.color,
        icon: dto.icon,
        order: dto.order || 0,
        createdAt: now,
        updatedAt: now,
      })
      .returning();

    return category;
  }

  async update(id: string, dto: UpdateCategoryDto) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(categories)
      .where(eq(categories.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Category with ID "${id}" not found`);
    }

    const [updated] = await this.databaseService.getDatabase()
      .update(categories)
      .set({
        ...(dto.title && { title: dto.title }),
        ...(dto.slug && { slug: dto.slug }),
        ...(dto.description !== undefined && { description: dto.description }),
        ...(dto.color !== undefined && { color: dto.color }),
        ...(dto.icon !== undefined && { icon: dto.icon }),
        ...(dto.order !== undefined && { order: dto.order }),
        updatedAt: new Date(),
      })
      .where(eq(categories.id, id))
      .returning();

    return updated;
  }

  async delete(id: string) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(categories)
      .where(eq(categories.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Category with ID "${id}" not found`);
    }

    await this.databaseService.getDatabase()
      .delete(categories)
      .where(eq(categories.id, id));

    return { success: true, message: 'Category deleted successfully' };
  }
}
