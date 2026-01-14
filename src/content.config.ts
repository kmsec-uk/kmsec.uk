// 1. Import utilities from `astro:content`
import { defineCollection } from 'astro:content';

// 2. Import loader(s)
import { glob } from 'astro/loaders';

// 3. Import Zod
import { z } from 'astro/zod';

// 4. Define your collection(s)
const blog = defineCollection({
  loader: glob({ pattern: "**/*.{md,mdx}", base: "./src/content/blog" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    tags: z.array(z.string()),
    date: z.date({coerce: true})
  })
});

// 4. Define your collection(s)
const tools = defineCollection({
  loader: glob({ pattern: "*.json", base: "./src/content/tools" }),
  schema: ({image}) =>z.object({
    title: z.string(),
    description: z.string(),
    image: image(),
    link: z.string(),
    priority: z.number()
  })
});
// 5. Export a single `collections` object to register your collection(s)
export const collections = { blog, tools };