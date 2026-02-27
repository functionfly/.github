import { BlogPost } from '@/api/content';

/**
 * Calculate estimated reading time based on content length
 * @param content - The text content to analyze
 * @param wordsPerMinute - Average reading speed (default: 200 words/minute)
 * @returns Estimated reading time in minutes
 */
export function calculateReadingTime(content: string, wordsPerMinute = 200): number {
  const wordCount = content.trim().split(/\s+/).length;
  return Math.max(1, Math.ceil(wordCount / wordsPerMinute));
}

/**
 * Format reading time for display
 * @param minutes - Reading time in minutes
 * @returns Formatted reading time string
 */
export function formatReadingTime(minutes: number): string {
  if (minutes < 1) return '< 1 min read';
  if (minutes === 1) return '1 min read';
  return `${minutes} min read`;
}

/**
 * Get relative time string (e.g., "2 days ago", "3 hours ago")
 * @param date - Date string or Date object
 * @returns Relative time string
 */
export function getRelativeTime(date: string | Date): string {
  const now = new Date();
  const past = new Date(date);
  const diffMs = now.getTime() - past.getTime();

  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMinutes < 1) return 'Just now';
  if (diffMinutes < 60) return `${diffMinutes} minute${diffMinutes === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  if (diffDays < 30) {
    const weeks = Math.floor(diffDays / 7);
    return `${weeks} week${weeks === 1 ? '' : 's'} ago`;
  }
  if (diffDays < 365) {
    const months = Math.floor(diffDays / 30);
    return `${months} month${months === 1 ? '' : 's'} ago`;
  }

  const years = Math.floor(diffDays / 365);
  return `${years} year${years === 1 ? '' : 's'} ago`;
}

/**
 * Get the featured post from a list of blog posts
 * @param posts - Array of blog posts
 * @returns The featured post (first post or most recent)
 */
export function getFeaturedPost(posts: BlogPost[]): BlogPost | null {
  if (!posts.length) return null;

  // For now, return the most recent published post as featured
  // In the future, this could be based on a featured flag or priority
  const publishedPosts = posts.filter(post => post.is_published);
  return publishedPosts.sort((a, b) =>
    new Date(b.published_at || b.created_at).getTime() -
    new Date(a.published_at || a.created_at).getTime()
  )[0] || null;
}

/**
 * Get posts excluding the featured post
 * @param posts - Array of blog posts
 * @param featuredPost - The featured post to exclude
 * @returns Array of posts excluding the featured post
 */
export function getRemainingPosts(posts: BlogPost[], featuredPost: BlogPost | null): BlogPost[] {
  if (!featuredPost) return posts;
  return posts.filter(post => post.id !== featuredPost.id);
}

/**
 * Generate author avatar placeholder or initials
 * @param authorName - The author's name
 * @returns Author avatar display (initials or placeholder)
 */
export function getAuthorAvatar(authorName: string): string {
  if (!authorName) return '?';

  const names = authorName.trim().split(' ');
  if (names.length >= 2) {
    return `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase();
  }

  return authorName.charAt(0).toUpperCase();
}