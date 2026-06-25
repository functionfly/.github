import React, { useEffect, useState } from "react";
import { Chamber, CornerBrace } from "./sc";

interface Author {
  name: string;
  slug: string;
  email?: string;
  photo?: { url: string } | null;
  bio?: string;
  role?: string;
}

interface Category {
  title: string;
  slug: string;
}

interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  heroImage?: { url: string; alt: string; caption?: string } | null;
  publishedAt?: string | null;
  author?: Author | null;
  category?: Category | null;
  tags?: string[];
}

interface BlogWidgetProps {
  limit?: number;
  title?: string;
}

// Use the blog site URL for links
const BLOG_SITE_URL = import.meta.env.PUBLIC_BLOG_SITE_URL || "https://blog.functionfly.com";

export const BlogWidget: React.FC<BlogWidgetProps> = ({ 
  limit = 3, 
  title = "Latest from the Blog" 
}) => {
  const [posts, setPosts] = useState<BlogPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPosts = async () => {
      try {
        // Fetch from Go backend API
        const response = await fetch(
          `/v1/blog/posts?limit=${limit}&status=published`
        );
        if (!response.ok) {
          throw new Error(`Failed to fetch posts: ${response.status}`);
        }
        const data = await response.json();
        setPosts(data.data || []);
      } catch (err) {
        console.error("Failed to fetch blog posts:", err);
        setError("Unable to load blog posts");
      } finally {
        setLoading(false);
      }
    };

    fetchPosts();
  }, [limit]);

  const formatDate = (dateString?: string | null) => {
    if (!dateString) return "";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  // Build the full blog URL - blog uses /post/{slug} format
  const getBlogUrl = (slug: string) => {
    return `${BLOG_SITE_URL}/post/${slug}`;
  };

  if (loading) {
    return (
      <div className="ff-blog-widget">
        <div className="ff-blog-widget-loading">
          <div className="ff-blog-widget-spinner" />
          <span>Loading posts...</span>
        </div>
      </div>
    );
  }

  if (error || posts.length === 0) {
    return (
      <div className="ff-blog-widget">
        <Chamber ribs className="ff-blog-widget-card ff-blog-widget-empty">
          <p className="ff-blog-widget-empty-text">
            {error || "No blog posts available"}
          </p>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="ff-blog-widget">
      <div className="ff-blog-widget-header">
        <h3 className="ff-blog-widget-title">{title}</h3>
        <a href={`${BLOG_SITE_URL}`} className="ff-blog-widget-link" >
          View all
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        </a>
      </div>
      
      <div className="ff-blog-widget-grid">
        {posts.slice(0, limit).map((post, index) => (
          <a 
            key={post.id} 
            href={getBlogUrl(post.slug)}
            className="ff-blog-widget-card-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Chamber ribs className="ff-blog-widget-card">
              <CornerBrace position="tl" />
              {index === 0 && <CornerBrace position="br" />}
              
              {post.heroImage?.url && (
                <div className="ff-blog-widget-image">
                  <img 
                    src={post.heroImage.url} 
                    alt={post.heroImage.alt || post.title}
                  />
                </div>
              )}
              
              <div className="ff-blog-widget-content">
                {post.category && (
                  <span className="ff-blog-widget-category">
                    {post.category.title}
                  </span>
                )}
                
                <h4 className="ff-blog-widget-post-title">{post.title}</h4>
                
                <p className="ff-blog-widget-excerpt">
                  {post.description}
                </p>
                
                <div className="ff-blog-widget-meta">
                  {post.author?.photo?.url && (
                    <img 
                      src={post.author.photo.url} 
                      alt={post.author.name}
                      className="ff-blog-widget-avatar"
                    />
                  )}
                  <span className="ff-blog-widget-author">
                    {post.author?.name || "Anonymous"}
                  </span>
                  <span className="ff-blog-widget-separator">·</span>
                  <span className="ff-blog-widget-date">
                    {formatDate(post.publishedAt)}
                  </span>
                </div>
              </div>
            </Chamber>
          </a>
        ))}
      </div>
    </div>
  );
};

export default BlogWidget;
