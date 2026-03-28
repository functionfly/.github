import { Navbar } from '@/components/common/Navbar';
import { BrowseFunctionsView } from '@/components/registry/BrowseFunctionsView';
import { MetaTags } from '@/components/seo/MetaTags';
import { Footer } from '@/pages/LandingPage/components';
export function BrowseFunctionsPage() {
  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      <MetaTags
        title="Browse Functions | Registry"
        description="Discover and explore premium serverless functions. Browse the registry, deploy instantly, or try live in the playground."
        keywords={['function registry', 'serverless', 'browse functions', 'deploy functions']}
      />
      <Navbar variant="landing" />
      <BrowseFunctionsView variant="public" />

      <Footer />
    </div>
  );
}
