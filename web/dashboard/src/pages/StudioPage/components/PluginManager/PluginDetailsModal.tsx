import { useState } from "react";
import { GlassCard, Badge, Button, Spinner } from "@functionfly/ui-core";
import { X, Star, Download, Shield, Clock, GitBranch, ChevronDown, ChevronUp, Check } from "lucide-react";
import { type Extension } from "@/api/marketplace";
import { useInstallPlugin, type InstallPluginRequest, type PluginType } from "@/hooks/usePlugin";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";

interface PluginDetailsModalProps {
  extension: Extension;
  onClose: () => void;
  onInstall: (plugin: InstallPluginRequest) => void;
}

export function PluginDetailsModal({ extension, onClose, onInstall }: PluginDetailsModalProps) {
  const [selectedVersion, setSelectedVersion] = useState(extension.version);
  const [installing, setInstalling] = useState(false);
  const { data: fullExtension } = useQuery({
    queryKey: ["extension-details", extension.id],
    queryFn: () => marketplaceApi.get(extension.id),
    enabled: !!extension.id,
  });

  const handleInstall = async () => {
    setInstalling(true);
    try {
      await onInstall({
        manifest: extension.manifest || { name: extension.name, version: selectedVersion },
        plugin_type: (extension.category as PluginType) || "ui",
        name: extension.name,
        version: selectedVersion,
        description: extension.description,
        author_name: extension.creator_id,
        category: extension.category,
        size_bytes: 0,
      });
      onClose();
    } catch (error) {
      console.error("Failed to install:", error);
    } finally {
      setInstalling(false);
    }
  };

  const formatNumber = (n: number) => {
    if (n >= 1000) return (n / 1000).toFixed(1) + "k";
    return n.toString();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div className="bg-[#1a1a2e] border border-white/10 rounded-xl w-full max-w-2xl max-h-[85vh] overflow-hidden shadow-2xl">
        <div className="flex items-center justify-between p-5 border-b border-white/10">
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-2xl font-bold text-white/80">
              {extension.name.charAt(0)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-xl font-semibold text-white">{extension.name}</h2>
                {extension.verified && <Shield className="w-5 h-5 text-green-400" />}
              </div>
              <p className="text-sm text-white/60">by {extension.creator_id}</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 rounded-lg hover:bg-white/10 text-white/60 hover:text-white">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-5 overflow-y-auto max-h-[60vh]">
          <div className="flex items-center gap-4 mb-6">
            <Badge className="bg-white/10 text-white/80 border-white/20">
              <Download className="w-3 h-3 mr-1" />
              {formatNumber(extension.install_count)} installs
            </Badge>
            <Badge className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
              <Star className="w-3 h-3 mr-1 fill-current" />
              {extension.rating_average.toFixed(1)} ({extension.rating_count} reviews)
            </Badge>
            <Badge className="bg-white/10 text-white/80 border-white/20">
              v{extension.version}
            </Badge>
          </div>

          <p className="text-white/80 leading-relaxed mb-6">{extension.description}</p>

          {extension.screenshots && extension.screenshots.length > 0 && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-white/60 mb-3">Screenshots</h3>
              <div className="flex gap-3 overflow-x-auto pb-2">
                {extension.screenshots.map((screenshot, i) => (
                  <img
                    key={i}
                    src={screenshot}
                    alt={`Screenshot ${i + 1}`}
                    className="h-32 w-auto rounded-lg bg-white/5"
                  />
                ))}
              </div>
            </div>
          )}

          <div className="mb-6">
            <h3 className="text-sm font-medium text-white/60 mb-3">Version</h3>
            <select
              value={selectedVersion}
              onChange={(e) => setSelectedVersion(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2.5 text-white"
            >
              <option value={extension.version}>{extension.version} (latest)</option>
            </select>
          </div>

          {extension.changelog && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-white/60 mb-3">Changelog</h3>
              <div className="p-4 bg-white/5 rounded-lg border border-white/10">
                <pre className="text-sm text-white/70 whitespace-pre-wrap font-mono">{extension.changelog}</pre>
              </div>
            </div>
          )}

          {fullExtension?.extension?.compatibility && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-white/60 mb-3">Compatibility</h3>
              <div className="flex flex-wrap gap-2">
                {Object.entries(fullExtension.extension.compatibility).map(([key, value]) => (
                  <Badge key={key} className="bg-white/10 text-white/80 border-white/20">
                    {key}: {String(value)}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {extension.tags && extension.tags.length > 0 && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-white/60 mb-3">Tags</h3>
              <div className="flex flex-wrap gap-2">
                {extension.tags.map((tag) => (
                  <Badge key={tag} variant="outline" className="text-white/60 border-white/20">
                    {tag}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          <div className="grid grid-cols-4 gap-4 p-4 bg-white/5 rounded-lg border border-white/10">
            <div className="text-center">
              <div className="text-lg font-bold text-white">{extension.trust_score || 0}</div>
              <div className="text-xs text-white/60">Trust</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-white">{extension.security_score || 0}</div>
              <div className="text-xs text-white/60">Security</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-white">{extension.sandbox_score || 0}</div>
              <div className="text-xs text-white/60">Sandbox</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-bold text-white">{extension.runtime_score || 0}</div>
              <div className="text-xs text-white/60">Runtime</div>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 p-5 border-t border-white/10 bg-white/5">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button
            onClick={handleInstall}
            disabled={installing}
            className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-400 hover:to-red-400"
          >
            {installing ? <Spinner className="w-4 h-4" /> : <><Download className="w-4 h-4 mr-2" /> Install v{selectedVersion}</>}
          </Button>
        </div>
      </div>
    </div>
  );
}