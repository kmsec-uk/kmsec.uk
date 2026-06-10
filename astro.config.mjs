// @ts-check
import { defineConfig } from 'astro/config';
import { unified } from '@astrojs/markdown-remark';

// import cloudflare from '@astrojs/cloudflare';

import tailwindcss from '@tailwindcss/vite';

import mdx from '@astrojs/mdx';

import rehypeExternalLinks from 'rehype-external-links';

// https://astro.build/config
export default defineConfig({
  site: "https://kmsec.uk",
  // not using any SSR features for blog yet...
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

  integrations: [mdx({
    processor: unified({
          rehypePlugins: [
      [
        rehypeExternalLinks,
        { target: '_blank', rel: [] }
      ],
    ]
    })
  })],
});