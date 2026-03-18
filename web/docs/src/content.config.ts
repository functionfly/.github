import { defineCollection } from 'astro:content';
import { z } from 'zod';

const docsSchema = z.object({
  title: z.string(),
  description: z.string().optional(),
  order: z.number().optional(),
});

export const collections = {
  docs: defineCollection({
    type: 'content',
    schema: docsSchema,
  }),
};
