import { useState } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import {
  Bell, BellOff, CheckCheck, X, AlertCircle, AlertTriangle, Info,
  CheckCircle2, XCircle, Clock, Archive, Trash2, Settings, Volume2,
  MessageSquare, Bug, Zap, Users, Bookmark
} from "lucide-react";

interface Notification {
  id: string;
  type: "info" | "success" | "warning" | "error" | "update";
  title: string;
  message: string;
  timestamp: Date;
  read: boolean;
  category: "system" | "workflow" | "plugin" | "billing" | "community";
  actionUrl?: string;
}

const mockNotifications: Notification[] = [
  {
    id: "n1",
    type: "success",
    title: "Graph execution completed",
    message: "Customer Churn Prediction graph finished successfully in 2.4s",
    timestamp: new Date(Date.now() - 300000),
    read: false,
    category: "workflow",
    actionUrl: "/graphs/customer-churn",
  },
  {
    id: "n2",
    type: "info",
    title: "Plugin update available",
    message: "GitHub Integration v2.1.0 is available with bug fixes",
    timestamp: new Date(Date.now() - 3600000),
    read: false,
    category: "plugin",
  },
  {
    id: "n3",
    type: "warning",
    title: "Low balance alert",
    message: "Your wallet balance is below $5.00. Add funds to avoid service interruption.",
    timestamp: new Date(Date.now() - 7200000),
    read: true,
    category: "billing",
  },
  {
    id: "n4",
    type: "error",
    title: "Execution failed",
    message: "Image Classification Pipeline failed: Out of memory at node 'Resize'",
    timestamp: new Date(Date.now() - 86400000),
    read: true,
    category: "workflow",
  },
  {
    id: "n5",
    type: "update",
    title: "Studio 2.5.0 released",
    message: "New WebGPU renderer, voice control, and 12 other improvements",
    timestamp: new Date(Date.now() - 172800000),
    read: true,
    category: "system",
  },
  {
    id: "n6",
    type: "info",
    title: "New comment on your post",
    message: "Sarah Chen commented on 'Best practices for graph design'",
    timestamp: new Date(Date.now() - 259200000),
    read: true,
    category: "community",
  },
];

const notificationSettings = {
  system: { email: true, push: true, inApp: true },
  workflow: { email: false, push: true, inApp: true },
  plugin: { email: true, push: false, inApp: true },
  billing: { email: true, push: true, inApp: true },
  community: { email: false, push: false, inApp: true },
};

