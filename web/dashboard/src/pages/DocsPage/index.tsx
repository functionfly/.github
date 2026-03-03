import { useState, useEffect, useRef } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Menu, X, Search, ChevronRight, Book, Code, Terminal, Zap, Shield, Settings, Layout, Cpu, Globe, Database } from "lucide-react";
import { MetaTags } from "@/components/seo/MetaTags";
import { DocsSidebar } from "./components/DocsSidebar";
import { DocsTableOfContents } from "./components/DocsTableOfContents";
import { DocsContent } from "./components/DocsContent";
import { DocsSearch } from "./components/DocsSearch";
import { docSections, type DocSection, type DocPage } from "./data/docs";
import "./docs.css";

export function DocsPage() {
  const { slug } = useParams<{ slug?: string }>();
  const navigate = useNavigate();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState<DocPage | null>(null);
  const [currentSection, setCurrentSection] = useState<DocSection | null>(null);
  const mainContentRef = useRef<HTMLDivElement>(null);

  // Find current page based on slug
  useEffect(() => {
    const pageSlug = slug || "welcome";

    for (const section of docSections) {
      const page = section.pages.find(p => p.slug === pageSlug);
      if (page) {
        setCurrentPage(page);
        setCurrentSection(section);
        break;
      }
    }
  }, [slug]);

  // Close mobile menu on route change
  useEffect(() => {
    setIsMobileMenuOpen(false);
  }, [slug]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd/Ctrl + K to open search
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setIsSearchOpen(true);
      }
      // Escape to close search or mobile menu
      if (e.key === "Escape") {
        setIsSearchOpen(false);
        setIsMobileMenuOpen(false);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handlePageChange = (pageSlug: string) => {
    navigate(`/docs/${pageSlug}`);
    mainContentRef.current?.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="docs-page min-h-screen bg-bg-primary">
      <MetaTags
        title={currentPage ? `${currentPage.title} | FunctionFly Docs` : "Documentation | FunctionFly"}
        description={currentPage?.description || "FunctionFly documentation - Learn how to deploy, manage, and scale serverless functions across multiple edge providers."}
        url={`/docs${slug ? `/${slug}` : ""}`}
      />

      {/* Top Navigation */}
      <header className="docs-header sticky top-0 z-50 border-b border-border-subtle bg-bg-primary/80 backdrop-blur-xl">
        <div className="max-w-[1600px] mx-auto px-4 lg:px-6">
          <div className="flex items-center justify-between h-16">
            {/* Left: Logo & Mobile Menu */}
            <div className="flex items-center gap-4">
              <button
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                className="lg:hidden p-2 -ml-2 text-text-secondary hover:text-text-primary transition-colors"
                aria-label="Toggle menu"
              >
                {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
              </button>

              <a href="/" className="flex items-center gap-2.5 group">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-brand-600 flex items-center justify-center shadow-glow-sm group-hover:shadow-glow transition-shadow">
                  <Zap className="w-5 h-5 text-white" />
                </div>
                <span className="font-bold text-lg text-text-primary hidden sm:block">FunctionFly</span>
              </a>
            </div>

            {/* Center: Search */}
            <button
              onClick={() => setIsSearchOpen(true)}
              className="hidden md:flex items-center gap-2 px-4 py-2 rounded-lg bg-bg-secondary border border-border-subtle text-text-muted hover:border-border-default hover:text-text-secondary transition-all w-64 lg:w-80"
            >
              <Search className="w-4 h-4" />
              <span className="text-sm flex-1 text-left">Search docs...</span>
              <kbd className="hidden lg:inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs font-mono bg-bg-tertiary text-text-muted border border-border-subtle">
                <span className="text-[10px]">⌘</span>K
              </kbd>
            </button>

            {/* Right: Nav Links */}
            <nav className="flex items-center gap-1 sm:gap-4" aria-label="Docs site navigation">
              <a href="/registry" className="hidden sm:flex items-center gap-1.5 px-3 py-1.5 text-sm text-text-secondary hover:text-text-primary transition-colors">
                <Code className="w-4 h-4" />
                <span>Registry</span>
              </a>
              <a href="/features" className="hidden sm:flex items-center gap-1.5 px-3 py-1.5 text-sm text-text-secondary hover:text-text-primary transition-colors">
                <Layout className="w-4 h-4" />
                <span>Features</span>
              </a>
              <a href="/dashboard" className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-white bg-brand-500 hover:bg-brand-600 rounded-lg transition-colors">
                <Terminal className="w-4 h-4" />
                <span className="hidden sm:inline">Dashboard</span>
              </a>
            </nav>
          </div>
        </div>
      </header>

      {/* Mobile Menu Overlay */}
      {isMobileMenuOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* Search Modal */}
      {isSearchOpen && (
        <DocsSearch
          onClose={() => setIsSearchOpen(false)}
          onSelect={handlePageChange}
        />
      )}

      {/* Main Layout */}
      <div className="max-w-[1600px] mx-auto">
        <div className="flex">
          {/* Left Sidebar */}
          <aside className={`
            fixed lg:sticky top-16 left-0 z-40
            w-72 h-[calc(100vh-4rem)]
            bg-bg-primary lg:bg-transparent
            border-r border-border-subtle
            transform transition-transform duration-300 ease-out
            ${isMobileMenuOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"}
          `}>
            <DocsSidebar
              sections={docSections}
              currentPage={currentPage}
              currentSection={currentSection}
              onPageSelect={handlePageChange}
            />
          </aside>

          {/* Main Content Area */}
          <main
            ref={mainContentRef}
            className="flex-1 min-w-0 h-[calc(100vh-4rem)] overflow-y-auto"
          >
            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 lg:py-12">
              {/* Breadcrumbs */}
              {currentSection && (
                <nav className="flex items-center gap-2 text-sm text-text-muted mb-6">
                  <Book className="w-4 h-4" />
                  <ChevronRight className="w-3 h-3" />
                  <span className="text-text-secondary">{currentSection.title}</span>
                  <ChevronRight className="w-3 h-3" />
                  <span className="text-text-primary">{currentPage?.title}</span>
                </nav>
              )}

              {/* Page Content */}
              {currentPage ? (
                <DocsContent page={currentPage} />
              ) : (
                <div className="text-center py-20">
                  <h1 className="text-2xl font-bold text-text-primary mb-2">Page Not Found</h1>
                  <p className="text-text-secondary">The documentation page you're looking for doesn't exist.</p>
                  <button
                    onClick={() => handlePageChange("welcome")}
                    className="mt-4 px-4 py-2 bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors"
                  >
                    Go to Documentation
                  </button>
                </div>
              )}
            </div>
          </main>

          {/* Right Sidebar - Table of Contents */}
          <aside className="hidden xl:block w-64 sticky top-16 h-[calc(100vh-4rem)] overflow-y-auto border-l border-border-subtle">
            <DocsTableOfContents content={currentPage?.content || ""} />
          </aside>
        </div>
      </div>
    </div>
  );
}

export default DocsPage;
