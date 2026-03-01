import { useRouteError, isRouteErrorResponse, Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { AlertTriangle, Home, ArrowLeft, RefreshCw } from "lucide-react";

export function ServerErrorPage() {
  const error = useRouteError();
  const errorMessage = isRouteErrorResponse(error) 
    ? error.statusText 
    : error instanceof Error 
      ? error.message 
      : "An unexpected error occurred";

  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-primary p-4">
      <div className="max-w-md w-full text-center">
        <div className="mb-8">
          <div className="w-24 h-24 mx-auto mb-6 rounded-full bg-error/10 flex items-center justify-center">
            <AlertTriangle className="w-12 h-12 text-error" />
          </div>
          <h1 className="text-6xl font-bold text-text-primary mb-2">500</h1>
          <h2 className="text-xl font-semibold text-text-primary mb-2">Internal Server Error</h2>
          <p className="text-text-secondary mb-2">
            Something went wrong on our end. We're working to fix the issue.
          </p>
          <p className="text-text-muted text-sm">
            Error: {errorMessage}
          </p>
        </div>
        
        <div className="space-y-4">
          <Button asChild className="w-full">
            <Link to="/dashboard">
              <Home className="w-4 h-4 mr-2" />
              Go to Dashboard
            </Link>
          </Button>
          <div className="flex gap-2">
            <Button 
              variant="outline" 
              className="flex-1"
              onClick={() => window.history.back()}
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Go Back
            </Button>
            <Button 
              variant="outline" 
              className="flex-1"
              onClick={handleRefresh}
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ServerErrorPage;
