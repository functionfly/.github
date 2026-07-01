import { useEffect, useState, useCallback } from 'react';

export interface TocItem {
  id: string;
  label: string;
}

interface TableOfContentsProps {
  items: TocItem[];
}

export function TableOfContents({ items }: TableOfContentsProps) {
  const [activeId, setActiveId] = useState<string>('');

  const itemIds = items.map((i) => i.id).join(',');

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        });
      },
      { rootMargin: '-80px 0px -60% 0px', threshold: 0 }
    );

    items.forEach(({ id }) => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });

    return () => observer.disconnect();
  }, [itemIds]);

  if (items.length === 0) return null;

  const handleClick = useCallback((e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, []);

  return (
    <nav
      className="function-page-toc"
      aria-label="Page sections"
    >
      <ul className="function-page-toc-list">
        {items.map(({ id, label }) => (
          <li key={id}>
            <a
              href={`#${id}`}
              className={`function-page-toc-link ${activeId === id ? 'function-page-toc-link--active' : ''}`}
              aria-current={activeId === id ? 'true' : undefined}
              onClick={(e) => handleClick(e, id)}
            >
              {label}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}