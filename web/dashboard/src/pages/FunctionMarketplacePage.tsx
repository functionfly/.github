import { FunctionMarketplace } from "@/components/swarm/FunctionMarketplace";
import { Footer } from "@/pages/LandingPage/components/Footer";

export function FunctionMarketplacePage() {
  return (
    <div className="min-h-screen flex flex-col">
      <div className="flex-1">
        <FunctionMarketplace />
      </div>
      <Footer />
    </div>
  );
}

export default FunctionMarketplacePage;
