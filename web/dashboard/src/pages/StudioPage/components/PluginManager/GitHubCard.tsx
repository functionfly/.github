import { useState } from "react";
import { Badge, GlassCard, Button, Spinner } from "@functionfly/ui-core";
import { Github, Link2, RefreshCw, Settings, ExternalLink, CheckCircle, XCircle } from "lucide-react";
import type { Plugin } from "@/api/plugins";

interface GitHubCardProps {
  plugin: Plugin;
  onConfigure?: (pluginId: string) => void;
  onEnable?: (pluginId: string) => void;
  onDisable?: (pluginId: string) => void;
  onSync?: (pluginId: string) => void;
}

interface Repo {
  id: number;
  name: string;
  full_name: string;
  private: boolean;
  default_branch: string;
}

export function GitHubCard({ plugin, onConfigure, onEnable, onDisable, onSync }: GitHubCardProps) {
  const [syncing, setSyncing] = useState(false);
  const [repos, setRepos] = useState<Repo[]>([]);
  const [showSettings, setShowSettings] = useState(false);

  const config = plugin.config || {};
  const isEnabled = plugin.status === "enabled";
  const hasToken = !!config.github_token;

  const handleSync = async () => {
    setSyncing(true);
    try {
      // Simulated sync - in real implementation would call API
      setRepos([
        { id: 1, name: "example-repo", full_name: "functionfly/example-repo", private: false, default_branch: "main" },
        { id: 2, name: "workflows", full_name: "functionfly/workflows", private: true, default_branch: "develop" },
      ]);
    } finally {
      setSyncing(false);
    }
  };

  const handleConnect = () => {
    // Would open OAuth flow to GitHub
    onConfigure?.(plugin.id);
  };

  return (
    <GlassCard className="p-4 space-y-4">
      <div className="flex items-start gap-3">
        <div className="w-10 h-10 rounded-lg bg-[#24292e] flex items-center justify-center">
          <Github className="w-6 h-6 text-white" />
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-white">{plugin.name}</h3>
            {plugin.verified && (
              <Badge variant="success" size="sm">
                <CheckCircle className="w-3 h-3" />
              </Badge>
            )}
          </div>
          <p className="text-xs text-white/60 mt-0.5">v{plugin.version} • by {plugin.author_name}</p>
        </div>
        <Badge variant={isEnabled ? "success" : "secondary"} size="sm">
          {isEnabled ? "Active" : "Disabled"}
        </Badge>
      </div>

      {!hasToken ? (
        <div className="p-3 rounded-lg bg-white/5 border border-white/10 text-center">
          <p className="text-xs text-white/60 mb-3">Connect your GitHub account to enable integration</p>
          <Button size="sm" variant="default" onClick={handleConnect}>
            <Github className="w-4 h-4 mr-1" />
            Connect GitHub
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs">
            <span className="text-white/60">Default repo:</span>
            <span className="text-white">{config.default_repo || "Not set"}</span>
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-white/60">Auto-sync:</span>
            <span className="text-white">{config.auto_sync ? "Enabled" : "Disabled"}</span>
          </div>

          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="ghost"
              onClick={handleSync}
              disabled={syncing}
              className="flex-1"
            >
              {syncing ? (
                <Spinner className="w-4 h-4" />
              ) : (
                <RefreshCw className="w-4 h-4" />
              )}
              Sync Repos
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setShowSettings(!showSettings)}
            >
              <Settings className="w-4 h-4" />
            </Button>
          </div>

          {repos.length > 0 && (
            <div className="space-y-1 max-h-32 overflow-y-auto">
              {repos.map((repo) => (
                <div
                  key={repo.id}
                  className="flex items-center justify-between p-2 rounded bg-white/5 text-xs"
                >
                  <div className="flex items-center gap-2">
                    <Link2 className="w-3 h-3 text-white/40" />
                    <span className="text-white">{repo.full_name}</span>
                    {repo.private && (
                      <Badge variant="secondary" size="sm">private</Badge>
                    )}
                  </div>
                  <a
                    href={`https://github.com/${repo.full_name}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-white/40 hover:text-white"
                  >
                    <ExternalLink className="w-3 h-3" />
                  </a>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="flex items-center gap-2 pt-2 border-t border-white/10">
        <Button
          size="sm"
          variant="ghost"
          onClick={() => isEnabled ? onDisable?.(plugin.id) : onEnable?.(plugin.id)}
          className="flex-1"
        >
          {isEnabled ? (
            <>
              <XCircle className="w-4 h-4 mr-1" />
              Disable
            </>
          ) : (
            <>
              <CheckCircle className="w-4 h-4 mr-1" />
              Enable
            </>
          )}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => onConfigure?.(plugin.id)}
        >
          <Settings className="w-4 h-4" />
        </Button>
      </div>
    </GlassCard>
  );
}