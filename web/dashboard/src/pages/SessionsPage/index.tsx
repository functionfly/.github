import { useState } from "react";
import { Link } from "react-router-dom";
import { Zap, Monitor, Smartphone, Globe, Trash2, RefreshCw, ShieldCheck, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { useAuthStore } from "@/stores/authStore";
import { useUserSessions, useRevokeSession, useRevokeOtherSessions } from "@/hooks";
import type { SessionItem } from "@/api/users";

export function SessionsPage() {
  const { user, logout } = useAuthStore();
  const [isRevoking, setIsRevoking] = useState<string | null>(null);

  // Use hooks instead of raw queries/mutations
  const { data, isLoading, error } = useUserSessions();
  const revokeMutation = useRevokeSession();
  const revokeOthersMutation = useRevokeOtherSessions();

  const sessions: SessionItem[] = data?.sessions ?? [];

  const handleRevokeSession = (sessionId: string) => {
    setIsRevoking(sessionId);
    revokeMutation.mutate(sessionId, {
      onSettled: () => setIsRevoking(null),
    });
  };

  const handleRevokeAllOthers = () => {
    if (!confirm("Are you sure you want to sign out all other devices? This action cannot be undone.")) {
      return;
    }
    setIsRevoking("all");
    revokeOthersMutation.mutate(undefined, {
      onSettled: () => setIsRevoking(null),
    });
  };

  const getDeviceIcon = (device: string) => {
    if (device.toLowerCase().includes("iphone") || device.toLowerCase().includes("android")) {
      return <Smartphone className="w-5 h-5" />;
    }
    if (device.toLowerCase().includes("mac") || device.toLowerCase().includes("windows") || device.toLowerCase().includes("linux")) {
      return <Monitor className="w-5 h-5" />;
    }
    return <Globe className="w-5 h-5" />;
  };

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Header */}
      <header className="border-b border-border-subtle bg-bg-secondary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link to="/dashboard" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" fill="currentColor" />
              </div>
              <span className="text-xl font-bold gradient-text">FunctionFly</span>
            </Link>
            <div className="flex items-center gap-4">
              <Link to="/settings">
                <Button variant="ghost" size="sm">Settings</Button>
              </Link>
              <Button variant="outline" size="sm" onClick={() => logout()}>
                Sign Out
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-text-primary">Active Sessions</h1>
          <p className="text-text-secondary mt-1">
            Manage your active sessions and sign out of devices you don't recognize.
          </p>
        </div>

        {/* Security Notice */}
        <Card className="mb-6 border-blue-200 bg-blue-50">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <ShieldCheck className="w-5 h-5 text-blue-600 mt-0.5" />
              <div>
                <p className="font-medium text-blue-900">Your account is secure</p>
                <p className="text-sm text-blue-700 mt-1">
                  Review your active sessions below. If you see any unfamiliar devices, 
                  consider changing your password or revoking access.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Sessions List */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>Active Sessions</CardTitle>
              <CardDescription>
                {sessions.length} device{sessions.length !== 1 ? 's' : ''} currently signed in
              </CardDescription>
            </div>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={handleRevokeAllOthers}
              disabled={isRevoking !== null || sessions.length <= 1}
            >
              <Trash2 className="w-4 h-4 mr-2" />
              Sign Out All Other Devices
            </Button>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex justify-center py-8">
                <LoadingSpinner text="Loading sessions..." />
              </div>
            ) : error ? (
              <div className="text-center py-8 text-red-600 dark:text-red-400 text-sm">
                Failed to load sessions. You may need to sign in again.
              </div>
            ) : (
              <div className="space-y-4">
                {sessions.map((session) => (
                  <div
                    key={session.id}
                    className={`flex items-center justify-between p-4 rounded-lg border ${
                      session.currentSession 
                        ? "border-green-200 bg-green-50" 
                        : "border-border-subtle bg-bg-secondary"
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <div className={`p-2 rounded-full ${
                        session.currentSession 
                          ? "bg-green-100 text-green-600" 
                          : "bg-bg-tertiary text-text-muted"
                      }`}>
                        {getDeviceIcon(session.device)}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-medium text-text-primary">{session.device}</p>
                          {session.currentSession && (
                            <span className="px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                              Current
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-3 text-sm text-text-muted mt-1">
                          <span className="flex items-center gap-1">
                            <Globe className="w-3 h-3" />
                            {session.ip}
                          </span>
                          <span>•</span>
                          <span>{session.location}</span>
                          <span>•</span>
                          <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {session.lastActive}
                          </span>
                        </div>
                      </div>
                    </div>
                    
                    {!session.currentSession && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRevokeSession(session.id)}
                        disabled={isRevoking === session.id}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        {isRevoking === session.id ? (
                          <RefreshCw className="w-4 h-4 animate-spin" />
                        ) : (
                          <>
                            <Trash2 className="w-4 h-4 mr-1" />
                            Sign Out
                          </>
                        )}
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Help Text */}
        <p className="text-center text-sm text-text-muted mt-6">
          Having trouble?{" "}
          <a href="mailto:support@functionfly.com" className="text-brand-500 hover:underline">
            Contact Support
          </a>
        </p>
      </main>
    </div>
  );
}
