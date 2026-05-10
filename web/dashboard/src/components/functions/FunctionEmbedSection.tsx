import { useState, useCallback, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import {
  Copy,
  Check,
  Package,
  ChevronDown,
  ChevronUp,
  Globe,
  Palette,
  Code2,
  Terminal,
  Zap,
  FileCode2,
  Eye,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

// =============================================================================
// Types
// =============================================================================

interface FunctionEmbedSectionProps {
  author: string;
  name: string;
  version?: string;
  className?: string;
}

// =============================================================================
// Helpers
// =============================================================================

const EMBED_BASE = "https://functionfly.com/embed";

function buildEmbedUrl(
  author: string,
  name: string,
  opts: { namespace: string; ui: boolean; theme: string; autoload: boolean }
): string {
  const params = new URLSearchParams();
  if (opts.namespace !== "ff") params.set("namespace", opts.namespace);
  if (opts.ui) params.set("ui", "true");
  if (opts.theme !== "auto") params.set("theme", opts.theme);
  if (!opts.autoload) params.set("autoload", "false");
  const qs = params.toString();
  return `${EMBED_BASE}/${author}/${name}.js${qs ? `?${qs}` : ""}`;
}

function buildPinnedUrl(
  author: string,
  name: string,
  version: string,
  opts: { namespace: string; ui: boolean; theme: string; autoload: boolean }
): string {
  const params = new URLSearchParams();
  if (opts.namespace !== "ff") params.set("namespace", opts.namespace);
  if (opts.ui) params.set("ui", "true");
  if (opts.theme !== "auto") params.set("theme", opts.theme);
  if (!opts.autoload) params.set("autoload", "false");
  const qs = params.toString();
  return `${EMBED_BASE}/${author}/${name}@${version}.js${qs ? `?${qs}` : ""}`;
}

function buildSnippetTag(url: string): string {
  return `<script src="${url}"></script>`;
}

// =============================================================================
// Copy helper with fallback
// =============================================================================

async function writeToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // fall through
  }
  try {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    const ok = document.execCommand("copy");
    document.body.removeChild(ta);
    return ok;
  } catch {
    return false;
  }
}

// =============================================================================
// Syntax highlighter style (module-level)
// =============================================================================

const codeStyle = {
  ...vscDarkPlus,
  "pre[class*='language-']": {
    ...vscDarkPlus["pre[class*='language-']"],
    background: "transparent",
    margin: 0,
    padding: "0.75rem 1rem",
    fontSize: "12px",
    lineHeight: "1.5",
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
  },
  "code[class*='language-']": {
    ...vscDarkPlus["code[class*='language-']"],
    background: "transparent",
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
    fontSize: "12px",
    lineHeight: "1.5",
  },
};

// =============================================================================
// Snippet display with copy
// =============================================================================

function SnippetBlock({
  code,
  language,
  label,
}: {
  code: string;
  language: string;
  label: string;
}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    const ok = await writeToClipboard(code);
    if (ok) {
      setCopied(true);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    }
  }, [code]);

  return (
    <div className="relative group">
      <div className="flex items-center justify-between px-3 py-1.5 border-b border-border-subtle bg-bg-tertiary/40 rounded-t-lg">
        <span className="text-[10px] uppercase tracking-wider text-text-muted font-medium">
          {label}
        </span>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 px-2 text-[10px] gap-1 text-text-muted hover:text-text-primary opacity-60 group-hover:opacity-100 transition-opacity"
          onClick={handleCopy}
        >
          {copied ? (
            <>
              <Check className="w-3 h-3 text-emerald-400" />
              <span className="text-emerald-400">Copied</span>
            </>
          ) : (
            <>
              <Copy className="w-3 h-3" />
              Copy
            </>
          )}
        </Button>
      </div>
      <div className="bg-[#0d0d12] rounded-b-lg overflow-x-auto">
        <SyntaxHighlighter language={language} style={codeStyle} wrapLines wrapLongLines>
          {code}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}

// =============================================================================
// Usage examples
// =============================================================================

