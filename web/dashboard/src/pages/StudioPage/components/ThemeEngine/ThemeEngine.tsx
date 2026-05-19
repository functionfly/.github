import { useState } from "react";
import { GlassCard, Button, Badge } from "@functionfly/ui-core";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Palette, Sun, Moon, Monitor, Sparkles, Check, Eye, Paintbrush,
  Hexagon, Circle, Layers, Zap, Droplet, Flame, Leaf, Variable
} from "lucide-react";

interface ThemePreset {
  id: string;
  name: string;
  description: string;
  colors: {
    primary: string;
    secondary: string;
    accent: string;
    background: string;
    surface: string;
    text: string;
    textSecondary: string;
    border: string;
  };
  icon: React.ReactNode;
}

const themePresets: ThemePreset[] = [
  {
    id: "midnight",
    name: "Midnight Void",
    description: "Deep space darkness with purple accents",
    colors: {
      primary: "#8b5cf6",
      secondary: "#6d28d9",
      accent: "#a78bfa",
      background: "#0a0a0f",
      surface: "#141420",
      text: "#f8fafc",
      textSecondary: "#94a3b8",
      border: "#1e1e30",
    },
    icon: <Moon className="w-5 h-5" />,
  },
  {
    id: "ocean",
    name: "Ocean Depths",
    description: "Deep blue waters with cyan highlights",
    colors: {
      primary: "#06b6d4",
      secondary: "#0891b2",
      accent: "#22d3ee",
      background: "#0c1222",
      surface: "#131929",
      text: "#f1f5f9",
      textSecondary: "#64748b",
      border: "#1e293b",
    },
    icon: <Droplet className="w-5 h-5" />,
  },
  {
    id: "forest",
    name: "Forest Canopy",
    description: "Natural greens with earthy tones",
    colors: {
      primary: "#22c55e",
      secondary: "#16a34a",
      accent: "#4ade80",
      background: "#0a1a0f",
      surface: "#142118",
      text: "#f1f5f9",
      textSecondary: "#64748b",
      border: "#1a2e1c",
    },
    icon: <Leaf className="w-5 h-5" />,
  },
  {
    id: "sunset",
    name: "Sunset Blaze",
    description: "Warm orange to red gradient",
    colors: {
      primary: "#f97316",
      secondary: "#ea580c",
      accent: "#fb923c",
      background: "#1a0a0a",
      surface: "#201414",
      text: "#f1f5f9",
      textSecondary: "#94a3b8",
      border: "#2e1a1a",
    },
    icon: <Flame className="w-5 h-5" />,
  },
  {
    id: "aurora",
    name: "Aurora Borealis",
    description: "Northern lights inspired",
    colors: {
      primary: "#10b981",
      secondary: "#06b6d4",
      accent: "#34d399",
      background: "#0a1018",
      surface: "#131820",
      text: "#f1f5f9",
      textSecondary: "#64748b",
      border: "#1e2830",
    },
    icon: <Sparkles className="w-5 h-5" />,
  },
  {
    id: "neon",
    name: "Neon Cyber",
    description: "High contrast neon aesthetics",
    colors: {
      primary: "#ec4899",
      secondary: "#a855f7",
      accent: "#f472b6",
      background: "#0f0a14",
      surface: "#181020",
      text: "#faf5ff",
      textSecondary: "#a78bfa",
      border: "#241830",
    },
    icon: <Zap className="w-5 h-5" />,
  },
];

const gradientPresets = [
  { id: "none", name: "None", colors: [] },
  { id: "orange-red", name: "Sunset", colors: ["#f97316", "#ef4444"] },
  { id: "purple-blue", name: "Twilight", colors: ["#8b5cf6", "#3b82f6"] },
  { id: "green-teal", name: "Forest", colors: ["#22c55e", "#14b8a6"] },
  { id: "pink-rose", name: "Rose", colors: ["#ec4899", "#f43f5e"] },
  { id: "cyan-blue", name: "Ocean", colors: ["#06b6d4", "#3b82f6"] },
];

