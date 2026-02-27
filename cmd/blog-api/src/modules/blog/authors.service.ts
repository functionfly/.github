import { Injectable, NotFoundException } from '@nestjs/common';
import { eq, asc } from 'drizzle-orm';
import { DatabaseService } from '../../database/database.service';
import { authors } from '../../db/schema/index';
import { CreateAuthorDto, UpdateAuthorDto } from './dto/author.dto';

@Injectable()
export class AuthorsService {
  constructor(private databaseService: DatabaseService) {}

  async findAll() {
    return this.databaseService.getDatabase()
      .select()
      .from(authors)
      .orderBy(asc(authors.name));
  }

  async findActive() {
    return this.databaseService.getDatabase()
      .select()
      .from(authors)
      .where(eq(authors.active, true))
      .orderBy(asc(authors.name));
  }

  async findById(id: string) {
    const [author] = await this.databaseService.getDatabase()
      .select()
      .from(authors)
      .where(eq(authors.id, id))
      .limit(1);

    if (!author) {
      throw new NotFoundException(`Author with ID "${id}" not found`);
    }

    return author;
  }

  async findBySlug(slug: string) {
    const [author] = await this.databaseService.getDatabase()
      .select()
      .from(authors)
      .where(eq(authors.slug, slug))
      .limit(1);

    if (!author) {
      throw new NotFoundException(`Author with slug "${slug}" not found`);
    }

    return author;
  }

  async create(dto: CreateAuthorDto) {
    const now = new Date();

    const [author] = await this.databaseService.getDatabase()
      .insert(authors)
      .values({
        name: dto.name,
        slug: dto.slug,
        bio: dto.bio,
        email: dto.email,
        website: dto.website,
        photo: dto.photo,
        socialLinks: dto.socialLinks,
        role: dto.role,
        active: dto.active ?? true,
        createdAt: now,
        updatedAt: now,
      })
      .returning();

    return author;
  }

  async update(id: string, dto: UpdateAuthorDto) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(authors)
      .where(eq(authors.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Author with ID "${id}" not found`);
    }

    const [updated] = await this.databaseService.getDatabase()
      .update(authors)
      .set({
        ...(dto.name && { name: dto.name }),
        ...(dto.slug && { slug: dto.slug }),
        ...(dto.bio !== undefined && { bio: dto.bio }),
        ...(dto.email !== undefined && { email: dto.email }),
        ...(dto.website !== undefined && { website: dto.website }),
        ...(dto.photo !== undefined && { photo: dto.photo }),
        ...(dto.socialLinks && { socialLinks: dto.socialLinks }),
        ...(dto.role !== undefined && { role: dto.role }),
        ...(dto.active !== undefined && { active: dto.active }),
        updatedAt: new Date(),
      })
      .where(eq(authors.id, id))
      .returning();

    return updated;
  }

  async delete(id: string) {
    const [existing] = await this.databaseService.getDatabase()
      .select()
      .from(authors)
      .where(eq(authors.id, id))
      .limit(1);

    if (!existing) {
      throw new NotFoundException(`Author with ID "${id}" not found`);
    }

    await this.databaseService.getDatabase()
      .delete(authors)
      .where(eq(authors.id, id));

    return { success: true, message: 'Author deleted successfully' };
  }
}
