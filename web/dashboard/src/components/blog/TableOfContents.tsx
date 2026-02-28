'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion } from 'framer-motion';
import { List, X, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from '@/components/ui/sheet';

interface TOCItem {
  id: string;
  text: string;
  level: number;
}

interface TableOfContentsProps {
  content?: string;
  selector?: string;
}

export function TableOfContents({ content, selector = '.blog-post-prose' }: TableOfContentsProps) {
  const [headings, setHeadings] = useState<TOCItem[]>([]);
  const [activeId, setActiveId] = useState<string>('');
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  // Extract headings from content or DOM
  useEffect(() => {
    const extractHeadings = () => {
      const items: TOCItem[] = [];

      if (content) {
        // Extract from markdown/HTML content
        const headingRegex = /^(#{1,6})\s+(.+)$/gm;
        let match;
        while ((match = headingRegex.exec(content)) !== null) {
          const level = match[1].length;
          const text = match[2].trim();
          const id = text
            .toLowerCase()
            .replace(/[^a-z0-9]+/g, '-')
            .replace(/(^-|-$)/g, '');
          items.push({ id, text, level });
        }
      } else {
        // Extract from DOM
        const container = document.querySelector(selector);
        if (container) {
          const elements = container.querySelectorAll('h1, h2, h3, h4, h5, h6');
          elements.forEach((el) => {
            const text = el.textContent?.trim() || '';
            const id = el.id || text
              .toLowerCase()
              .replace(/[^a-z0-9]+/g, '-')
              .replace(/(^-|-$)/g, '');
            
            // Ensure element has an ID
            if (!el.id) {
              el.id = id;
            }
            
            items.push({
              id,
              text,
              level: parseInt(el.tagName.substring(1)),
            });
          });
        }
      }

      setHeadings(items);
    };

    // Wait for content to render
    const timer = setTimeout(extractHeadings, 100);
    return () => clearTimeout(timer);
  }, [content, selector]);

  // Track active heading on scroll
  useEffect(() => {
    const handleScroll = () => {
      if (headings.length === 0) return;

      const scrollPosition = window.scrollY + 100;

      // Find the current heading
      for (let i = headings.length - 1; i >= 0; i--) {
        const element = document.getElementById(headings[i].id);
        if (element && element.offsetTop <= scrollPosition) {
          setActiveId(headings[i].id);
          return;
        }
      }

      setActiveId(headings[0]?.id || '');
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    handleScroll(); // Initial check

    return () => window.removeEventListener('scroll', handleScroll);
  }, [headings]);

  const scrollToHeading = useCallback((id: string) => {
    const element = document.getElementById(id);
    if (element) {
      const offset = 80; // Account for fixed header
      const elementPosition = element.getBoundingClientRect().top + window.scrollY;
      window.scrollTo({
        top: elementPosition - offset,
        behavior: 'smooth',
      });
    }
    setIsMobileOpen(false);
  }, []);

  if (headings.length < 2) {
    return null;
  }

  // Desktop TOC
  const DesktopTOC = () => (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      className="hidden lg:block sticky top-24 max-h-[calc(100vh-6rem)] overflow-y-auto"
    >
      <div className="pr-4">
        <h4 className="text-sm font-semibold mb-4 text-foreground">On this page</h4>
        <nav aria-label="Table of contents">
          <ul className="space-y-1">
            {headings.map((heading) => (
              <li key={heading.id}>
                <button
                  onClick={() => scrollToHeading(heading.id)}
                  className={`
                    block w-full text-left text-sm py-1.5 px-3 rounded-md transition-all duration-200
                    ${activeId === heading.id
                      ? 'bg-brand-500/10 text-brand-600 dark:text-brand-400 font-medium'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
                    }
                  `}
                  style={{ paddingLeft: `${(heading.level - 1) * 12 + 12}px` }}
                  aria-current={activeId === heading.id ? 'location' : undefined}
                >
                  {heading.text}
                </button>
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </motion.div>
  );

  // Mobile TOC (Sheet)
  const MobileTOC = () => (
    <Sheet open={isMobileOpen} onOpenChange={setIsMobileOpen}>
      <SheetTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="lg:hidden fixed bottom-6 right-6 z-40 rounded-full shadow-lg"
        >
          <List className="h-4 w-4 mr-2" />
          Table of Contents
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-[300px] sm:w-[350px]">
        <SheetTitle className="mb-4">On this page</SheetTitle>
        <nav aria-label="Table of contents">
          <ul className="space-y-1">
            {headings.map((heading) => (
              <li key={heading.id}>
                <button
                  onClick={() => scrollToHeading(heading.id)}
                  className={`
                    flex items-center gap-2 w-full text-left text-sm py-2 px-3 rounded-md transition-all duration-200
                    ${activeId === heading.id
                      ? 'bg-brand-500/10 text-brand-600 dark:text-brand-400 font-medium'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
                    }
                  `}
                  style={{ paddingLeft: `${(heading.level - 1) * 12 + 12}px` }}
                >
                  <ChevronRight className="h-3 w-3 shrink-0" />
                  <span className="line-clamp-1">{heading.text}</span>
                </button>
              </li>
            ))}
          </ul>
        </nav>
      </SheetContent>
    </Sheet>
  );

  return (
    <>
      <DesktopTOC />
      <MobileTOC />
    </>
  );
}