export function GlobalNotificationCenter() {
  const [activeTab, setActiveTab] = useState("all");
  const [notifications, setNotifications] = useState<Notification[]>(mockNotifications);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [settings, setSettings] = useState(notificationSettings);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const markAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, read: true } : n))
    );
  };

  const markAllAsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  };

  const deleteNotification = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const clearAll = () => {
    setNotifications([]);
  };

  const filteredNotifications = notifications.filter((n) => {
    if (activeTab === "all") return true;
    if (activeTab === "unread") return !n.read;
    return n.category === activeTab;
  });

  const getTypeIcon = (type: Notification["type"]) => {
    switch (type) {
      case "success":
        return <CheckCircle2 className="w-5 h-5 text-emerald-400" />;
      case "error":
        return <XCircle className="w-5 h-5 text-red-400" />;
      case "warning":
        return <AlertTriangle className="w-5 h-5 text-yellow-400" />;
      case "info":
        return <Info className="w-5 h-5 text-blue-400" />;
      case "update":
        return <Zap className="w-5 h-5 text-purple-400" />;
    }
  };

  const getCategoryBadge = (category: Notification["category"]) => {
    const styles = {
      system: "bg-blue-500/20 text-blue-400 border-blue-500/30",
      workflow: "bg-orange-500/20 text-orange-400 border-orange-500/30",
      plugin: "bg-purple-500/20 text-purple-400 border-purple-500/30",
      billing: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
      community: "bg-pink-500/20 text-pink-400 border-pink-500/30",
    };
    return styles[category];
  };

  const formatTime = (date: Date) => {
    const diff = Date.now() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return `${days}d ago`;
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="relative">
            <div className="w-10 h-10 rounded-xl bg-blue-500/20 flex items-center justify-center">
              <Bell className="w-5 h-5 text-blue-400" />
            </div>
            {unreadCount > 0 && (
              <div className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-red-500 text-white text-xs flex items-center justify-center font-bold">
                {unreadCount}
              </div>
            )}
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">Notifications</h2>
            <p className="text-sm text-white/60">
              {unreadCount > 0 ? `${unreadCount} unread` : "All caught up"}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {unreadCount > 0 && (
            <Button size="sm" variant="ghost" onClick={markAllAsRead} className="gap-1 text-white/60 hover:text-white">
              <CheckCheck className="w-4 h-4" />
              Mark all read
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={clearAll} className="gap-1 text-white/60 hover:text-white">
            <Trash2 className="w-4 h-4" />
            Clear
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="all"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Bell className="h-4 w-4 shrink-0" />
              All
            </TabsTrigger>
            <TabsTrigger
              value="unread"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <AlertCircle className="h-4 w-4 shrink-0" />
              Unread
              {unreadCount > 0 && (
                <Badge className="ml-1 text-[10px] bg-red-500/20 text-red-400 border-red-500/30">
                  {unreadCount}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger
              value="system"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Settings className="h-4 w-4 shrink-0" />
              System
            </TabsTrigger>
            <TabsTrigger
              value="workflow"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Zap className="h-4 w-4 shrink-0" />
              Workflows
            </TabsTrigger>
            <TabsTrigger
              value="billing"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <AlertTriangle className="h-4 w-4 shrink-0" />
              Billing
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value={activeTab} className="mt-0">
            {filteredNotifications.length === 0 ? (
              <GlassCard className="flex flex-col items-center justify-center h-48">
                <BellOff className="w-10 h-10 text-white/30 mb-3" />
                <p className="text-white/60">No notifications</p>
                <p className="text-sm text-white/40">You're all caught up!</p>
              </GlassCard>
            ) : (
              <div className="space-y-2">
                {filteredNotifications.map((notification) => (
                  <GlassCard
                    key={notification.id}
                    className={cn(
                      "p-4 transition-all duration-200",
                      !notification.read && "bg-white/[0.07] border-l-2 border-l-orange-500"
                    )}
                  >
                    <div className="flex items-start gap-4">
                      <div className="w-10 h-10 rounded-xl bg-white/5 flex items-center justify-center shrink-0">
                        {getTypeIcon(notification.type)}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between gap-2 mb-1">
                          <div className="flex items-center gap-2">
                            <h3 className={cn("font-medium", notification.read ? "text-white/80" : "text-white")}>
                              {notification.title}
                            </h3>
                            {!notification.read && (
                              <div className="w-2 h-2 rounded-full bg-orange-400" />
                            )}
                          </div>
                          <div className="flex items-center gap-1 shrink-0">
                            <button
                              onClick={() => markAsRead(notification.id)}
                              className="p-1 rounded text-white/30 hover:text-white/60 hover:bg-white/5"
                              title="Mark as read"
                            >
                              <CheckCheck className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => deleteNotification(notification.id)}
                              className="p-1 rounded text-white/30 hover:text-red-400 hover:bg-red-400/10"
                              title="Delete"
                            >
                              <X className="w-4 h-4" />
                            </button>
                          </div>
                        </div>
                        <p className="text-sm text-white/60 line-clamp-2 mb-2">
                          {notification.message}
                        </p>
                        <div className="flex items-center gap-3">
                          <span className="text-xs text-white/40 flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {formatTime(notification.timestamp)}
                          </span>
                          <Badge variant="outline" className={cn("text-[10px]", getCategoryBadge(notification.category))}>
                            {notification.category}
                          </Badge>
                          {notification.actionUrl && (
                            <a
                              href={notification.actionUrl}
                              className="text-xs text-orange-400 hover:text-orange-300"
                            >
                              View →
                            </a>
                          )}
                        </div>
                      </div>
                    </div>
                  </GlassCard>
                ))}
              </div>
            )}
          </TabsContent>
        </div>
      </Tabs>

      <div className="p-4 border-t border-white/10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Volume2 className="w-4 h-4 text-white/40" />
            <span className="text-sm text-white/60">Notification sound</span>
          </div>
          <Switch checked={soundEnabled} onCheckedChange={setSoundEnabled} />
        </div>
      </div>
    </div>
  );
}