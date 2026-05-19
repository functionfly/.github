import { useState } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Download, RefreshCw, Check, AlertTriangle, Clock, HardDrive,
  Zap, ArrowUp, ArrowDown, Github, Tag, Package, Monitor, Smartphone,
  CheckCircle2, XCircle, Info, ExternalLink
} from "lucide-react";

interface UpdateInfo {
  version: string;
  releaseDate: Date;
  size: string;
  type: "major" | "minor" | "patch" | "hotfix";
  changelog: string[];
  mandatory: boolean;
  breakingChanges: boolean;
}

interface UpdateChannel {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
}

const currentVersion = "2.4.2";

const updateInfo: UpdateInfo = {
  version: "2.5.0",
  releaseDate: new Date(Date.now() + 86400000 * 3),
  size: "245 MB",
  type: "minor",
  changelog: [
    "New WebGPU renderer for 3x faster graph rendering",
    "Voice control support for hands-free operation",
    "Improved neural network visualization",
    "12 performance improvements and bug fixes",
    "New keyboard shortcuts for node manipulation",
  ],
  mandatory: false,
  breakingChanges: false,
};

const updateChannels: UpdateChannel[] = [
  { id: "stable", name: "Stable", description: "Recommended for production", icon: <CheckCircle2 className="w-4 h-4 text-emerald-400" /> },
  { id: "beta", name: "Beta", description: "Early access to new features", icon: <Zap className="w-4 h-4 text-orange-400" /> },
  { id: "nightly", name: "Nightly", description: "Latest builds, potentially unstable", icon: <AlertTriangle className="w-4 h-4 text-red-400" /> },
];

const changelogHistory = [
  {
    version: "2.4.2",
    date: new Date(Date.now() - 86400000 * 14),
    type: "patch",
    changes: ["Fixed memory leak in plugin manager", "Improved graph export performance"],
  },
  {
    version: "2.4.1",
    date: new Date(Date.now() - 86400000 * 21),
    type: "patch",
    changes: ["Fixed crash when loading large graphs", "UI responsiveness improvements"],
  },
  {
    version: "2.4.0",
    date: new Date(Date.now() - 86400000 * 45),
    type: "minor",
    changes: [
      "New experimental feature lab",
      "Universal search across workspace",
      "Workspace snapshot manager",
      "Enhanced keyboard shortcut visualizer",
    ],
  },
];

