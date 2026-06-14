import type { APIRoute } from "astro";
import { getAllPosts } from "../lib/api";
import { BLOG_SITE_URL, SITE_DESCRIPTION, SITE_NAME } from "../lib/constants";
import { slateBodyToHtml } from "../lib/slate-to-html";

export const GET: APIRoute = async () => {
  const posts = await getAllPosts(50);

  const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>${SITE_NAME}</title>
    <link>${BLOG_SITE_URL}</link>
    <description>${SITE_DESCRIPTION}</description>
    <language>en-US</language>
    <lastBuildDate>${new Date().toUTCString()}</lastBuildDate>
    <atom:link href="${BLOG_SITE_URL}/rss.xml" rel="self" type="application/rss+xml" />
    ${posts
      .filter((post) => post.status === "published" && post.publishedAt)
      .map((post) => {
        const postUrl = `${BLOG_SITE_URL}/${post.slug}`;
        const pubDate = new Date(post.publishedAt!).toUTCString();
        // Convert Slate body to HTML for content
        const bodyHtml = slateBodyToHtml(post.body);
        // Strip HTML tags and decode entities for description
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = post.description;
        const plainText = tempDiv.textContent || tempDiv.innerText || '';
        const plainDesc = decodeHtmlEntities(plainText);

        return `
    <item>
      <title>${escapeXml(post.title)}</title>
      <link>${postUrl}</link>
      <guid isPermaLink="true">${postUrl}</guid>
      <pubDate>${pubDate}</pubDate>
      <description>${escapeXml(plainDesc)}</description>
      ${post.author ? `<author>${escapeXml(post.author.email || post.author.name)}</author>` : ""}
      ${post.tags ? post.tags.map((tag) => `<category>${escapeXml(tag)}</category>`).join("") : ""}
      <content:encoded><![CDATA[${bodyHtml}]]></content:encoded>
    </item>`;
      })
      .join("")}
  </channel>
</rss>`;

  return new Response(rss, {
    headers: {
      "Content-Type": "application/xml; charset=utf-8",
      "Cache-Control": "public, max-age=3600", // Cache for 1 hour
    },
  });
};

function escapeXml(unsafe: string): string {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

function decodeHtmlEntities(str: string): string {
  return str
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&#x27;/g, "'");
}