function getUsageExamples(
  author: string,
  name: string,
  ns: string,
  inputExample: string
) {
  const runCode = `// Promise-based
const result = await ${ns}.run(${inputExample});
console.log(result.data);`;

  const callbackCode = `// Callback-based
${ns}.run(${inputExample}, {
  onSuccess: (data) => console.log(data),
  onError: (err) => console.error(err),
});`;

  const formCode = `<form id="myForm">
  <input name="text" placeholder="Enter text..." />
  <button type="submit">Submit</button>
</form>

<script>
${ns}.form(document.getElementById("myForm"), {
  onSuccess: (data) => alert("Result: " + JSON.stringify(data)),
  onError: (err) => alert("Error: " + err.message),
});
</script>`;

  const widgetCode = `<div id="ff-widget"></div>
<script>
${ns}.widget(document.getElementById("ff-widget"), {
  title: "${name}",
  placeholder: "Enter input...",
  buttonText: "Run",
  onSuccess: (data) => console.log(data),
});
</script>`;

  const eventsCode = `${ns}.on("ready", () => console.log("Embed loaded"));
${ns}.on("execute:start", (input) => console.log("Running...", input));
${ns}.on("execute:success", (result) => console.log("Done!", result));
${ns}.on("execute:error", (err) => console.error("Failed:", err));`;

  return { runCode, callbackCode, formCode, widgetCode, eventsCode };
}

// =============================================================================
// Main Component
// =============================================================================

