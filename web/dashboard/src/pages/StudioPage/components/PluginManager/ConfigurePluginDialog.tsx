import { useState, useMemo } from "react";
import { Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@functionfly/ui-core";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Code, MessageCircle, CreditCard, Cloud, Database, HardDrive, Brain, Webhook, Clock, Gauge, Mail, Link2, RefreshCw, ExternalLink, Loader2, Check, X, Key, Settings, Bell, Globe } from "lucide-react";
import { type Plugin } from "@/api/plugins";
import { useGitHubConnection, useGitHubConnect, useGitHubDisconnect } from "@/hooks/useGitHubConnection";
import { useQuery } from "@tanstack/react-query";
import { githubApi } from "@/api/github";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface ConfigurePluginDialogProps {
  plugin: Plugin;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSaveConfig: (pluginId: string, config: Record<string, string>) => void;
}

interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  private: boolean;
  default_branch: string;
}

const PLUGIN_ICONS: Record<string, React.ReactNode> = {
  "github": <Code className="w-5 h-5" />,
  "slack": <MessageCircle className="w-5 h-5" />,
  "stripe": <CreditCard className="w-5 h-5" />,
  "vercel": <Cloud className="w-5 h-5" />,
  "datadog": <Database className="w-5 h-5" />,
  "sendgrid": <Mail className="w-5 h-5" />,
  "s3": <HardDrive className="w-5 h-5" />,
  "postgres": <Database className="w-5 h-5" />,
  "ai-context": <Brain className="w-5 h-5" />,
  "webhook": <Webhook className="w-5 h-5" />,
  "scheduler": <Clock className="w-5 h-5" />,
  "rate-limiter": <Gauge className="w-5 h-5" />,
};

const getPluginIcon = (name: string): React.ReactNode => {
  const lower = name.toLowerCase();
  for (const [key, icon] of Object.entries(PLUGIN_ICONS)) {
    if (lower.includes(key)) return icon;
  }
  return <Settings className="w-5 h-5" />;
};

