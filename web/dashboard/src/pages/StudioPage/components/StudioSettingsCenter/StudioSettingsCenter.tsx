import type { StudioSettings } from "@/api/studioSettings";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { useStudioSettings } from "@/hooks/useStudioSettings";
import { cn } from "@/lib/utils";
import { useKeyboardShortcutsStore } from "@/stores/keyboardShortcutsStore";
import { useThemeStore } from "@/stores/themeStore";
import { Badge, Button, GlassCard, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@functionfly/ui-core";
import {
  Bell,
  Check,
  Eye,
  Globe,
  Keyboard,
  Layout,
  LayoutDashboard,
  Monitor,
  Moon,
  Palette,
  RotateCcw, Save, Search,
  Shield,
  Sparkles,
  Sun, Type,
  Volume2,
  Wifi
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";

interface SettingSection {
  id: string;
  label: string;
  icon: React.ReactNode;
  description: string;
}

const settingSections: SettingSection[] = [
  { id: "appearance", label: "Appearance", icon: <Palette className="w-4 h-4" />, description: "Theme, colors, and visual settings" },
  { id: "layout", label: "Layout", icon: <Layout className="w-4 h-4" />, description: "Workspace and editor layout" },
  { id: "editor", label: "Editor", icon: <Type className="w-4 h-4" />, description: "Editor preferences and behavior" },
  { id: "shortcuts", label: "Shortcuts", icon: <Keyboard className="w-4 h-4" />, description: "Keyboard shortcuts and bindings" },
  { id: "notifications", label: "Notifications", icon: <Bell className="w-4 h-4" />, description: "Alerts and notification preferences" },
  { id: "privacy", label: "Privacy & Security", icon: <Shield className="w-4 h-4" />, description: "Security and privacy settings" },
  { id: "performance", label: "Performance", icon: <Monitor className="w-4 h-4" />, description: "Resource usage and optimization" },
  { id: "network", label: "Network", icon: <Wifi className="w-4 h-4" />, description: "Connection and proxy settings" },
];

const colorPresets = [
  { id: "orange", name: "FunctionFly Orange", primary: "#f97316", secondary: "#ea580c" },
  { id: "blue", name: "Ocean Blue", primary: "#3b82f6", secondary: "#1d4ed8" },
  { id: "purple", name: "Violet Dream", primary: "#8b5cf6", secondary: "#6d28d9" },
  { id: "emerald", name: "Forest Green", primary: "#10b981", secondary: "#059669" },
  { id: "rose", name: "Rose Gold", primary: "#f43f5e", secondary: "#e11d48" },
  { id: "slate", name: "Midnight Slate", primary: "#64748b", secondary: "#475569" },
];

const applyColorPalette = (presetId: string) => {
  const preset = colorPresets.find((p) => p.id === presetId);
  if (preset) {
    console.log('[StudioSettings] Applying color palette:', preset.id, preset.primary);

    // Calculate color variations for gradient scale
    const lighter = adjustColorBrightness(preset.primary, 40);
    const darker = adjustColorBrightness(preset.primary, -40);

    // Set all brand-related CSS variables (Tailwind uses --color-brand-N)
    document.documentElement.style.setProperty('--color-brand-400', lighter);
    document.documentElement.style.setProperty('--color-brand-500', preset.primary);
    document.documentElement.style.setProperty('--color-brand-600', preset.secondary);
    document.documentElement.style.setProperty('--color-brand-700', darker);

    // Also set velocity variants if they exist
    document.documentElement.style.setProperty('--color-velocity-400', lighter);
    document.documentElement.style.setProperty('--color-velocity-500', preset.primary);
    document.documentElement.style.setProperty('--color-velocity-600', preset.secondary);

    // Core accent variables
    document.documentElement.style.setProperty('--text-accent', preset.primary);
    document.documentElement.style.setProperty('--button-primary', preset.primary);
    document.documentElement.style.setProperty('--button-primary-hover', preset.secondary);
    document.documentElement.style.setProperty('--border-focus', preset.primary + '80');

    // Inject dynamic CSS to override Tailwind's brand color classes globally
    // This ensures ALL elements using brand-500/400 etc get the new color
    let styleEl = document.getElementById('dynamic-brand-colors');
    if (!styleEl) {
      styleEl = document.createElement('style');
      styleEl.id = 'dynamic-brand-colors';
      document.head.appendChild(styleEl);
    }
    styleEl.textContent = `
      /* Brand color overrides - Studio specific */
      [data-studio] .text-brand-400,
      .studio-root .text-brand-400 { color: var(--color-brand-400) !important; }
      [data-studio] .text-brand-300,
      .studio-root .text-brand-300 { color: var(--color-brand-300) !important; }
      [data-studio] .bg-brand-500,
      .studio-root .bg-brand-500 { background-color: var(--color-brand-500) !important; }
      [data-studio] .border-brand-500,
      .studio-root .border-brand-500 { border-color: var(--color-brand-500) !important; }
      /* Generic fallback - color variables */
      .text-brand-400 { color: var(--color-brand-400, ${preset.primary}) !important; }
      .text-brand-300 { color: var(--color-brand-300, ${lighter}) !important; }
      .bg-brand-500 { background-color: var(--color-brand-500, ${preset.primary}) !important; }
      .border-brand-500 { border-color: var(--color-brand-500, ${preset.primary}) !important; }
    `;
    // Also apply via CSS variables for any component using var()
    document.documentElement.style.setProperty('--studio-brand-primary', preset.primary);
    document.documentElement.style.setProperty('--studio-brand-secondary', preset.secondary);
  }
};

// Helper to adjust color brightness (positive = lighter, negative = darker)
const adjustColorBrightness = (hex: string, amount: number): string => {
  const num = parseInt(hex.replace('#', ''), 16);
  const r = Math.min(255, Math.max(0, (num >> 16) + amount));
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00FF) + amount));
  const b = Math.min(255, Math.max(0, (num & 0x0000FF) + amount));
  return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
};

