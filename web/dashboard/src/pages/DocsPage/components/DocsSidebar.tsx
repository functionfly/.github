import { useState } from "react";
import { ChevronDown, ChevronRight, type LucideIcon } from "lucide-react";
import { type DocSection, type DocPage } from "../data/docs";

interface DocsSidebarProps {
  sections: DocSection[];
  currentPage: DocPage | null;
  currentSection: DocSection | null;
  onPageSelect: (slug: string) => void;
}

export function DocsSidebar({
  sections,
  currentPage,
  currentSection,
  onPageSelect
}: DocsSidebarProps) {
  // Track expanded sections
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => {
    // Start with current section expanded
    const initial = new Set<string>();
    if (currentSection) {
      initial.add(currentSection.id);
    } else {
      // Default expand first section
      initial.add(sections[0]?.id);
    }
    return initial;
  });

  const toggleSection = (sectionId: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev);
      if (next.has(sectionId)) {
        next.delete(sectionId);
      } else {
        next.add(sectionId);
      }
      return next;
    });
  };

  return (
    <div className="h-full overflow-y-auto py-4 px-3">
      {/* Section Navigation */}
      <nav className="space-y-1">
        {sections.map((section) => {
          const isExpanded = expandedSections.has(section.id);
          const isActiveSection = currentSection?.id === section.id;
          const Icon = section.icon;

          return (
            <div key={section.id} className="mb-2">
              {/* Section Header */}
              <button
                onClick={() => toggleSection(section.id)}
                className={`
                  w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium
                  transition-colors duration-200
                  ${isActiveSection
                    ? "text-brand-500 bg-brand-500/10"
                    : "text-text-secondary hover:text-text-primary hover:bg-bg-secondary"
                  }
                `}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                <span className="flex-1 text-left">{section.title}</span>
                {isExpanded ? (
                  <ChevronDown className="w-4 h-4 flex-shrink-0" />
                ) : (
                  <ChevronRight className="w-4 h-4 flex-shrink-0" />
                )}
              </button>

              {/* Section Pages */}
              {isExpanded && (
                <div className="mt-1 ml-4 pl-3 border-l border-border-subtle space-y-0.5">
                  {section.pages.map((page) => {
                    const isActive = currentPage?.slug === page.slug;

                    return (
                      <button
                        key={page.slug}
                        onClick={() => onPageSelect(page.slug)}
                        className={`
                          w-full text-left px-3 py-1.5 rounded-md text-sm
                          transition-colors duration-200
                          ${isActive
                            ? "text-brand-500 bg-brand-500/10 font-medium"
                            : "text-text-muted hover:text-text-secondary hover:bg-bg-secondary/50"
                          }
                        `}
                      >
                        {page.title}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </nav>

      {/* Divider */}
      <div className="my-4 border-t border-border-subtle" />

      {/* Quick Links */}
      <div className="px-3">
        <h3 className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-3">
          Quick Links
        </h3>
        <div className="space-y-1">
          <a
            href="/dashboard"
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
          >
            Dashboard
          </a>
          <a
            href="/registry"
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
          >
            Browse Functions
          </a>
          <a
            href="/faq"
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
          >
            FAQ
          </a>
          <a
            href="/contact"
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-text-secondary hover:text-text-primary hover:bg-bg-secondary transition-colors"
          >
            Support
          </a>
        </div>
      </div>
    </div>
  );
}