export function StudioUpdateManager() {
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [activeChannel, setActiveChannel] = useState("stable");
  const [autoUpdate, setAutoUpdate] = useState(true);
  const [checkOnStartup, setCheckOnStartup] = useState(true);
  const [updateDownloaded, setUpdateDownloaded] = useState(false);

  const handleDownload = async () => {
    setIsDownloading(true);
    setDownloadProgress(0);
    for (let i = 0; i <= 100; i += 5) {
      await new Promise((r) => setTimeout(r, 200));
      setDownloadProgress(i);
    }
    setIsDownloading(false);
    setUpdateDownloaded(true);
  };

  const handleInstall = () => {
    console.log("Installing update...");
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "major":
        return <ArrowUp className="w-4 h-4 text-red-400" />;
      case "minor":
        return <ArrowUp className="w-4 h-4 text-orange-400" />;
      case "patch":
        return <RefreshCw className="w-4 h-4 text-blue-400" />;
      case "hotfix":
        return <Zap className="w-4 h-4 text-yellow-400" />;
      default:
        return <RefreshCw className="w-4 h-4 text-white/40" />;
    }
  };

  const daysUntilRelease = Math.ceil((updateInfo.releaseDate.getTime() - Date.now()) / 86400000);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
            <Download className="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">Update Manager</h2>
            <p className="text-sm text-white/60">
              Current version: <span className="text-white font-mono">{currentVersion}</span>
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Button size="sm" variant="outline" className="gap-2">
            <RefreshCw className="w-4 h-4" />
            Check for Updates
          </Button>
        </div>
      </div>

      <Tabs defaultValue="update" className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="update"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Download className="h-4 w-4 shrink-0" />
              Update
            </TabsTrigger>
            <TabsTrigger
              value="channels"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Zap className="h-4 w-4 shrink-0" />
              Channels
            </TabsTrigger>
            <TabsTrigger
              value="changelog"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Clock className="h-4 w-4 shrink-0" />
              Changelog
            </TabsTrigger>
            <TabsTrigger
              value="settings"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <HardDrive className="h-4 w-4 shrink-0" />
              Settings
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="update" className="mt-0">
            <div className="space-y-6">
              <GlassCard className="p-6">
                <div className="flex items-start gap-4 mb-6">
                  <div className="w-14 h-14 rounded-xl bg-orange-500/20 flex items-center justify-center shrink-0">
                    <Package className="w-7 h-7 text-orange-400" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-lg font-semibold text-white">
                        Studio {updateInfo.version}
                      </h3>
                      <Badge className="text-[10px] bg-orange-500/20 text-orange-400 border-orange-500/30">
                        {updateInfo.type.toUpperCase()}
                      </Badge>
                      {updateInfo.mandatory && (
                        <Badge className="text-[10px] bg-red-500/20 text-red-400 border-red-500/30">
                          REQUIRED
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm text-white/60 mb-2">
                      {daysUntilRelease > 0
                        ? `Releasing in ${daysUntilRelease} day${daysUntilRelease > 1 ? "s" : ""}`
                        : "Available now"}
                    </p>
                    <div className="flex items-center gap-4 text-xs text-white/50">
                      <span className="flex items-center gap-1">
                        <Tag className="w-3 h-3" />
                        {updateInfo.size}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {formatDate(updateInfo.releaseDate)}
                      </span>
                    </div>
                  </div>
                </div>

                {isDownloading ? (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-white/80">Downloading...</span>
                      <span className="text-white/60">{downloadProgress}%</span>
                    </div>
                    <Progress value={downloadProgress} className="h-2" />
                  </div>
                ) : updateDownloaded ? (
                  <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 text-emerald-400">
                      <CheckCircle2 className="w-5 h-5" />
                      <span className="text-sm font-medium">Update ready to install</span>
                    </div>
                    <Button onClick={handleInstall} className="ml-auto gap-2 bg-emerald-500 hover:bg-emerald-400">
                      <Zap className="w-4 h-4" />
                      Install & Restart
                    </Button>
                  </div>
                ) : (
                  <div className="flex items-center gap-3">
                    <Button onClick={handleDownload} className="gap-2 bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-400 hover:to-red-400">
                      <Download className="w-4 h-4" />
                      Download Update
                    </Button>
                    <Button variant="outline" className="gap-2">
                      <Info className="w-4 h-4" />
                      Learn More
                    </Button>
                  </div>
                )}
              </GlassCard>

              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4 flex items-center gap-2">
                  <Zap className="w-4 h-4 text-orange-400" />
                  What's New in {updateInfo.version}
                </h3>
                <ul className="space-y-3">
                  {updateInfo.changelog.map((change, i) => (
                    <li key={i} className="flex items-start gap-3 text-sm">
                      <div className="w-1.5 h-1.5 rounded-full bg-orange-400 mt-2 shrink-0" />
                      <span className="text-white/80">{change}</span>
                    </li>
                  ))}
                </ul>
              </GlassCard>
            </div>
          </TabsContent>

          <TabsContent value="channels" className="mt-0">
            <div className="space-y-4">
              <div className="grid gap-4">
                {updateChannels.map((channel) => (
                  <GlassCard
                    key={channel.id}
                    className={cn(
                      "p-4 cursor-pointer transition-all duration-200",
                      activeChannel === channel.id
                        ? "ring-2 ring-orange-500/30 bg-white/10"
                        : "hover:bg-white/5"
                    )}
                    onClick={() => setActiveChannel(channel.id)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-lg bg-white/5 flex items-center justify-center">
                          {channel.icon}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-medium text-white">{channel.name}</h3>
                            {channel.id === "stable" && (
                              <Badge variant="outline" className="text-[10px] text-white/50">
                                Recommended
                              </Badge>
                            )}
                          </div>
                          <p className="text-sm text-white/60">{channel.description}</p>
                        </div>
                      </div>
                      {activeChannel === channel.id && (
                        <CheckCircle2 className="w-5 h-5 text-orange-400" />
                      )}
                    </div>
                  </GlassCard>
                ))}
              </div>

              <GlassCard className="p-4">
                <div className="flex items-center gap-3 text-sm text-white/60">
                  <Info className="w-4 h-4" />
                  <span>
                    Switching channels may require a restart. Beta and Nightly builds
                    receive updates more frequently but may be less stable.
                  </span>
                </div>
              </GlassCard>
            </div>
          </TabsContent>

          <TabsContent value="changelog" className="mt-0">
            <div className="space-y-4">
              {changelogHistory.map((release, i) => (
                <GlassCard key={release.version} className="p-5">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                      {getTypeIcon(release.type)}
                      <h3 className="font-semibold text-white">Version {release.version}</h3>
                      <Badge variant="outline" className="text-[10px] text-white/50">
                        {release.type}
                      </Badge>
                    </div>
                    <span className="text-xs text-white/40">{formatDate(release.date)}</span>
                  </div>
                  <ul className="space-y-2">
                    {release.changes.map((change, j) => (
                      <li key={j} className="flex items-start gap-2 text-sm text-white/70">
                        <div className="w-1 h-1 rounded-full bg-white/30 mt-2" />
                        {change}
                      </li>
                    ))}
                  </ul>
                  {i < changelogHistory.length - 1 && (
                    <a href="#" className="inline-flex items-center gap-1 text-xs text-orange-400 hover:text-orange-300 mt-3">
                      View full changelog <ExternalLink className="w-3 h-3" />
                    </a>
                  )}
                </GlassCard>
              ))}
            </div>
          </TabsContent>

          <TabsContent value="settings" className="mt-0">
            <div className="space-y-4 max-w-xl">
              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4">Update Behavior</h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Zap className="w-5 h-5 text-orange-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Automatic updates</p>
                        <p className="text-xs text-white/60">Download and install updates automatically</p>
                      </div>
                    </div>
                    <Switch checked={autoUpdate} onCheckedChange={setAutoUpdate} />
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <RefreshCw className="w-5 h-5 text-blue-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Check on startup</p>
                        <p className="text-xs text-white/60">Automatically check for updates when Studio starts</p>
                      </div>
                    </div>
                    <Switch checked={checkOnStartup} onCheckedChange={setCheckOnStartup} />
                  </div>
                </div>
              </GlassCard>

              <GlassCard className="p-5">
                <h3 className="font-semibold text-white mb-4">Update History</h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-lg bg-white/5">
                    <div className="flex items-center gap-3">
                      <Monitor className="w-5 h-5 text-white/40" />
                      <div>
                        <p className="text-sm text-white/80">Studio {currentVersion}</p>
                        <p className="text-xs text-white/50">Current version</p>
                      </div>
                    </div>
                    <Badge className="bg-emerald-500/20 text-emerald-400 border-emerald-500/30">Up to date</Badge>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-white/5">
                    <div className="flex items-center gap-3">
                      <Smartphone className="w-5 h-5 text-white/40" />
                      <div>
                        <p className="text-sm text-white/80">Mobile companion app</p>
                        <p className="text-xs text-white/50">v1.3.1 - Last updated 7 days ago</p>
                      </div>
                    </div>
                    <Badge variant="outline" className="text-white/50">No update</Badge>
                  </div>
                </div>
              </GlassCard>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}