import { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";

interface Heading {
  id: string;
  text: string;
  level: number;
}

interface DocsTableOfContentsProps {
  content: string;
}

export function DocsTableOfContents({ content }: DocsTableOfContentsProps) {
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [activeId, setActiveId] = useState<string>("");
  const location = useLocation();

  // Parse headings from markdown content
  useEffect(() => {
    if (!content) {
      setHeadings([]);
      return;
    }

    // Extract headings from markdown content
    // Match lines starting with # ## ### etc.
    const headingRegex = /^(#{1,3})\s+(.+)$/gm;
    const matches: Heading[] = [];
    let match;

    while ((match = headingRegex.exec(content)) !== null) {
      const level = match[1].length;
      const text = match[2].trim();
      // Create an ID from the text (similar to how markdown parsers do it)
      const id = text
        .toLowerCase()
        .replace(/[^\w\s-]/g, "")
        .replace(/\s+/g, "-");

      matches.push({ id, text, level });
    }

    setHeadings(matches);
  }, [content]);

  // Intersection Observer to track active heading
  useEffect(() => {
    if (headings.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        });
      },
      {
        rootMargin: "-100px 0px -60% 0px",
        threshold: 0,
      }
    );

    // Observe all heading elements
    headings.forEach((heading) => {
      const element = document.getElementById(heading.id);
      if (element) {
        observer.observe(element);
      }
    });

    return () => observer.disconnect();
  }, [headings, location.pathname]);

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      const offset = 100; // Account for sticky header
      const top = element.getBoundingClientRect().top + window.scrollY - offset;
      window.scrollTo({ top, behavior: "smooth" });
    }
  };

  if (headings.length === 0) {
    return (
      <div className="py-6 px-4">
        <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-4">
          On this page
        </h3>
        <p className="text-sm text-text-muted">No headings found</p>
      </div>
    );
  }

  return (
    <div className="py-6 px-4">
      <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-4">
        On this page
      </h3>
      <nav className="space-y-1">
        {headings.map((heading, index) => {
          const isActive = activeId === heading.id;
          const indentClass = heading.level === 1
            ? ""
            : heading.level === 2
              ? "ml-3"
              : "ml-6";

          return (
            <a
              key={`${heading.id}-${index}`}
              href={`#${heading.id}`}
              onClick={(e) => handleClick(e, heading.id)}
              className={`
                block py-1 text-sm transition-colors duration-200
                ${indentClass}
                ${isActive
                  ? "text-brand-500 font-medium"
                  : "text-text-muted hover:text-text-secondary"
                }
                ${heading.level === 1 ? "font-medium" : ""}
              `}
            >
              {heading.text}
            </a>
          );
        })}
      </nav>
    </div>
  );
}
