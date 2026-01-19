// @ts-check
import { defineConfig } from 'astro/config';

// import cloudflare from '@astrojs/cloudflare';

import tailwindcss from '@tailwindcss/vite';

import mdx from '@astrojs/mdx';

import rehypeExternalLinks from 'rehype-external-links';

// https://astro.build/config
export default defineConfig({
  // adapter: cloudflare({
  //   platformProxy: {
  //     enabled: true
  //   },

  //   imageService: "cloudflare"
  // }),
  output: 'static',
  vite: {
    plugins: [tailwindcss()]
  },

  integrations: [mdx()],
    markdown: {
    rehypePlugins: [
      [
        rehypeExternalLinks,
        { target: '_blank', rel: [] }
      ],
    ]
  },
});