export function FunctionEmbedSection({
  author,
  name,
  version,
  className,
}: FunctionEmbedSectionProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [namespace, setNamespace] = useState("ff");
  const [uiEnabled, setUiEnabled] = useState(false);
  const [theme, setTheme] = useState<"light" | "dark" | "auto">("auto");
  const [autoload, setAutoload] = useState(true);
  const [copied, setCopied] = useState(false);

  // Validate namespace: must be a valid JS identifier
  const safeNamespace = useMemo(() => {
    const trimmed = namespace.trim();
    return /^[a-zA-Z_$][a-zA-Z0-9_$]*$/.test(trimmed) ? trimmed : "ff";
  }, [namespace]);

  const opts = useMemo(
    () => ({ namespace: safeNamespace, ui: uiEnabled, theme, autoload }),
    [safeNamespace, uiEnabled, theme, autoload]
  );

  const embedUrl = useMemo(() => buildEmbedUrl(author, name, opts), [author, name, opts]);
  const pinnedUrl = useMemo(
    () => (version ? buildPinnedUrl(author, name, version, opts) : null),
    [author, name, version, opts]
  );
  const snippet = useMemo(() => buildSnippetTag(embedUrl), [embedUrl]);
  const pinnedSnippet = useMemo(
    () => (pinnedUrl ? buildSnippetTag(pinnedUrl) : null),
    [pinnedUrl]
  );

  const examples = useMemo(
    () =>
      getUsageExamples(author, name, safeNamespace, '{ text: "Hello world" }'),
    [author, name, safeNamespace]
  );

  const handleCopySnippet = useCallback(async () => {
    const ok = await writeToClipboard(snippet);
    if (ok) {
      setCopied(true);
      toast.success("Embed code copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    }
  }, [snippet]);

  // Resolve final URL for display
  const displayUrl = embedUrl.replace("https://functionfly.com", "");

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.65 }}
      className={cn("function-page-section", className)}
    >
      <div
        className={cn(
          "border border-border-subtle rounded-xl overflow-hidden bg-bg-secondary transition-shadow duration-300",
          !isExpanded && "cursor-pointer hover:border-border-default",
          isExpanded && "shadow-lg shadow-black/10"
        )}
        onClick={() => {
          if (!isExpanded) setIsExpanded(true);
        }}
        role={!isExpanded ? "button" : undefined}
        tabIndex={!isExpanded ? 0 : undefined}
        onKeyDown={(e) => {
          if (!isExpanded && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            setIsExpanded(true);
          }
        }}
        aria-label={!isExpanded ? "Embed this function on your website" : undefined}
      >
        {/* ─── Collapsed Header ─── */}
        {!isExpanded && (
          <div className="flex items-center justify-between gap-4 px-5 py-4">
            <div className="flex items-center gap-3 min-w-0">
              <div className="w-10 h-10 rounded-lg bg-brand-500/10 border border-brand-500/20 flex items-center justify-center shrink-0">
                <Package className="w-5 h-5 text-brand-500" />
              </div>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold text-text-primary">
                    Embed this function
                  </span>
                  <Badge
                    variant="outline"
                    className="text-[10px] h-4 px-1.5 font-mono border-emerald-500/30 text-emerald-400 bg-emerald-500/10"
                  >
                    One script tag
                  </Badge>
                </div>
                <p className="text-xs text-text-muted mt-0.5">
                  Run {author}/{name} on any website — no backend required
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <span className="text-xs text-brand-400 font-medium hidden sm:inline">
                Get embed code
              </span>
              <ChevronDown className="w-4 h-4 text-brand-400" />
            </div>
          </div>
        )}

        {/* ─── Expanded Content ─── */}
        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.3, ease: "easeInOut" }}
              className="overflow-hidden"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Header bar */}
              <div className="flex items-center justify-between px-5 py-3 border-b border-border-subtle bg-bg-tertiary/40">
                <div className="flex items-center gap-3">
                  <Package className="w-5 h-5 text-brand-500" />
                  <div>
                    <h3 className="text-sm font-semibold text-text-primary">
                      Embed this function
                    </h3>
                    <p className="text-xs text-text-muted">
                      One script tag. No backend required.
                    </p>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 w-7 p-0 text-text-muted hover:text-text-primary"
                  onClick={() => setIsExpanded(false)}
                  aria-label="Collapse embed section"
                >
                  <ChevronUp className="w-4 h-4" />
                </Button>
              </div>

              <div className="p-5 space-y-6">
                {/* Configuration Grid */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div className="space-y-1.5">
                    <Label
                      htmlFor="embed-namespace"
                      className="text-xs text-text-secondary flex items-center gap-1.5"
                    >
                      <Code2 className="w-3 h-3" />
                      Namespace
                    </Label>
                    <Input
                      id="embed-namespace"
                      value={namespace}
                      onChange={(e) => setNamespace(e.target.value)}
                      placeholder="ff"
                      className="h-8 text-xs font-mono"
                    />
                    {namespace !== safeNamespace && (
                      <p className="text-[10px] text-red-400">
                        Must be a valid JS identifier (e.g. ff, myApp)
                      </p>
                    )}
                  </div>

                  <div className="space-y-1.5">
                    <Label
                      htmlFor="embed-theme"
                      className="text-xs text-text-secondary flex items-center gap-1.5"
                    >
                      <Palette className="w-3 h-3" />
                      Theme
                    </Label>
                    <Select
                      value={theme}
                      onValueChange={(v) =>
                        setTheme(v as "light" | "dark" | "auto")
                      }
                    >
                      <SelectTrigger id="embed-theme" className="h-8 text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="auto">Auto (System)</SelectItem>
                        <SelectItem value="dark">Dark</SelectItem>
                        <SelectItem value="light">Light</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="flex items-end gap-3">
                    <div className="space-y-1.5">
                      <Label className="text-xs text-text-secondary cursor-pointer">
                        UI Widget
                      </Label>
                      <div className="flex items-center gap-2 h-8">
                        <Switch
                          checked={uiEnabled}
                          onCheckedChange={setUiEnabled}
                          className="scale-75 origin-left"
                        />
                        <span className="text-xs text-text-muted">
                          {uiEnabled ? "On" : "Off"}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-end gap-3">
                    <div className="space-y-1.5">
                      <Label className="text-xs text-text-secondary cursor-pointer">
                        Autoload
                      </Label>
                      <div className="flex items-center gap-2 h-8">
                        <Switch
                          checked={autoload}
                          onCheckedChange={setAutoload}
                          className="scale-75 origin-left"
                        />
                        <span className="text-xs text-text-muted">
                          {autoload ? "On" : "Off"}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Generated Embed Code */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs text-text-secondary">
                      Embed Code
                    </Label>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant="outline"
                        className="text-[9px] h-4 px-1.5 font-mono border-border-subtle text-text-muted"
                      >
                        {displayUrl}
                      </Badge>
                    </div>
                  </div>
                  <div className="relative group">
                    <div className="bg-[#0d0d12] border border-border-subtle rounded-lg p-4 font-mono text-xs overflow-x-auto">
                      <code className="text-emerald-400 whitespace-pre-wrap break-all">
                        {snippet}
                      </code>
                    </div>
                    <Button
                      size="sm"
                      className="absolute top-2 right-2 h-7 px-2.5 gap-1.5 text-xs bg-brand-500 hover:bg-brand-600 text-white"
                      onClick={handleCopySnippet}
                    >
                      {copied ? (
                        <>
                          <Check className="w-3.5 h-3.5" />
                          Copied
                        </>
                      ) : (
                        <>
                          <Copy className="w-3.5 h-3.5" />
                          Copy
                        </>
                      )}
                    </Button>
                  </div>

                  {/* Pinned version */}
                  {pinnedSnippet && pinnedSnippet !== snippet && (
                    <div className="space-y-1">
                      <Label className="text-[10px] text-text-muted">
                        Pinned to v{version}
                      </Label>
                      <div className="bg-bg-tertiary/50 border border-border-subtle rounded-lg p-3 font-mono text-[11px] overflow-x-auto">
                        <code className="text-text-muted">{pinnedSnippet}</code>
                      </div>
                    </div>
                  )}
                </div>

                {/* Usage Examples */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Zap className="w-4 h-4 text-brand-500" />
                    <span className="text-sm font-semibold text-text-primary">
                      Usage Examples
                    </span>
                  </div>

                  <Tabs defaultValue="run" className="w-full">
                    <TabsList className="h-8 w-full justify-start gap-0.5 bg-transparent border-b border-border-subtle rounded-none px-0">
                      <TabsTrigger
                        value="run"
                        className="rounded-none rounded-t-md text-[11px] h-7 px-3 data-[state=active]:bg-bg-tertiary data-[state=active]:border data-[state=active]:border-border-subtle data-[state=active]:border-b-transparent"
                      >
                        <Terminal className="w-3 h-3 mr-1.5" />
                        Run
                      </TabsTrigger>
                      <TabsTrigger
                        value="form"
                        className="rounded-none rounded-t-md text-[11px] h-7 px-3 data-[state=active]:bg-bg-tertiary data-[state=active]:border data-[state=active]:border-border-subtle data-[state=active]:border-b-transparent"
                      >
                        <FileCode2 className="w-3 h-3 mr-1.5" />
                        Form
                      </TabsTrigger>
                      <TabsTrigger
                        value="widget"
                        className="rounded-none rounded-t-md text-[11px] h-7 px-3 data-[state=active]:bg-bg-tertiary data-[state=active]:border data-[state=active]:border-border-subtle data-[state=active]:border-b-transparent"
                      >
                        <Globe className="w-3 h-3 mr-1.5" />
                        Widget
                      </TabsTrigger>
                      <TabsTrigger
                        value="events"
                        className="rounded-none rounded-t-md text-[11px] h-7 px-3 data-[state=active]:bg-bg-tertiary data-[state=active]:border data-[state=active]:border-border-subtle data-[state=active]:border-b-transparent"
                      >
                        <Zap className="w-3 h-3 mr-1.5" />
                        Events
                      </TabsTrigger>
                      <TabsTrigger
                        value="preview"
                        className="rounded-none rounded-t-md text-[11px] h-7 px-3 data-[state=active]:bg-bg-tertiary data-[state=active]:border data-[state=active]:border-border-subtle data-[state=active]:border-b-transparent"
                      >
                        <Eye className="w-3 h-3 mr-1.5" />
                        Preview
                      </TabsTrigger>
                    </TabsList>

                    <TabsContent value="run" className="mt-0">
                      <SnippetBlock
                        code={examples.runCode}
                        language="javascript"
                        label="Programmatic execution"
                      />
                    </TabsContent>
                    <TabsContent value="form" className="mt-0">
                      <SnippetBlock
                        code={examples.formCode}
                        language="html"
                        label="HTML form binding"
                      />
                    </TabsContent>
                    <TabsContent value="widget" className="mt-0">
                      <SnippetBlock
                        code={examples.widgetCode}
                        language="html"
                        label="UI widget mount"
                      />
                    </TabsContent>
                    <TabsContent value="events" className="mt-0">
                      <SnippetBlock
                        code={examples.eventsCode}
                        language="javascript"
                        label="Lifecycle events"
                      />
                    </TabsContent>
                    <TabsContent value="preview" className="mt-0">
                      <div className="space-y-2">
                        <div className="flex items-center justify-between px-3 py-1.5 border-b border-border-subtle bg-bg-tertiary/40 rounded-t-lg">
                          <span className="text-[10px] uppercase tracking-wider text-text-muted font-medium">
                            Live Preview
                          </span>
                          <span className="text-[10px] text-text-muted">
                            {uiEnabled ? 'Widget mode' : 'Run mode'}
                          </span>
                        </div>
                        <div className="bg-[#0d0d12] rounded-b-lg overflow-hidden" style={{ minHeight: '300px' }}>
                          <iframe
                            src={`/embed/${author}/${name}?namespace=${safeNamespace}&ui=${uiEnabled}&theme=${theme}&autoload=${autoload}`}
                            className="w-full h-[300px] border-0"
                            title="Embed Preview"
                            sandbox="allow-scripts allow-same-origin"
                          />
                        </div>
                        <p className="text-[10px] text-text-muted text-center">
                          Preview shows a live demo. Actual embed requires the script tag on your page.
                        </p>
                      </div>
                    </TabsContent>
                  </Tabs>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </motion.div>
  );
}

FunctionEmbedSection.displayName = "FunctionEmbedSection";

export default FunctionEmbedSection;
