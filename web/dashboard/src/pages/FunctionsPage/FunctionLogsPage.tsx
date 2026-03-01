import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  ArrowLeft,
  Search,
  Download,
  RefreshCw,
  Filter,
  AlertTriangle,
  Info,
  XCircle,
  Clock,
  Calendar,
  ChevronDown,
  ChevronUp,
  Copy,
  Expand,
  X,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import "@/styles/components.css";

interface LogEntry {
  id: string;
  timestamp: string;
  level: "info" | "warn" | "error" | "debug";
  message: string;
  source: string;
  metadata?: Record<string, unknown>;
  requestId?: string;
  duration?: number;
}

interface LogStats {
  total: number;
  errors: number;
  warnings: number;
  info: number;
}

export function FunctionLogsPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  
  // Filters
  const [searchQuery, setSearchQuery] = useState("");
  const [levelFilter, setLevelFilter] = useState<string>("all");
  const [sourceFilter, setSourceFilter] = useState<string>("all");
  const [timeRange, setTimeRange] = useState<string>("1h");
  
  // Stats
  const [stats, setStats] = useState<LogStats>({ total: 0, errors: 0, warnings: 0, info: 0 });
  
  // Selected log for detail view
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);
  const [showDetailPanel, setShowDetailPanel] = useState(false);
  
  // Sort order
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  // Fetch logs
  const fetchLogs = async (showSpinner = false) => {
    if (!author || !name) return;

    try {
      if (showSpinner) setIsRefreshing(true);
      
      const response = await apiClient.get<{ logs: LogEntry[] }>(
        `/v1/functions/${author}/${name}/logs`,
        {
          params: {
            timeRange,
            level: levelFilter !== "all" ? levelFilter : undefined,
            source: sourceFilter !== "all" ? sourceFilter : undefined,
          },
        }
      );
      
      const logData = response.logs || [];
      setLogs(logData);
      
      // Calculate stats
      const newStats = {
        total: logData.length,
        errors: logData.filter((l) => l.level === "error").length,
        warnings: logData.filter((l) => l.level === "warn").length,
        info: logData.filter((l) => l.level === "info").length,
      };
      setStats(newStats);
    } catch (err) {
      console.error("Failed to load logs:", err);
      setError("Failed to load logs");
      toast.error("Failed to load logs");
    } finally {
      setLoading(false);
      if (showSpinner) setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchLogs();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(() => fetchLogs(true), 30000);
    return () => clearInterval(interval);
  }, [author, name, timeRange, levelFilter, sourceFilter]);

  // Apply search and sorting
  useEffect(() => {
    let filtered = [...logs];
    
    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(
        (log) =>
          log.message.toLowerCase().includes(query) ||
          log.source.toLowerCase().includes(query) ||
          log.requestId?.toLowerCase().includes(query)
      );
    }
    
    // Apply level filter
    if (levelFilter !== "all") {
      filtered = filtered.filter((log) => log.level === levelFilter);
    }
    
    // Apply source filter
    if (sourceFilter !== "all") {
      filtered = filtered.filter((log) => log.source === sourceFilter);
    }
    
    // Apply sorting
    filtered.sort((a, b) => {
      const dateA = new Date(a.timestamp).getTime();
      const dateB = new Date(b.timestamp).getTime();
      return sortOrder === "desc" ? dateB - dateA : dateA - dateB;
    });
    
    setFilteredLogs(filtered);
  }, [logs, searchQuery, levelFilter, sourceFilter, sortOrder]);

  const handleRefresh = () => {
    fetchLogs(true);
  };

  const handleExport = () => {
    const dataStr = JSON.stringify(filteredLogs, null, 2);
    const blob = new Blob([dataStr], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `function-logs-${author}-${name}-${new Date().toISOString()}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    toast.success("Logs exported successfully");
  };

  const getLogIcon = (level: string) => {
    switch (level) {
      case "error":
        return <XCircle className="w-4 h-4 text-red-400" />;
      case "warn":
        return <AlertTriangle className="w-4 h-4 text-yellow-400" />;
      case "info":
        return <Info className="w-4 h-4 text-blue-400" />;
      default:
        return <div className="w-4 h-4" />;
    }
  };

  const getLevelBadgeColor = (level: string) => {
    switch (level) {
      case "error":
        return "bg-red-500/10 text-red-400 border-red-500/20";
      case "warn":
        return "bg-yellow-500/10 text-yellow-400 border-yellow-500/20";
      case "info":
        return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      default:
        return "bg-gray-500/10 text-gray-400 border-gray-500/20";
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  const uniqueSources = Array.from(new Set(logs.map((log) => log.source)));

  if (loading && !logs.length) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 bg-muted rounded animate-pulse" />
          <div className="space-y-2">
            <div className="w-48 h-6 bg-muted rounded animate-pulse" />
            <div className="w-32 h-4 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="p-6 border rounded-lg">
          <div className="w-full h-96 bg-muted rounded animate-pulse" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate(`/functions`)}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">
                {author}/{name}
              </h1>
              <Badge variant="secondary">Logs</Badge>
            </div>
            <p className="text-text-secondary">
              View and analyze function execution logs
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${isRefreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button variant="outline" onClick={handleExport}>
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4">
        <Card className="card">
          <CardContent className="card-content p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Total Logs</p>
                <p className="text-2xl font-bold text-text-primary">{stats.total}</p>
              </div>
              <div className="p-2 bg-primary/10 rounded-lg">
                <Clock className="w-5 h-5 text-primary" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Errors</p>
                <p className="text-2xl font-bold text-red-400">{stats.errors}</p>
              </div>
              <div className="p-2 bg-red-500/10 rounded-lg">
                <XCircle className="w-5 h-5 text-red-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Warnings</p>
                <p className="text-2xl font-bold text-yellow-400">{stats.warnings}</p>
              </div>
              <div className="p-2 bg-yellow-500/10 rounded-lg">
                <AlertTriangle className="w-5 h-5 text-yellow-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="card">
          <CardContent className="card-content p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-text-muted">Info</p>
                <p className="text-2xl font-bold text-blue-400">{stats.info}</p>
              </div>
              <div className="p-2 bg-blue-500/10 rounded-lg">
                <Info className="w-5 h-5 text-blue-400" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card className="card">
        <CardContent className="card-content p-4">
          <div className="flex items-center gap-4 flex-wrap">
            {/* Search */}
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search logs..."
                className="pl-9"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
                  onClick={() => setSearchQuery("")}
                >
                  <X className="w-3 h-3" />
                </Button>
              )}
            </div>

            {/* Time Range */}
            <Select value={timeRange} onValueChange={setTimeRange}>
              <SelectTrigger className="w-[140px]">
                <Clock className="w-4 h-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="15m">Last 15 minutes</SelectItem>
                <SelectItem value="1h">Last 1 hour</SelectItem>
                <SelectItem value="6h">Last 6 hours</SelectItem>
                <SelectItem value="24h">Last 24 hours</SelectItem>
                <SelectItem value="7d">Last 7 days</SelectItem>
              </SelectContent>
            </Select>

            {/* Level Filter */}
            <Select value={levelFilter} onValueChange={setLevelFilter}>
              <SelectTrigger className="w-[140px]">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue placeholder="Level" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Levels</SelectItem>
                <SelectItem value="error">Error</SelectItem>
                <SelectItem value="warn">Warning</SelectItem>
                <SelectItem value="info">Info</SelectItem>
                <SelectItem value="debug">Debug</SelectItem>
              </SelectContent>
            </Select>

            {/* Source Filter */}
            <Select value={sourceFilter} onValueChange={setSourceFilter}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="Source" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Sources</SelectItem>
                {uniqueSources.map((source) => (
                  <SelectItem key={source} value={source}>
                    {source}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Sort Order */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => setSortOrder(sortOrder === "desc" ? "asc" : "desc")}
            >
              {sortOrder === "desc" ? (
                <>
                  <ChevronDown className="w-4 h-4 mr-2" />
                  Newest
                </>
              ) : (
                <>
                  <ChevronUp className="w-4 h-4 mr-2" />
                  Oldest
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Logs List */}
      <Card className="card">
        <CardHeader className="card-header">
          <CardTitle className="card-title">
            Log Entries ({filteredLogs.length})
          </CardTitle>
        </CardHeader>
        <CardContent className="card-content p-0">
          {filteredLogs.length === 0 ? (
            <div className="p-8 text-center text-text-muted">
              <Filter className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p>No logs found matching your filters</p>
            </div>
          ) : (
            <ScrollArea className="h-[500px]">
              <div className="divide-y divide-border">
                {filteredLogs.map((log) => (
                  <div
                    key={log.id}
                    className="p-4 hover:bg-bg-tertiary transition-colors cursor-pointer"
                    onClick={() => {
                      setSelectedLog(log);
                      setShowDetailPanel(true);
                    }}
                  >
                    <div className="flex items-start gap-3">
                      <div className="mt-0.5">{getLogIcon(log.level)}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <Badge
                            variant="outline"
                            className={`text-xs ${getLevelBadgeColor(log.level)}`}
                          >
                            {log.level.toUpperCase()}
                          </Badge>
                          <span className="text-xs text-text-muted font-mono">
                            {log.source}
                          </span>
                          {log.requestId && (
                            <span className="text-xs text-text-muted font-mono">
                              {log.requestId}
                            </span>
                          )}
                          {log.duration && (
                            <span className="text-xs text-text-muted">
                              {log.duration}ms
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-text-primary break-words">
                          {log.message}
                        </p>
                      </div>
                      <div className="text-xs text-text-muted whitespace-nowrap">
                        <div className="flex items-center gap-1">
                          <Calendar className="w-3 h-3" />
                          {formatTimestamp(log.timestamp)}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
        </CardContent>
      </Card>

      {/* Detail Panel */}
      {showDetailPanel && selectedLog && (
        <div className="fixed inset-0 z-50 flex justify-end bg-black/50">
          <div className="w-[500px] h-full bg-bg-secondary border-l border-border p-6 overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-text-primary">Log Details</h2>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setShowDetailPanel(false)}
              >
                <X className="w-4 h-4" />
              </Button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="text-sm text-text-muted">Level</label>
                <Badge
                  variant="outline"
                  className={`mt-1 ${getLevelBadgeColor(selectedLog.level)}`}
                >
                  {selectedLog.level.toUpperCase()}
                </Badge>
              </div>

              <div>
                <label className="text-sm text-text-muted">Timestamp</label>
                <p className="text-text-primary mt-1">
                  {formatTimestamp(selectedLog.timestamp)}
                </p>
              </div>

              <div>
                <label className="text-sm text-text-muted">Source</label>
                <p className="text-text-primary mt-1">{selectedLog.source}</p>
              </div>

              {selectedLog.requestId && (
                <div>
                  <label className="text-sm text-text-muted">Request ID</label>
                  <div className="flex items-center gap-2 mt-1">
                    <code className="text-text-primary font-mono text-sm">
                      {selectedLog.requestId}
                    </code>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => copyToClipboard(selectedLog.requestId!)}
                    >
                      <Copy className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              )}

              {selectedLog.duration && (
                <div>
                  <label className="text-sm text-text-muted">Duration</label>
                  <p className="text-text-primary mt-1">{selectedLog.duration}ms</p>
                </div>
              )}

              <div>
                <label className="text-sm text-text-muted">Message</label>
                <div className="flex items-start gap-2 mt-1">
                  <p className="text-text-primary break-words flex-1">
                    {selectedLog.message}
                  </p>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 flex-shrink-0"
                    onClick={() => copyToClipboard(selectedLog.message)}
                  >
                    <Copy className="w-3 h-3" />
                  </Button>
                </div>
              </div>

              {selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && (
                <div>
                  <label className="text-sm text-text-muted">Metadata</label>
                  <pre className="mt-1 p-3 bg-bg-tertiary rounded-lg text-xs overflow-x-auto">
                    {JSON.stringify(selectedLog.metadata, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default FunctionLogsPage;