export function ThemeEngine() {
  const [activeTab, setActiveTab] = useState("presets");
  const [selectedTheme, setSelectedTheme] = useState("midnight");
  const [selectedGradient, setSelectedGradient] = useState("none");
  const [gradientIntensity, setGradientIntensity] = useState(50);
  const [borderRadius, setBorderRadius] = useState(12);
  const [blurAmount, setBlurAmount] = useState(20);
  const [shadowsEnabled, setShadowsEnabled] = useState(true);
  const [glowEnabled, setGlowEnabled] = useState(true);

  const currentTheme = themePresets.find((t) => t.id === selectedTheme)!;
  const currentGradient = gradientPresets.find((g) => g.id === selectedGradient)!;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div>
          <h2 className="text-xl font-semibold text-white">Theme Engine</h2>
          <p className="text-sm text-white/60">Customize visual appearance</p>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="text-white/60 border-white/20">
            <Palette className="w-3 h-3 mr-1" />
            {currentTheme.name}
          </Badge>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="presets"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Sparkles className="h-4 w-4 shrink-0" />
              Presets
            </TabsTrigger>
            <TabsTrigger
              value="gradients"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Paintbrush className="h-4 w-4 shrink-0" />
              Gradients
            </TabsTrigger>
            <TabsTrigger
              value="effects"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Layers className="h-4 w-4 shrink-0" />
              Effects
            </TabsTrigger>
            <TabsTrigger
              value="preview"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Eye className="h-4 w-4 shrink-0" />
              Preview
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="presets" className="mt-0">
            <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
              {themePresets.map((theme) => (
                <button
                  key={theme.id}
                  onClick={() => setSelectedTheme(theme.id)}
                  className={cn(
                    "relative p-4 rounded-xl border transition-all duration-200 text-left group",
                    selectedTheme === theme.id
                      ? "bg-white/10 border-orange-500/50 shadow-lg shadow-orange-500/10"
                      : "bg-white/5 border-white/10 hover:bg-white/10 hover:border-white/20"
                  )}
                >
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <div className="text-white/80">{theme.icon}</div>
                      <span className="text-sm font-medium text-white">{theme.name}</span>
                    </div>
                    {selectedTheme === theme.id && (
                      <Check className="w-5 h-5 text-orange-400" />
                    )}
                  </div>

                  <div className="flex gap-1.5 mb-2">
                    {Object.entries(theme.colors)
                      .filter(([key]) => ["primary", "secondary", "accent"].includes(key))
                      .map(([, value]) => (
                        <div
                          key={value}
                          className="w-6 h-6 rounded-md shadow-sm"
                          style={{ backgroundColor: value }}
                        />
                      ))}
                  </div>

                  <p className="text-xs text-white/60">{theme.description}</p>

                  {selectedTheme === theme.id && (
                    <div className="absolute inset-0 rounded-xl border-2 border-orange-500/30 pointer-events-none" />
                  )}
                </button>
              ))}
            </div>

            <div className="mt-6">
              <h3 className="text-sm font-semibold text-white mb-3">Theme Colors</h3>
              <div className="grid grid-cols-4 gap-3">
                {Object.entries(currentTheme.colors).map(([key, value]) => (
                  <div key={key} className="space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-white/60 capitalize">{key}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <div
                        className="w-8 h-8 rounded-md border border-white/20"
                        style={{ backgroundColor: value }}
                      />
                      <span className="text-xs text-white/80 font-mono">{value}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="gradients" className="mt-0">
            <div className="space-y-6">
              <div>
                <h3 className="text-sm font-semibold text-white mb-3">Gradient Presets</h3>
                <div className="grid grid-cols-3 gap-3">
                  {gradientPresets.map((gradient) => (
                    <button
                      key={gradient.id}
                      onClick={() => setSelectedGradient(gradient.id)}
                      className={cn(
                        "relative h-20 rounded-xl border transition-all duration-200 overflow-hidden",
                        selectedGradient === gradient.id
                          ? "border-orange-500/50 ring-2 ring-orange-500/20"
                          : "border-white/10 hover:border-white/20"
                      )}
                    >
                      {gradient.colors.length > 0 ? (
                        <div
                          className="absolute inset-0"
                          style={{
                            background: `linear-gradient(135deg, ${gradient.colors[0]}, ${gradient.colors[1]})`,
                          }}
                        />
                      ) : (
                        <div className="absolute inset-0 bg-white/5 flex items-center justify-center">
                          <span className="text-xs text-white/40">None</span>
                        </div>
                      )}
                      <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 to-transparent p-2">
                        <span className="text-xs text-white font-medium">{gradient.name}</span>
                      </div>
                    </button>
                  ))}
                </div>
              </div>

              {selectedGradient !== "none" && (
                <div className="space-y-4 p-4 rounded-xl bg-white/5 border border-white/10">
                  <h3 className="text-sm font-semibold text-white">Gradient Intensity</h3>
                  <Slider
                    value={[gradientIntensity]}
                    onValueChange={([v]) => setGradientIntensity(v)}
                    min={0}
                    max={100}
                    step={5}
                    className="w-full"
                  />
                  <div className="flex justify-between text-xs text-white/60">
                    <span>Subtle</span>
                    <span>{gradientIntensity}%</span>
                    <span>Intense</span>
                  </div>
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="effects" className="mt-0">
            <div className="space-y-6">
              <div className="space-y-4 p-4 rounded-xl bg-white/5 border border-white/10">
                <h3 className="text-sm font-semibold text-white">Border Radius</h3>
                <Slider
                  value={[borderRadius]}
                  onValueChange={([v]) => setBorderRadius(v)}
                  min={0}
                  max={24}
                  step={1}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-white/60">
                  <span>Sharp</span>
                  <span>{borderRadius}px</span>
                  <span>Rounded</span>
                </div>
              </div>

              <div className="space-y-4 p-4 rounded-xl bg-white/5 border border-white/10">
                <h3 className="text-sm font-semibold text-white">Backdrop Blur</h3>
                <Slider
                  value={[blurAmount]}
                  onValueChange={([v]) => setBlurAmount(v)}
                  min={0}
                  max={40}
                  step={2}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-white/60">
                  <span>None</span>
                  <span>{blurAmount}px</span>
                  <span>Heavy</span>
                </div>
              </div>

              <div className="space-y-3">
                <h3 className="text-sm font-semibold text-white">Toggles</h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-4 rounded-lg bg-white/5 border border-white/10">
                    <div className="flex items-center gap-3">
                      <Hexagon className="w-5 h-5 text-purple-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Shadows</p>
                        <p className="text-xs text-white/60">Depth and layering effects</p>
                      </div>
                    </div>
                    <Switch checked={shadowsEnabled} onCheckedChange={setShadowsEnabled} />
                  </div>
                  <div className="flex items-center justify-between p-4 rounded-lg bg-white/5 border border-white/10">
                    <div className="flex items-center gap-3">
                      <Sparkles className="w-5 h-5 text-orange-400" />
                      <div>
                        <p className="text-sm font-medium text-white">Glow Effects</p>
                        <p className="text-xs text-white/60">Neon and ambient lighting</p>
                      </div>
                    </div>
                    <Switch checked={glowEnabled} onCheckedChange={setGlowEnabled} />
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="preview" className="mt-0">
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-white">Live Preview</h3>
              <div
                className="p-6 rounded-xl border transition-all duration-300"
                style={{
                  backgroundColor: currentTheme.colors.surface,
                  borderColor: currentTheme.colors.border,
                  borderRadius: `${borderRadius}px`,
                  backdropFilter: `blur(${blurAmount}px)`,
                }}
              >
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center"
                      style={{ backgroundColor: currentTheme.colors.primary }}
                    >
                      <Palette className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <h4 className="font-medium" style={{ color: currentTheme.colors.text }}>
                        Sample Card
                      </h4>
                      <p className="text-sm" style={{ color: currentTheme.colors.textSecondary }}>
                        This is how your theme looks
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      className="px-4 py-2 rounded-lg text-sm font-medium text-white"
                      style={{ backgroundColor: currentTheme.colors.primary }}
                    >
                      Primary
                    </button>
                    <button
                      className="px-4 py-2 rounded-lg text-sm font-medium border"
                      style={{
                        borderColor: currentTheme.colors.border,
                        color: currentTheme.colors.textSecondary,
                      }}
                    >
                      Secondary
                    </button>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-xl bg-white/5 border border-white/10">
                <p className="text-xs text-white/60 mb-2">CSS Variables Generated</p>
                <pre className="text-xs text-white/80 font-mono overflow-x-auto">
                  {`--theme-primary: ${currentTheme.colors.primary};
--theme-secondary: ${currentTheme.colors.secondary};
--theme-accent: ${currentTheme.colors.accent};
--theme-background: ${currentTheme.colors.background};
--theme-surface: ${currentTheme.colors.surface};
--theme-text: ${currentTheme.colors.text};
--theme-border-radius: ${borderRadius}px;
--theme-blur: ${blurAmount}px;`}
                </pre>
              </div>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}