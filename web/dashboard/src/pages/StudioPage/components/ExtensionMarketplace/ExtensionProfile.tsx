import { Badge, GlassCard, Button } from "@functionfly/ui-core";
import { Star, Download, Shield, Clock, ExternalLink } from "lucide-react";
import type { Extension } from "@/api/marketplace";

interface ExtensionProfileProps {
  extension: Extension;
  onInstall?: (id: string) => void;
  onFavorite?: (id: string) => void;
}

export function ExtensionProfile({ extension, onInstall, onFavorite }: ExtensionProfileProps) {
  const ratingStars = Math.round(extension.rating_average || 0);
  const trustLevel = extension.trust_score >= 80 ? "High" : extension.trust_score >= 50 ? "Medium" : "Low";
  const trustColor = extension.trust_score >= 80 ? "text-green-400" : extension.trust_score >= 50 ? "text-yellow-400" : "text-red-400";

  return (
    <div className="space-y-4">
      <div className="flex items-start gap-4">
        <div className="w-16 h-16 rounded-lg bg-bg-secondary flex items-center justify-center overflow-hidden">
          {extension.icon_url ? (
            <img src={extension.icon_url} alt={extension.name} className="w-full h-full object-cover" />
          ) : (
            <span className="text-2xl font-bold text-white/40">{extension.name[0]}</span>
          )}
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h3 className="text-lg font-semibold text-white">{extension.name}</h3>
            {extension.verified && (
              <Badge variant="success" size="sm">
                <Shield className="w-3 h-3 mr-1" />
                Verified
              </Badge>
            )}
            {extension.featured && (
              <Badge variant="outline" size="sm" className="text-yellow-400 border-yellow-400/50">
                Featured
              </Badge>
            )}
          </div>
          <p className="text-sm text-white/60 mt-1">{extension.description}</p>
          <div className="flex items-center gap-4 mt-2 text-xs text-white/40">
            <span className="flex items-center gap-1">
              <Star className="w-3 h-3 text-yellow-400" />
              {extension.rating_average.toFixed(1)} ({extension.rating_count})
            </span>
            <span className="flex items-center gap-1">
              <Download className="w-3 h-3" />
              {extension.install_count.toLocaleString()} installs
            </span>
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              v{extension.version}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <GlassCard className="p-3">
          <div className="text-[10px] text-white/40 uppercase tracking-wider mb-1">Trust Score</div>
          <div className="flex items-center gap-2">
            <span className={`text-lg font-semibold ${trustColor}`}>{trustLevel}</span>
            <div className="flex-1 h-2 bg-white/10 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full ${extension.trust_score >= 80 ? "bg-green-400" : extension.trust_score >= 50 ? "bg-yellow-400" : "bg-red-400"}`}
                style={{ width: `${extension.trust_score}%` }}
              />
            </div>
            <span className="text-sm text-white/60">{extension.trust_score.toFixed(0)}%</span>
          </div>
        </GlassCard>

        <GlassCard className="p-3">
          <div className="text-[10px] text-white/40 uppercase tracking-wider mb-1">Security</div>
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-blue-400" />
            <span className="text-lg font-semibold text-white">{extension.security_score.toFixed(0)}%</span>
            <div className="flex-1 h-2 bg-white/10 rounded-full overflow-hidden">
              <div
                className="h-full rounded-full bg-blue-400"
                style={{ width: `${extension.security_score}%` }}
              />
            </div>
          </div>
        </GlassCard>

        <GlassCard className="p-3">
          <div className="text-[10px] text-white/40 uppercase tracking-wider mb-1">Sandbox</div>
          <div className="flex items-center gap-2">
            <span className="text-lg font-semibold text-white">{extension.sandbox_score.toFixed(0)}%</span>
            <div className="flex-1 h-2 bg-white/10 rounded-full overflow-hidden">
              <div
                className="h-full rounded-full bg-purple-400"
                style={{ width: `${extension.sandbox_score}%` }}
              />
            </div>
          </div>
        </GlassCard>

        <GlassCard className="p-3">
          <div className="text-[10px] text-white/40 uppercase tracking-wider mb-1">Runtime</div>
          <div className="flex items-center gap-2">
            <span className="text-lg font-semibold text-white">{extension.runtime_score.toFixed(0)}%</span>
            <div className="flex-1 h-2 bg-white/10 rounded-full overflow-hidden">
              <div
                className="h-full rounded-full bg-orange-400"
                style={{ width: `${extension.runtime_score}%` }}
              />
            </div>
          </div>
        </GlassCard>
      </div>

      {extension.tags && extension.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {extension.tags.map((tag) => (
            <Badge key={tag} variant="outline" size="sm" className="text-white/60 border-white/20">
              {tag}
            </Badge>
          ))}
        </div>
      )}

      <div className="flex items-center gap-2">
        <Button size="sm" variant="default" onClick={() => onInstall?.(extension.id)}>
          Install
        </Button>
        <Button size="sm" variant="ghost" onClick={() => onFavorite?.(extension.id)}>
          <Star className="w-4 h-4" />
        </Button>
        {extension.homepage_url && (
          <Button size="sm" variant="ghost" asChild>
            <a href={extension.homepage_url} target="_blank" rel="noopener noreferrer">
              <ExternalLink className="w-4 h-4" />
            </a>
          </Button>
        )}
      </div>
    </div>
  );
}