import rss from '@astrojs/rss';
import { posts } from '~/collections';
export function GET(context) {
  return rss({
    title: 'kmsec.uk',
    description: '(mainly) a security blog',

    site: context.site,
    // Array of `<item>`s in output xml
    // See "Generating items" section for examples using content collections and glob imports
    items: posts.map(post => ({
        title: post.data.title,
        description: post.data.description,
        link: "https://kmsec.uk/blog/" + post.id,
        pubDate: post.data.date
    })),
    // (optional) inject custom xml
    customData: `<language>en</language>`,
  });
}