export function ConfigurePluginDialog({ plugin, open, onOpenChange, onSaveConfig }: ConfigurePluginDialogProps) {
  const [config, setConfig] = useState<Record<string, string>>(plugin.config || {});
  const [isSaving, setIsSaving] = useState(false);

  const { data: githubConnection, isLoading: connectionLoading } = useGitHubConnection();
  const connectMutation = useGitHubConnect();
  const disconnectMutation = useGitHubDisconnect();

  const { data: reposData, isLoading: reposLoading, refetch: refetchRepos } = useQuery({
    queryKey: ['github-repos-config'],
    queryFn: () => githubApi.listRepos({ per_page: 50 }),
    enabled: !!githubConnection,
  });

  const pluginType = useMemo(() => {
    const lower = plugin.name.toLowerCase();
    if (lower.includes("github")) return "github";
    if (lower.includes("slack")) return "slack";
    if (lower.includes("stripe")) return "stripe";
    if (lower.includes("vercel")) return "vercel";
    if (lower.includes("datadog")) return "datadog";
    if (lower.includes("sendgrid")) return "sendgrid";
    if (lower.includes("s3") || lower.includes("aws")) return "aws-s3";
    if (lower.includes("postgres") || lower.includes("sql")) return "postgres";
    if (lower.includes("ai-context") || lower.includes("context")) return "ai-context";
    if (lower.includes("webhook")) return "webhook";
    if (lower.includes("scheduler")) return "scheduler";
    if (lower.includes("rate-limiter") || lower.includes("rate limit")) return "rate-limiter";
    return "generic";
  }, [plugin.name]);

  const handleSave = async () => {
    setIsSaving(true);
    try {
      await onSaveConfig(plugin.id, config);
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to save config:", error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleConnectGitHub = () => {
    connectMutation.mutate();
  };

  const handleDisconnectGitHub = () => {
    disconnectMutation.mutate();
  };

  const updateConfig = (key: string, value: string) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-[#1a1a2e] border-white/10 max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-white flex items-center gap-2">
            <span className="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center">
              {getPluginIcon(plugin.name)}
            </span>
            Configure {plugin.name}
          </DialogTitle>
          <DialogDescription className="text-white/60">
            v{plugin.version} • Configure plugin settings and integrations
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* GitHub Integration */}
          {pluginType === "github" && (
            <GitHubConfigSection
              config={config}
              updateConfig={updateConfig}
              githubConnection={githubConnection}
              connectionLoading={connectionLoading}
              reposData={reposData}
              reposLoading={reposLoading}
              connectMutation={connectMutation}
              disconnectMutation={disconnectMutation}
              onConnectGitHub={handleConnectGitHub}
              onDisconnectGitHub={handleDisconnectGitHub}
              onRefreshRepos={() => refetchRepos()}
            />
          )}

          {/* Slack Alerts */}
          {pluginType === "slack" && (
            <SlackConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Stripe Billing */}
          {pluginType === "stripe" && (
            <StripeConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Vercel Deployments */}
          {pluginType === "vercel" && (
            <VercelConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Datadog Monitoring */}
          {pluginType === "datadog" && (
            <DatadogConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* SendGrid Email */}
          {pluginType === "sendgrid" && (
            <SendGridConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* AWS S3 Storage */}
          {(pluginType === "aws-s3") && (
            <AwsS3ConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* PostgreSQL Connector */}
          {pluginType === "postgres" && (
            <PostgresConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* AI Context Manager */}
          {pluginType === "ai-context" && (
            <AIContextConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Webhook Builder */}
          {pluginType === "webhook" && (
            <WebhookConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Workflow Scheduler */}
          {pluginType === "scheduler" && (
            <SchedulerConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Rate Limiter */}
          {pluginType === "rate-limiter" && (
            <RateLimiterConfigSection config={config} updateConfig={updateConfig} />
          )}

          {/* Generic / Fallback */}
          {pluginType === "generic" && (
            <GenericConfigSection config={config} updateConfig={updateConfig} setConfig={setConfig} />
          )}
        </div>

        <DialogFooter className="border-t border-white/10 pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
            Save Configuration
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ============================================================================
// Configuration Sections for Each Plugin Type
// ============================================================================

interface ConfigSectionProps {
  config: Record<string, string>;
  updateConfig: (key: string, value: string) => void;
  setConfig?: React.Dispatch<React.SetStateAction<Record<string, string>>>;
}

interface GitHubSectionProps extends ConfigSectionProps {
  githubConnection: any;
  connectionLoading: boolean;
  reposData: any;
  reposLoading: boolean;
  connectMutation: any;
  disconnectMutation: any;
  onConnectGitHub: () => void;
  onDisconnectGitHub: () => void;
  onRefreshRepos: () => void;
}

function GitHubConfigSection({
  config, updateConfig, githubConnection, connectionLoading,
  reposData, reposLoading, connectMutation, disconnectMutation,
  onConnectGitHub, onDisconnectGitHub, onRefreshRepos
}: GitHubSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Code className="w-4 h-4" />
          GitHub Connection
        </h3>
        {githubConnection && (
          <span className="text-xs text-green-400 flex items-center gap-1">
            <Check className="w-3 h-3" />
            Connected as @{githubConnection.github_username}
          </span>
        )}
      </div>

      {!githubConnection ? (
        <div className="p-4 rounded-lg bg-white/5 border border-white/10">
          <p className="text-sm text-white/60 mb-3">
            Connect your GitHub account to enable repository synchronization and workflow automation.
          </p>
          <Button
            variant="default"
            onClick={onConnectGitHub}
            disabled={connectMutation.isPending}
            className="w-full"
          >
            {connectMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Code className="w-4 h-4" />}
            Connect GitHub Account
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10">
            {githubConnection.github_avatar_url && (
              <img src={githubConnection.github_avatar_url} alt="" className="w-10 h-10 rounded-full" />
            )}
            <div className="flex-1">
              <p className="text-sm font-medium text-white">@{githubConnection.github_username}</p>
              <p className="text-xs text-white/40">{githubConnection.github_name || "GitHub User"}</p>
            </div>
            <Button size="sm" variant="ghost" onClick={onDisconnectGitHub} disabled={disconnectMutation.isPending} className="text-red-400 hover:text-red-300">
              {disconnectMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <X className="w-4 h-4" />}
            </Button>
          </div>

          <div className="space-y-2">
            <label className="text-xs text-white/60">Default Repository</label>
            <Select value={config.default_repo || ""} onValueChange={(v) => updateConfig("default_repo", v)}>
              <SelectTrigger className="bg-white/5 border-white/10 text-white">
                <SelectValue placeholder="Select repository" />
              </SelectTrigger>
              <SelectContent className="bg-[#1a1a2e] border-white/10">
                {reposLoading ? (
                  <div className="flex items-center justify-center p-4"><Loader2 className="w-4 h-4 animate-spin" /></div>
                ) : reposData?.repos?.length ? (
                  reposData.repos.map((repo: GitHubRepo) => (
                    <SelectItem key={repo.id} value={repo.full_name}>
                      <div className="flex items-center gap-2"><Link2 className="w-3 h-3" />{repo.full_name}</div>
                    </SelectItem>
                  ))
                ) : (
                  <div className="p-4 text-xs text-white/40 text-center">No repositories found</div>
                )}
              </SelectContent>
            </Select>
            <button type="button" onClick={onRefreshRepos} className="text-xs text-white/40 hover:text-white/60 flex items-center gap-1">
              <RefreshCw className="w-3 h-3" /> Refresh repositories
            </button>
          </div>

          <ToggleField
            label="Auto-sync"
            description="Automatically sync changes from GitHub"
            value={config.auto_sync === "true"}
            onChange={(v) => updateConfig("auto_sync", v ? "true" : "false")}
          />

          <div className="space-y-2">
            <label className="text-xs text-white/60">Sync Branch</label>
            <Input value={config.sync_branch || "main"} onChange={(e) => updateConfig("sync_branch", e.target.value)} placeholder="main" className="bg-white/5 border-white/10 text-white" />
            <p className="text-xs text-white/40">Branch to monitor for changes</p>
          </div>

          <div className="space-y-2">
            <label className="text-xs text-white/60">Webhook URL</label>
            <Input value={config.webhook_url || ""} onChange={(e) => updateConfig("webhook_url", e.target.value)} placeholder="https://api.example.com/webhooks/github" className="bg-white/5 border-white/10 text-white font-mono text-xs" />
            <p className="text-xs text-white/40">URL to receive GitHub webhook events</p>
          </div>
        </div>
      )}
    </div>
  );
}

function SlackConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <MessageCircle className="w-4 h-4" />
          Slack Workspace
        </h3>
        {config.access_token && (
          <span className="text-xs text-green-400 flex items-center gap-1">
            <Check className="w-3 h-3" /> Connected
          </span>
        )}
      </div>

      {!config.access_token ? (
        <div className="p-4 rounded-lg bg-white/5 border border-white/10">
          <p className="text-sm text-white/60 mb-3">
            Connect your Slack workspace to send notifications and alerts to channels.
          </p>
          <Button variant="default" className="w-full">
            <MessageCircle className="w-4 h-4" />
            Connect Slack Workspace
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          <div className="space-y-2">
            <label className="text-xs text-white/60">Default Channel</label>
            <Input value={config.default_channel || ""} onChange={(e) => updateConfig("default_channel", e.target.value)} placeholder="#general" className="bg-white/5 border-white/10 text-white" />
          </div>

          <ToggleField
            label="Include Workflow Context"
            description="Add workflow name and status to notifications"
            value={config.include_context !== "false"}
            onChange={(v) => updateConfig("include_context", v ? "true" : "false")}
          />

          <ToggleField
            label="Include Error Details"
            description="Include stack traces and error context"
            value={config.include_errors === "true"}
            onChange={(v) => updateConfig("include_errors", v ? "true" : "false")}
          />

          <div className="space-y-2">
            <label className="text-xs text-white/60">Rate Limit (messages/minute)</label>
            <Input type="number" value={config.rate_limit || "60"} onChange={(e) => updateConfig("rate_limit", e.target.value)} className="bg-white/5 border-white/10 text-white" />
            <p className="text-xs text-white/40">Slack has a limit of 1000 messages per minute</p>
          </div>
        </div>
      )}
    </div>
  );
}

function StripeConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <CreditCard className="w-4 h-4" />
          Stripe Configuration
        </h3>
        {config.secret_key && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Configured</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Secret Key</label>
          <Input type="password" value={config.secret_key || ""} onChange={(e) => updateConfig("secret_key", e.target.value)} placeholder="sk_live_..." className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Webhook Signing Secret</label>
          <Input value={config.webhook_secret || ""} onChange={(e) => updateConfig("webhook_secret", e.target.value)} placeholder="whsec_..." className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <ToggleField
          label="Test Mode"
          description="Use Stripe test environment"
          value={config.test_mode === "true"}
          onChange={(v) => updateConfig("test_mode", v ? "true" : "false")}
        />

        <ToggleField
          label="Sync Invoices"
          description="Automatically sync invoice events"
          value={config.sync_invoices === "true"}
          onChange={(v) => updateConfig("sync_invoices", v ? "true" : "false")}
        />

        <ToggleField
          label="Sync Subscriptions"
          description="Track subscription lifecycle events"
          value={config.sync_subscriptions === "true"}
          onChange={(v) => updateConfig("sync_subscriptions", v ? "true" : "false")}
        />
      </div>
    </div>
  );
}

function VercelConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Cloud className="w-4 h-4" />
          Vercel Project
        </h3>
        {config.token && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Connected</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Vercel Token</label>
          <Input type="password" value={config.token || ""} onChange={(e) => updateConfig("token", e.target.value)} placeholder="xxxxxxxxxxxxxxxxxxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Project ID</label>
          <Input value={config.project_id || ""} onChange={(e) => updateConfig("project_id", e.target.value)} placeholder="project_xxx" className="bg-white/5 border-white/10 text-white" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Environment</label>
          <Select value={config.environment || "production"} onValueChange={(v) => updateConfig("environment", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="production">Production</SelectItem>
              <SelectItem value="preview">Preview</SelectItem>
              <SelectItem value="development">Development</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <ToggleField
          label="Auto-deploy on Success"
          description="Trigger deployment when workflow completes"
          value={config.auto_deploy === "true"}
          onChange={(v) => updateConfig("auto_deploy", v ? "true" : "false")}
        />

        <ToggleField
          label="Notify on Failure"
          description="Send notification when deployment fails"
          value={config.notify_on_failure !== "false"}
          onChange={(v) => updateConfig("notify_on_failure", v ? "true" : "false")}
        />
      </div>
    </div>
  );
}

function DatadogConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Database className="w-4 h-4" />
          Datadog Configuration
        </h3>
        {config.api_key && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Connected</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">API Key</label>
          <Input type="password" value={config.api_key || ""} onChange={(e) => updateConfig("api_key", e.target.value)} placeholder="xxxxxxxxxxxxxxxxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">App Key</label>
          <Input type="password" value={config.app_key || ""} onChange={(e) => updateConfig("app_key", e.target.value)} placeholder="xxxxxxxxxxxxxxxxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Site</label>
          <Select value={config.site || "datadoghq.com"} onValueChange={(v) => updateConfig("site", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="datadoghq.com">US1 (datadoghq.com)</SelectItem>
              <SelectItem value="datadoghq.eu">EU1 (datadoghq.eu)</SelectItem>
              <SelectItem value="us3.datadoghq.com">US3</SelectItem>
              <SelectItem value="us5.datadoghq.com">US5</SelectItem>
              <SelectItem value="ddog-gov.com">US1-GOV</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <ToggleField
          label="Send Traces"
          description="Include workflow execution traces"
          value={config.send_traces !== "false"}
          onChange={(v) => updateConfig("send_traces", v ? "true" : "false")}
        />

        <ToggleField
          label="Send Logs"
          description="Forward workflow logs to Datadog"
          value={config.send_logs === "true"}
          onChange={(v) => updateConfig("send_logs", v ? "true" : "false")}
        />

        <ToggleField
          label="Send Metrics"
          description="Send custom metrics for monitoring"
          value={config.send_metrics === "true"}
          onChange={(v) => updateConfig("send_metrics", v ? "true" : "false")}
        />
      </div>
    </div>
  );
}

function SendGridConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Mail className="w-4 h-4" />
          SendGrid Configuration
        </h3>
        {config.api_key && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Configured</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">API Key</label>
          <Input type="password" value={config.api_key || ""} onChange={(e) => updateConfig("api_key", e.target.value)} placeholder="SG.xxxxxxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Default From Email</label>
          <Input value={config.from_email || ""} onChange={(e) => updateConfig("from_email", e.target.value)} placeholder="noreply@example.com" className="bg-white/5 border-white/10 text-white" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Default From Name</label>
          <Input value={config.from_name || ""} onChange={(e) => updateConfig("from_name", e.target.value)} placeholder="FunctionFly" className="bg-white/5 border-white/10 text-white" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Template ID (optional)</label>
          <Input value={config.template_id || ""} onChange={(e) => updateConfig("template_id", e.target.value)} placeholder="d-xxxxxxxxxx" className="bg-white/5 border-white/10 text-white" />
        </div>

        <ToggleField
          label="Track Opens"
          description="Enable open tracking in emails"
          value={config.track_opens === "true"}
          onChange={(v) => updateConfig("track_opens", v ? "true" : "false")}
        />

        <ToggleField
          label="Track Clicks"
          description="Enable click tracking in emails"
          value={config.track_clicks === "true"}
          onChange={(v) => updateConfig("track_clicks", v ? "true" : "false")}
        />
      </div>
    </div>
  );
}

function AwsS3ConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <HardDrive className="w-4 h-4" />
          AWS S3 Configuration
        </h3>
        {config.bucket && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Configured</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">AWS Access Key ID</label>
          <Input value={config.access_key_id || ""} onChange={(e) => updateConfig("access_key_id", e.target.value)} placeholder="AKIA..." className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">AWS Secret Access Key</label>
          <Input type="password" value={config.secret_access_key || ""} onChange={(e) => updateConfig("secret_access_key", e.target.value)} placeholder="xxxxxxxxxxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Bucket Name</label>
          <Input value={config.bucket || ""} onChange={(e) => updateConfig("bucket", e.target.value)} placeholder="my-bucket" className="bg-white/5 border-white/10 text-white" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Region</label>
          <Select value={config.region || "us-east-1"} onValueChange={(v) => updateConfig("region", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              {["us-east-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-northeast-1"].map(r => (
                <SelectItem key={r} value={r}>{r}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <ToggleField
          label="Enable SSE-KMS"
          description="Use server-side encryption with KMS"
          value={config.sse_kms === "true"}
          onChange={(v) => updateConfig("sse_kms", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">KMS Key ID (optional)</label>
          <Input value={config.kms_key_id || ""} onChange={(e) => updateConfig("kms_key_id", e.target.value)} placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>
      </div>
    </div>
  );
}

function PostgresConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Database className="w-4 h-4" />
          PostgreSQL Connection
        </h3>
        {config.connection_string && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Connected</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Connection String</label>
          <Input value={config.connection_string || ""} onChange={(e) => updateConfig("connection_string", e.target.value)} placeholder="postgresql://user:pass@host:5432/db" className="bg-white/5 border-white/10 text-white font-mono text-sm" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Max Connections</label>
          <Input type="number" value={config.max_connections || "10"} onChange={(e) => updateConfig("max_connections", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>

        <ToggleField
          label="SSL Mode"
          description="Require SSL for connections"
          value={config.ssl_mode !== "disable"}
          onChange={(v) => updateConfig("ssl_mode", v ? "require" : "disable")}
        />

        <ToggleField
          label="Connection Pooling"
          description="Use PgBouncer for connection pooling"
          value={config.use_pooler === "true"}
          onChange={(v) => updateConfig("use_pooler", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Idle Timeout (seconds)</label>
          <Input type="number" value={config.idle_timeout || "300"} onChange={(e) => updateConfig("idle_timeout", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>
      </div>
    </div>
  );
}

function AIContextConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Brain className="w-4 h-4" />
          AI Context Settings
        </h3>
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Default Model Provider</label>
          <Select value={config.provider || "openai"} onValueChange={(v) => updateConfig("provider", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="openai">OpenAI</SelectItem>
              <SelectItem value="anthropic">Anthropic</SelectItem>
              <SelectItem value="google">Google AI</SelectItem>
              <SelectItem value="azure">Azure OpenAI</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Context Window Size</label>
          <Select value={config.context_window || "medium"} onValueChange={(v) => updateConfig("context_window", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="small">Small (4K tokens)</SelectItem>
              <SelectItem value="medium">Medium (32K tokens)</SelectItem>
              <SelectItem value="large">Large (128K tokens)</SelectItem>
              <SelectItem value="xlarge">Extra Large (200K tokens)</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <ToggleField
          label="Semantic Caching"
          description="Cache semantically similar queries"
          value={config.semantic_cache !== "false"}
          onChange={(v) => updateConfig("semantic_cache", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Cache Similarity Threshold</label>
          <Input type="number" value={config.cache_threshold || "0.85"} onChange={(e) => updateConfig("cache_threshold", e.target.value)} step="0.05" min="0" max="1" className="bg-white/5 border-white/10 text-white" />
          <p className="text-xs text-white/40">Higher = stricter matching (0-1)</p>
        </div>

        <ToggleField
          label="Context Compression"
          description="Compress old context to save tokens"
          value={config.compression === "true"}
          onChange={(v) => updateConfig("compression", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Max Conversation History</label>
          <Input type="number" value={config.max_history || "50"} onChange={(e) => updateConfig("max_history", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>
      </div>
    </div>
  );
}

function WebhookConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Webhook className="w-4 h-4" />
          Webhook Configuration
        </h3>
        {config.webhook_url && <span className="text-xs text-green-400 flex items-center gap-1"><Check className="w-3 h-3" /> Configured</span>}
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Webhook URL</label>
          <Input value={config.webhook_url || ""} onChange={(e) => updateConfig("webhook_url", e.target.value)} placeholder="https://example.com/webhook" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">HTTP Method</label>
          <Select value={config.method || "POST"} onValueChange={(v) => updateConfig("method", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="POST">POST</SelectItem>
              <SelectItem value="PUT">PUT</SelectItem>
              <SelectItem value="PATCH">PATCH</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Authentication Type</label>
          <Select value={config.auth_type || "none"} onValueChange={(v) => updateConfig("auth_type", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="none">None</SelectItem>
              <SelectItem value="bearer">Bearer Token</SelectItem>
              <SelectItem value="basic">Basic Auth</SelectItem>
              <SelectItem value="apikey">API Key</SelectItem>
              <SelectItem value="hmac">HMAC Signature</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {config.auth_type === "bearer" && (
          <div className="space-y-2">
            <label className="text-xs text-white/60">Bearer Token</label>
            <Input type="password" value={config.bearer_token || ""} onChange={(e) => updateConfig("bearer_token", e.target.value)} placeholder="xxxxx" className="bg-white/5 border-white/10 text-white font-mono" />
          </div>
        )}

        {config.auth_type === "basic" && (
          <>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Username</label>
              <Input value={config.basic_username || ""} onChange={(e) => updateConfig("basic_username", e.target.value)} className="bg-white/5 border-white/10 text-white" />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-white/60">Password</label>
              <Input type="password" value={config.basic_password || ""} onChange={(e) => updateConfig("basic_password", e.target.value)} className="bg-white/5 border-white/10 text-white" />
            </div>
          </>
        )}

        {config.auth_type === "apikey" && (
          <div className="space-y-2">
            <label className="text-xs text-white/60">API Key Header Name</label>
            <Input value={config.apikey_header || "X-API-Key"} onChange={(e) => updateConfig("apikey_header", e.target.value)} className="bg-white/5 border-white/10 text-white" />
            <Input type="password" value={config.apikey_value || ""} onChange={(e) => updateConfig("apikey_value", e.target.value)} placeholder="API Key value" className="bg-white/5 border-white/10 text-white font-mono" />
          </div>
        )}

        <ToggleField
          label="Verify SSL"
          description="Validate SSL certificates"
          value={config.verify_ssl !== "false"}
          onChange={(v) => updateConfig("verify_ssl", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Timeout (seconds)</label>
          <Input type="number" value={config.timeout || "30"} onChange={(e) => updateConfig("timeout", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>

        <ToggleField
          label="Retry on Failure"
          description="Automatically retry failed requests"
          value={config.retry_on_failure === "true"}
          onChange={(v) => updateConfig("retry_on_failure", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Retry Count</label>
          <Input type="number" value={config.retry_count || "3"} onChange={(e) => updateConfig("retry_count", e.target.value)} disabled={config.retry_on_failure !== "true"} className="bg-white/5 border-white/10 text-white" />
        </div>
      </div>
    </div>
  );
}

function SchedulerConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Clock className="w-4 h-4" />
          Scheduler Settings
        </h3>
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Default Schedule (Cron Expression)</label>
          <Input value={config.default_cron || "0 * * * *"} onChange={(e) => updateConfig("default_cron", e.target.value)} placeholder="0 * * * *" className="bg-white/5 border-white/10 text-white font-mono" />
          <p className="text-xs text-white/40">Minute • Hour • Day • Month • Weekday</p>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Timezone</label>
          <Select value={config.timezone || "UTC"} onValueChange={(v) => updateConfig("timezone", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              {["UTC", "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", "Europe/London", "Europe/Paris", "Asia/Tokyo", "Asia/Singapore"].map(tz => (
                <SelectItem key={tz} value={tz}>{tz}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <ToggleField
          label="DST Handling"
          description="Automatically adjust for daylight saving time"
          value={config.dst_handling !== "false"}
          onChange={(v) => updateConfig("dst_handling", v ? "true" : "false")}
        />

        <ToggleField
          label="Prevent Overlapping"
          description="Don't run if previous execution is still running"
          value={config.prevent_overlap === "true"}
          onChange={(v) => updateConfig("prevent_overlap", v ? "true" : "false")}
        />

        <ToggleField
          label="Catchup Missed Runs"
          description="Run missed executions when scheduler recovers"
          value={config.catchup_missed === "true"}
          onChange={(v) => updateConfig("catchup_missed", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Max Concurrent Runs</label>
          <Input type="number" value={config.max_concurrent || "1"} onChange={(e) => updateConfig("max_concurrent", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>
      </div>
    </div>
  );
}

function RateLimiterConfigSection({ config, updateConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white flex items-center gap-2">
          <Gauge className="w-4 h-4" />
          Rate Limiter Settings
        </h3>
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <label className="text-xs text-white/60">Algorithm</label>
          <Select value={config.algorithm || "token_bucket"} onValueChange={(v) => updateConfig("algorithm", v)}>
            <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
            <SelectContent className="bg-[#1a1a2e] border-white/10">
              <SelectItem value="token_bucket">Token Bucket</SelectItem>
              <SelectItem value="sliding_window">Sliding Window</SelectItem>
              <SelectItem value="fixed_window">Fixed Window</SelectItem>
              <SelectItem value="leaky_bucket">Leaky Bucket</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Requests per Minute</label>
          <Input type="number" value={config.requests_per_minute || "60"} onChange={(e) => updateConfig("requests_per_minute", e.target.value)} className="bg-white/5 border-white/10 text-white" />
        </div>

        <div className="space-y-2">
          <label className="text-xs text-white/60">Burst Size</label>
          <Input type="number" value={config.burst_size || "10"} onChange={(e) => updateConfig("burst_size", e.target.value)} className="bg-white/5 border-white/10 text-white" />
          <p className="text-xs text-white/40">Maximum requests that can be made in a burst</p>
        </div>

        <ToggleField
          label="Distributed Mode"
          description="Use Redis for distributed rate limiting"
          value={config.distributed === "true"}
          onChange={(v) => updateConfig("distributed", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Key Prefix</label>
          <Input value={config.key_prefix || "rate_limit:"} onChange={(e) => updateConfig("key_prefix", e.target.value)} placeholder="rate_limit:" className="bg-white/5 border-white/10 text-white font-mono" />
        </div>

        <ToggleField
          label="Return 429 on Limit"
          description="Return HTTP 429 when limit exceeded"
          value={config.return_429 !== "false"}
          onChange={(v) => updateConfig("return_429", v ? "true" : "false")}
        />

        <div className="space-y-2">
          <label className="text-xs text-white/60">Block Duration (seconds)</label>
          <Input type="number" value={config.block_duration || "60"} onChange={(e) => updateConfig("block_duration", e.target.value)} className="bg-white/5 border-white/10 text-white" />
          <p className="text-xs text-white/40">How long to block after exceeding limit</p>
        </div>
      </div>
    </div>
  );
}

function GenericConfigSection({ config, updateConfig, setConfig }: ConfigSectionProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-sm font-medium text-white">Plugin Settings</h3>
      <div className="space-y-2">
        <label className="text-xs text-white/60">Plugin Enabled</label>
        <ToggleField
          label="Enable Plugin"
          description="Allow this plugin to run"
          value={config.enabled !== "false"}
          onChange={(v) => updateConfig("enabled", v ? "true" : "false")}
        />
      </div>

      <div className="space-y-2">
        <label className="text-xs text-white/60">Logging Level</label>
        <Select value={config.log_level || "info"} onValueChange={(v) => updateConfig("log_level", v)}>
          <SelectTrigger className="bg-white/5 border-white/10 text-white"><SelectValue /></SelectTrigger>
          <SelectContent className="bg-[#1a1a2e] border-white/10">
            <SelectItem value="debug">Debug</SelectItem>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="warn">Warning</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <details className="group">
        <summary className="text-xs text-white/40 hover:text-white/60 cursor-pointer list-none flex items-center gap-1">
          <span>Advanced: Edit as JSON</span>
          <span className="text-white/20 group-open:rotate-90 transition-transform">▶</span>
        </summary>
        <div className="mt-2">
          <textarea
            value={JSON.stringify(config, null, 2)}
            onChange={(e) => {
              try {
                const parsed = JSON.parse(e.target.value);
                if (setConfig) {
                  setConfig(parsed);
                } else {
                  Object.entries(parsed).forEach(([k, v]) => {
                    updateConfig(k, typeof v === "string" ? v : JSON.stringify(v));
                  });
                }
              } catch {}
            }}
            className="w-full h-40 bg-white/5 border border-white/10 rounded-lg p-3 text-xs font-mono text-white/80 resize-none"
          />
        </div>
      </details>
    </div>
  );
}

// ============================================================================
// Helper Components
// ============================================================================

interface ToggleFieldProps {
  label: string;
  description: string;
  value: boolean;
  onChange: (value: boolean) => void;
}

function ToggleField({ label, description, value, onChange }: ToggleFieldProps) {
  return (
    <div className="flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/10">
      <div>
        <p className="text-sm text-white">{label}</p>
        <p className="text-xs text-white/40">{description}</p>
      </div>
      <button
        type="button"
        onClick={() => onChange(!value)}
        className={cn(
          "relative w-11 h-6 rounded-full transition-colors",
          value ? "bg-green-500" : "bg-white/20"
        )}
      >
        <span className={cn("absolute top-1 left-1 w-4 h-4 rounded-full bg-white transition-transform", value && "translate-x-5")} />
      </button>
    </div>
  );
}