const fontSizes = [
  { value: 12, label: "12px (Compact)" },
  { value: 14, label: "14px (Default)" },
  { value: 16, label: "16px (Large)" },
  { value: 18, label: "18px (Extra Large)" },
];

export function StudioSettingsCenter() {
  const { settings, defaultSettings, isLoading, isDirty, updateSetting, saveSettings, resetSettings, isSaving } = useStudioSettings();
  const { setTheme } = useThemeStore();
  const { shortcuts, globalShortcuts } = useKeyboardShortcutsStore();
  const [activeTab, setActiveTab] = useState("appearance");
  const [searchQuery, setSearchQuery] = useState("");
  const [localSettings, setLocalSettings] = useState<StudioSettings | null>(null);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    if (settings && !localSettings) {
      console.log('[StudioSettingsCenter] Settings loaded, applying palette:', settings.primary_color);
      setLocalSettings(settings);
      setHasChanges(false);
      applyColorPalette(settings.primary_color);

      // Apply font size on load
      const newSize = String(settings.font_size) + 'px';
      document.documentElement.style.setProperty('--studio-font-size', newSize);

      let fontSizeStyle = document.getElementById('dynamic-font-size');
      if (!fontSizeStyle) {
        fontSizeStyle = document.createElement('style');
        fontSizeStyle.id = 'dynamic-font-size';
        document.head.appendChild(fontSizeStyle);
      }
      fontSizeStyle.textContent = `
        body, body * { font-size: ${newSize} !important; }
        .studio-root, .studio-root * { font-size: ${newSize} !important; }
      `;
    }
  }, [settings, localSettings]);

  const handleUpdate = useCallback((key: keyof StudioSettings, value: any) => {
    console.log('[StudioSettingsCenter] handleUpdate:', key, value);
    setLocalSettings((prev) => ({ ...prev, [key]: value }));
    setHasChanges(true);

    if (key === 'theme') {
      setTheme(value as 'dark' | 'light' | 'system');
    }
    if (key === 'primary_color') {
      applyColorPalette(value);
    }
    if (key === 'font_size') {
      const newSize = String(value) + 'px';
      console.log('[StudioSettings] Setting font size:', newSize);

      // Set on document root
      document.documentElement.style.setProperty('--studio-font-size', newSize);

      // Inject dynamic CSS with high specificity
      let fontSizeStyle = document.getElementById('dynamic-font-size');
      if (!fontSizeStyle) {
        fontSizeStyle = document.createElement('style');
        fontSizeStyle.id = 'dynamic-font-size';
        document.head.appendChild(fontSizeStyle);
      }
      fontSizeStyle.textContent = `
        body, body * {
          font-size: ${newSize} !important;
        }
        .studio-root, .studio-root * {
          font-size: ${newSize} !important;
        }
      `;
    }
  }, [setTheme]);

  const handleEditorFeatureUpdate = useCallback((feature: keyof StudioSettings['editor_features'], value: boolean) => {
    setLocalSettings((prev) => ({
      ...prev,
      editor_features: { ...prev.editor_features, [feature]: value },
    }));
    setHasChanges(true);
  }, []);

  const handleSave = useCallback(async () => {
    console.log('[StudioSettingsCenter] Saving settings:', localSettings);
    if (!localSettings) {
      console.error('[StudioSettingsCenter] No settings to save!');
      return;
    }
    try {
      await saveSettings(localSettings);
      console.log('[StudioSettingsCenter] Settings saved successfully');
      setHasChanges(false);
    } catch (err) {
      console.error('[StudioSettingsCenter] Failed to save settings:', err);
    }
  }, [localSettings, saveSettings]);

  const handleReset = useCallback(async () => {
    await resetSettings();
    setLocalSettings(defaultSettings);
    setHasChanges(false);
  }, [resetSettings, defaultSettings]);

  const filteredSections = settingSections.filter(
    (s) =>
      s.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const shortcutsByCategory = {
    global: globalShortcuts,
    navigation: shortcuts.filter(s => s.category === 'navigation'),
    actions: shortcuts.filter(s => s.category === 'actions'),
    editor: shortcuts.filter(s => s.category === 'editor'),
    playground: shortcuts.filter(s => s.category === 'playground'),
  };

  const renderContent = () => {
    if (isLoading || !localSettings) {
      return (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500" />
        </div>
      );
    }

    switch (activeTab) {
      case "appearance":
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Theme</h3>
              <div className="grid grid-cols-3 gap-4">
                {[
                  { id: "dark", label: "Dark", icon: <Moon className="w-5 h-5" /> },
                  { id: "light", label: "Light", icon: <Sun className="w-5 h-5" /> },
                  { id: "system", label: "System", icon: <Monitor className="w-5 h-5" /> },
                ].map((t) => (
                  <button
                    key={t.id}
                    onClick={() => handleUpdate('theme', t.id)}
                    className={cn(
                      "flex flex-col items-center gap-2 p-4 rounded-xl border transition-all duration-200",
                      localSettings.theme === t.id
                        ? "bg-bg-hover border-border-focus text-text-primary"
                        : "bg-bg-secondary border-border-subtle text-text-secondary hover:bg-bg-hover hover:border-border-primary"
                    )}
                  >
                    {t.icon}
                    <span className="text-sm font-medium">{t.label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Color Palette</h3>
              <div className="grid grid-cols-2 gap-4">
                {colorPresets.map((preset) => (
                  <button
                    key={preset.id}
                    onClick={() => handleUpdate('primary_color', preset.id)}
                    className={cn(
                      "relative flex flex-row items-center gap-3 p-3 rounded-xl border transition-all duration-200",
                      localSettings.primary_color === preset.id
                        ? "bg-bg-hover border-border-focus"
                        : "bg-bg-secondary border-border-subtle hover:bg-bg-hover"
                    )}
                  >
                    <div
                      className="w-10 h-10 rounded-lg shadow-lg shrink-0"
                      style={{ background: `linear-gradient(135deg, ${preset.primary}, ${preset.secondary})` }}
                    />
                    <div className="min-w-0 flex-1 text-left">
                      <p className="text-sm font-medium text-text-primary leading-tight">{preset.name}</p>
                      <p className="text-xs text-text-muted">{preset.primary}</p>
                    </div>
                    {localSettings.primary_color === preset.id && (
                      <Check className="w-5 h-5 shrink-0" style={{ color: 'var(--text-accent)' }} />
                    )}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Font Size</h3>
              <div className="relative h-[120px] flex items-center justify-center">
                <div className="absolute inset-0 flex items-center justify-center overflow-hidden">
                  <div
                    className="flex gap-4 items-center justify-center"
                    style={{
                      transform: `perspective(1000px) rotateX(${15}deg)`,
                      transformStyle: 'preserve-3d',
                    }}
                  >
                    {fontSizes.map((size) => {
                      const isSelected = localSettings.font_size === size.value;
                      const offset = localSettings.font_size - size.value;
                      return (
                        <button
                          key={size.value}
                          onClick={() => handleUpdate('font_size', size.value)}
                          className={cn(
                            "relative flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-300",
                            isSelected
                              ? "bg-bg-hover border-border-focus scale-110 z-10"
                              : "bg-bg-secondary border-border-subtle hover:bg-bg-hover scale-90 opacity-60"
                          )}
                          style={{
                            transform: `translateZ(${offset * 20}px) scale(${1 - Math.abs(offset) * 0.15})`,
                            opacity: 1 - Math.abs(offset) * 0.3,
                          }}
                        >
                          <span
                            className="text-2xl font-bold mb-1"
                            style={{
                              color: isSelected ? 'var(--text-accent)' : 'var(--text-primary)',
                              fontSize: size.value + 'px',
                            }}
                          >
                            Aa
                          </span>
                          <span className="text-xs text-text-muted">{size.value}px</span>
                          {isSelected && (
                            <div
                              className="absolute -bottom-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full"
                              style={{ backgroundColor: 'var(--text-accent)' }}
                            />
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
                <div className="absolute bottom-0 flex items-center gap-2">
                  <button
                    onClick={() => {
                      const currentIndex = fontSizes.findIndex(f => f.value === localSettings.font_size);
                      if (currentIndex > 0) {
                        handleUpdate('font_size', fontSizes[currentIndex - 1].value);
                      }
                    }}
                    className="p-2 rounded-lg bg-bg-secondary border border-border-subtle hover:bg-bg-hover text-text-muted"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <span className="text-sm text-text-muted px-2">{localSettings.font_size}px</span>
                  <button
                    onClick={() => {
                      const currentIndex = fontSizes.findIndex(f => f.value === localSettings.font_size);
                      if (currentIndex < fontSizes.length - 1) {
                        handleUpdate('font_size', fontSizes[currentIndex + 1].value);
                      }
                    }}
                    className="p-2 rounded-lg bg-bg-secondary border border-border-subtle hover:bg-bg-hover text-text-muted"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-text-primary">Visual Effects</h3>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
                  <div className="flex items-center gap-3">
                    <Sparkles className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                    <div>
                      <p className="text-sm font-medium text-text-primary">Animations</p>
                      <p className="text-xs text-text-muted">Enable smooth transitions and effects</p>
                    </div>
                  </div>
                  <Switch
                    checked={localSettings.animations_enabled}
                    onCheckedChange={(v) => handleUpdate('animations_enabled', v)}
                  />
                </div>
                <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
                  <div className="flex items-center gap-3">
                    <Layout className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                    <div>
                      <p className="text-sm font-medium text-text-primary">Transparency</p>
                      <p className="text-xs text-text-muted">Glassmorphism and backdrop blur effects</p>
                    </div>
                  </div>
                  <Switch
                    checked={localSettings.transparency_enabled}
                    onCheckedChange={(v) => handleUpdate('transparency_enabled', v)}
                  />
                </div>
              </div>
            </div>
          </div>
        );

      case "layout":
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Sidebar Position</h3>
              <div className="grid grid-cols-2 gap-4">
                <button
                  onClick={() => handleUpdate('sidebar_position', "left")}
                  className={cn(
                    "p-4 rounded-xl border transition-all duration-200",
                    localSettings.sidebar_position === "left"
                      ? "bg-bg-hover border-border-focus"
                      : "bg-bg-secondary border-border-subtle hover:bg-bg-hover"
                  )}
                >
                  <LayoutDashboard className="w-6 h-6 text-text-secondary mb-2" />
                  <p className="text-sm font-medium text-text-primary">Left</p>
                </button>
                <button
                  onClick={() => handleUpdate('sidebar_position', "right")}
                  className={cn(
                    "p-4 rounded-xl border transition-all duration-200",
                    localSettings.sidebar_position === "right"
                      ? "bg-bg-hover border-border-focus"
                      : "bg-bg-secondary border-border-subtle hover:bg-bg-hover"
                  )}
                >
                  <LayoutDashboard className="w-6 h-6 text-text-secondary mb-2" style={{ transform: "scaleX(-1)" }} />
                  <p className="text-sm font-medium text-text-primary">Right</p>
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Layout className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Compact Mode</p>
                  <p className="text-xs text-text-muted">Reduce spacing for more content density</p>
                </div>
              </div>
              <Switch
                checked={localSettings.compact_mode}
                onCheckedChange={(v) => handleUpdate('compact_mode', v)}
              />
            </div>
          </div>
        );

      case "editor":
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Save className="w-5 h-5 style={{ color: 'var(--text-accent)' }}" />
                <div>
                  <p className="text-sm font-medium text-text-primary">Auto-Save</p>
                  <p className="text-xs text-text-muted">Automatically save changes</p>
                </div>
              </div>
              <Switch
                checked={localSettings.auto_save}
                onCheckedChange={(v) => handleUpdate('auto_save', v)}
              />
            </div>

            {localSettings.auto_save && (
              <div className="p-4 rounded-lg bg-bg-secondary border border-border-subtle">
                <p className="text-sm font-medium text-text-primary mb-4">Auto-Save Interval</p>
                <Slider
                  value={[localSettings.auto_save_interval]}
                  onValueChange={([v]) => handleUpdate('auto_save_interval', v)}
                  min={10}
                  max={120}
                  step={5}
                  className="w-full"
                />
                <p className="text-xs text-text-muted mt-2">{localSettings.auto_save_interval} seconds</p>
              </div>
            )}

            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Editor Features</h3>
              <div className="space-y-3">
                {[
                  { id: "bracket_matching" as const, label: "Bracket Matching", enabled: localSettings.editor_features.bracket_matching },
                  { id: "minimap" as const, label: "Minimap", enabled: localSettings.editor_features.minimap },
                  { id: "line_numbers" as const, label: "Line Numbers", enabled: localSettings.editor_features.line_numbers },
                  { id: "word_wrap" as const, label: "Word Wrap", enabled: localSettings.editor_features.word_wrap },
                ].map((feature) => (
                  <div
                    key={feature.id}
                    className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-subtle"
                  >
                    <span className="text-sm text-text-secondary">{feature.label}</span>
                    <Switch
                      checked={feature.enabled}
                      onCheckedChange={(v) => handleEditorFeatureUpdate(feature.id, v)}
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>
        );

      case "shortcuts":
        return (
          <div className="space-y-6">
            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Keyboard className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Show Shortcut Hints</p>
                  <p className="text-xs text-text-muted">Display keyboard shortcuts when hovering over actions</p>
                </div>
              </div>
              <Switch
                checked={localSettings.show_shortcut_hints ?? true}
                onCheckedChange={(v) => handleUpdate('show_shortcut_hints', v)}
              />
            </div>
            {(['global', 'navigation', 'actions', 'editor', 'playground'] as const).map((cat) => (
              <div key={cat}>
                <h4 className="text-sm font-semibold text-text-primary mb-3 capitalize">{cat} Shortcuts</h4>
                <div className="space-y-1">
                  {shortcutsByCategory[cat].map((shortcut) => (
                    <div
                      key={shortcut.key}
                      className="flex items-center justify-between px-3 py-2 rounded-md bg-bg-secondary border border-border-subtle"
                    >
                      <span className="text-sm text-text-secondary">{shortcut.description}</span>
                      <kbd className="text-xs font-mono text-text-muted bg-bg-primary border border-border-subtle rounded px-2 py-0.5">
                        {shortcut.displayKey}
                      </kbd>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        );

      case "notifications":
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-semibold text-text-primary mb-4">Notification Level</h3>
              <Select
                value={localSettings.notification_level}
                onValueChange={(v) => handleUpdate('notification_level', v)}
              >
                <SelectTrigger className="w-full bg-bg-secondary border-border-subtle text-text-primary">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Notifications</SelectItem>
                  <SelectItem value="important">Important Only</SelectItem>
                  <SelectItem value="critical">Critical Only</SelectItem>
                  <SelectItem value="none">None</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Volume2 className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Sound Effects</p>
                  <p className="text-xs text-text-muted">Play sounds for notifications</p>
                </div>
              </div>
              <Switch
                checked={localSettings.sound_enabled}
                onCheckedChange={(v) => handleUpdate('sound_enabled', v)}
              />
            </div>
          </div>
        );

      case "privacy":
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Eye className="w-5 h-5 text-status-error" />
                <div>
                  <p className="text-sm font-medium text-text-primary">Usage Analytics</p>
                  <p className="text-xs text-text-muted">Help improve Studio by sharing usage data</p>
                </div>
              </div>
              <Switch
                checked={localSettings.usage_analytics_enabled ?? false}
                onCheckedChange={(v) => handleUpdate('usage_analytics_enabled', v)}
              />
            </div>
            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Shield className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Crash Reports</p>
                  <p className="text-xs text-text-muted">Share crash data to help fix issues</p>
                </div>
              </div>
              <Switch
                checked={localSettings.crash_reports_enabled ?? true}
                onCheckedChange={(v) => handleUpdate('crash_reports_enabled', v)}
              />
            </div>
          </div>
        );

      case "performance":
        return (
          <div className="space-y-4">
            <GlassCard className="p-4 border border-border-subtle">
              <p className="text-sm text-text-secondary">
                Performance settings are optimized for your hardware. These options may require a restart to take effect.
              </p>
            </GlassCard>

            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Monitor className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">GPU Acceleration</p>
                  <p className="text-xs text-text-muted">Use GPU for rendering (recommended)</p>
                </div>
              </div>
              <Switch
                checked={localSettings.gpu_acceleration_enabled ?? true}
                onCheckedChange={(v) => handleUpdate('gpu_acceleration_enabled', v)}
              />
            </div>

            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Sparkles className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Developer Tools</p>
                  <p className="text-xs text-text-muted">Enable Chromium DevTools in Studio</p>
                </div>
              </div>
              <Switch
                checked={localSettings.developer_tools_enabled ?? false}
                onCheckedChange={(v) => handleUpdate('developer_tools_enabled', v)}
              />
            </div>

            <div className="p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <Layout className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                  <div>
                    <p className="text-sm font-medium text-text-primary">Memory Limit</p>
                    <p className="text-xs text-text-muted">Maximum memory for Studio (MB, 0 = unlimited)</p>
                  </div>
                </div>
                <span className="text-sm text-text-muted font-mono">
                  {localSettings.memory_limit_mb ?? 0} MB
                </span>
              </div>
              <Slider
                value={[localSettings.memory_limit_mb ?? 0]}
                onValueChange={([v]) => handleUpdate('memory_limit_mb', v)}
                min={0}
                max={8192}
                step={256}
                className="w-full"
              />
            </div>
          </div>
        );

      case "network":
        return (
          <div className="space-y-4">
            <GlassCard className="p-4 border border-border-subtle">
              <p className="text-sm text-text-secondary">
                Configure a proxy if Studio is behind a corporate firewall or VPN.
              </p>
            </GlassCard>

            <div className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-subtle">
              <div className="flex items-center gap-3">
                <Globe className="w-5 h-5" style={{ color: 'var(--text-accent)' }} />
                <div>
                  <p className="text-sm font-medium text-text-primary">Enable Proxy</p>
                  <p className="text-xs text-text-muted">Route traffic through a proxy server</p>
                </div>
              </div>
              <Switch
                checked={localSettings.proxy_enabled ?? false}
                onCheckedChange={(v) => handleUpdate('proxy_enabled', v)}
              />
            </div>

            {localSettings.proxy_enabled && (
              <>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-text-primary">Proxy URL</label>
                  <Input
                    placeholder="http://proxy.example.com:8080"
                    value={localSettings.proxy_url ?? ''}
                    onChange={(e) => handleUpdate('proxy_url', e.target.value)}
                    className="bg-bg-secondary border-border-subtle text-text-primary"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-text-primary">Bypass List</label>
                  <Input
                    placeholder="localhost,127.0.0.1,.internal"
                    value={localSettings.proxy_bypass ?? ''}
                    onChange={(e) => handleUpdate('proxy_bypass', e.target.value)}
                    className="bg-bg-secondary border-border-subtle text-text-primary"
                  />
                  <p className="text-xs text-text-muted">Comma-separated hosts to skip the proxy</p>
                </div>
              </>
            )}
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full bg-bg-primary">
      <div className="flex items-center justify-between p-5 border-b border-border-subtle">
        <div>
          <h2 className="text-xl font-semibold text-text-primary">Settings Center</h2>
          <p className="text-sm text-text-muted">Configure your studio environment</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
            <Input
              placeholder="Search settings..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-[250px] pl-9 bg-bg-secondary border-border-subtle text-text-primary placeholder:text-text-muted"
            />
          </div>
          <Badge variant="outline" className="text-text-muted border-border-subtle">
            <Sparkles className="w-3 h-3 mr-1" />
            Studio Settings
          </Badge>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="w-64 sm:w-72 lg:w-80 border-r border-border-subtle p-4 overflow-y-auto shrink-0">
          <div className="flex flex-col gap-1">
            {filteredSections.map((section) => (
              <button
                key={section.id}
                onClick={() => setActiveTab(section.id)}
                className={cn(
                  "w-full flex items-center justify-start gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                  activeTab === section.id
                    ? "bg-bg-hover text-text-primary"
                    : "text-text-muted hover:text-text-primary hover:bg-bg-hover"
                )}
              >
                {section.icon}
                <span>{section.label}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {renderContent()}
        </div>
      </div>

      <div className="flex items-center justify-between p-4 border-t border-border-subtle">
        <Button
          variant="outline"
          className="gap-2 border-border-subtle text-text-secondary hover:bg-bg-hover"
          onClick={handleReset}
          disabled={isSaving || isLoading}
        >
          <RotateCcw className="w-4 h-4" />
          Reset to Defaults
        </Button>
        <Button
          className="gap-2 text-white"
          style={{ backgroundColor: 'var(--button-primary)' }}
          onClick={handleSave}
          disabled={isSaving || isLoading}
        >
          <Save className="w-4 h-4" />
          {isSaving ? "Saving..." : "Save Changes"}
        </Button>
      </div>
    </div>
  );
}