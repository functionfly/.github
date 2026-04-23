import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { embedApi, type EmbedSnippet, type EmbedSnippetParams } from "@/api/embed";
import { Copy, Check, Code, Loader2 } from "lucide-react";
import "@/styles/components.css";

interface EmbedCodeGeneratorProps {
  author: string;
  name: string;
  version?: string;
  className?: string;
}

export function EmbedCodeGenerator({
  author,
  name,
  version,
  className = "",
}: EmbedCodeGeneratorProps) {
  const { t } = useTranslation();
  const [namespace, setNamespace] = useState("ff");
  const [autoload, setAutoload] = useState(true);
  const [uiEnabled, setUiEnabled] = useState(false);
  const [theme, setTheme] = useState<"light" | "dark" | "auto">("auto");
  const [snippet, setSnippet] = useState<EmbedSnippet | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  // Fetch embed snippet when parameters change
  useEffect(() => {
    const fetchSnippet = async () => {
      setLoading(true);
      try {
        const params: EmbedSnippetParams = {
          namespace,
          autoload,
          ui: uiEnabled,
          theme,
        };
        const response = await embedApi.getEmbedSnippet(author, name, params);
        setSnippet(response);
      } catch (error) {
        console.error("Failed to fetch embed snippet:", error);
        // Generate fallback snippet locally
        const queryParams = new URLSearchParams();
        if (uiEnabled) queryParams.append("ui", "true");
        if (theme !== "auto") queryParams.append("theme", theme);

        const baseUrl = `${window.location.origin}/embed/${author}/${name}.js`;
        const queryString = queryParams.toString();
        const fullUrl = queryString ? `${baseUrl}?${queryString}` : baseUrl;

        setSnippet({
          snippet: `<script src="${fullUrl}"></script>`,
          pinned_snippet: version
            ? `<script src="${baseUrl.replace('.js', `@${version}.js`)}${queryString ? `?${queryString}` : ""}"></script>`
            : `<script src="${fullUrl}"></script>`,
        });
      } finally {
        setLoading(false);
      }
    };

    const debounce = setTimeout(fetchSnippet, 300);
    return () => clearTimeout(debounce);
  }, [author, name, version, namespace, autoload, uiEnabled, theme]);

  const handleCopy = async () => {
    if (!snippet) return;

    try {
      await navigator.clipboard.writeText(snippet.snippet);
      setCopied(true);
      toast.success(t('embedCode.toastCopied'));
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error("Failed to copy:", error);
      toast.error(t('embedCode.toastCopyFailed'));
    }
  };

  return (
    <Card className={`card ${className}`}>
      <CardHeader className="card-header">
        <CardTitle className="card-title flex items-center gap-2">
          <Code className="w-5 h-5" />
          {t('embedCode.title')}
        </CardTitle>
      </CardHeader>
      <CardContent className="card-content space-y-4">
        {/* Configuration Options */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="namespace" className="text-text-secondary">
              {t('embedCode.namespace')}
            </Label>
            <Input
              id="namespace"
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              placeholder="ff"
              className="font-mono"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="theme" className="text-text-secondary">
              {t('embedCode.theme')}
            </Label>
            <Select
              value={theme}
              onValueChange={(value) => setTheme(value as "light" | "dark" | "auto")}
            >
              <SelectTrigger id="theme">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">{t('embedCode.themeLight')}</SelectItem>
                <SelectItem value="dark">{t('embedCode.themeDark')}</SelectItem>
                <SelectItem value="auto">{t('embedCode.themeAuto')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="ui-toggle" className="text-text-secondary cursor-pointer">
              {t('embedCode.uiWidget')}
            </Label>
            <Switch
              id="ui-toggle"
              checked={uiEnabled}
              onCheckedChange={setUiEnabled}
            />
          </div>

          <div className="flex items-center justify-between space-x-2">
            <Label htmlFor="autoload-toggle" className="text-text-secondary cursor-pointer">
              {t('embedCode.autoload')}
            </Label>
            <Switch
              id="autoload-toggle"
              checked={autoload}
              onCheckedChange={setAutoload}
            />
          </div>
        </div>

        {/* Generated Code Snippet */}
        <div className="space-y-2">
          <Label className="text-text-secondary">{t('embedCode.embedCode')}</Label>
          <div className="relative">
            <div className="bg-bg-secondary border rounded-lg p-4 font-mono text-sm overflow-x-auto">
              {loading ? (
                <div className="flex items-center gap-2 text-text-muted">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  {t('embedCode.generating')}
                </div>
              ) : (
                <code className="text-text-primary whitespace-pre-wrap break-all">
                  {snippet?.snippet || t('embedCode.unableToGenerate')}
                </code>
              )}
            </div>
            <Button
              size="sm"
              variant="outline"
              className="absolute top-2 right-2"
              onClick={handleCopy}
              disabled={loading || !snippet}
            >
              {copied ? (
                <Check className="w-4 h-4 text-green-500" />
              ) : (
                <Copy className="w-4 h-4" />
              )}
              <span className="ml-2">{copied ? t('embedCode.copied') : t('embedCode.copy')}</span>
            </Button>
          </div>
        </div>

        {/* Pinned Version Info */}
        {version && snippet && (
          <div className="space-y-2">
            <Label className="text-text-secondary">{t('embedCode.pinnedVersion')}</Label>
            <div className="bg-bg-secondary border rounded-lg p-3 font-mono text-xs overflow-x-auto">
              <code className="text-text-muted">{snippet.pinned_snippet}</code>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
