import { ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";

const Navigation = () => {
  return (
    <nav className="border-b border-border-subtle bg-bg-primary/80 backdrop-blur-md sticky top-0 z-50 relative overflow-hidden">
      {/* Background gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-brand-500/5 via-transparent to-purple-500/5" />

      <div className="relative max-w-7xl mx-auto px-4 lg:px-6">
        <div className="flex items-center justify-between h-16">
          <Link
            to="/"
            className="flex items-center gap-2 text-text-primary hover:text-brand-500 transition-all duration-300 group"
          >
            <div className="p-1 rounded-lg bg-bg-hover group-hover:bg-brand-500/10 transition-colors">
              <ArrowLeft className="w-4 h-4" />
            </div>
            <span className="font-medium">Back to Home</span>
          </Link>

          <div className="flex items-center gap-3">
            <div className="w-2 h-2 rounded-full bg-gradient-to-r from-brand-500 to-purple-500 animate-pulse" />
            <h1 className="text-xl font-bold bg-gradient-to-r from-text-primary to-text-secondary bg-clip-text text-transparent">
              Integrations
            </h1>
          </div>

          <div className="w-24" /> {/* Spacer for centering */}
        </div>
      </div>
    </nav>
  );
};

export default Navigation;