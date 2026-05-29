import type { FunctionGenerationResult } from '@/api/composer';
import {
  Activity,
  Download,
  Edit3,
  FileCode2,
  Maximize2,
  Minimize2,
  Settings2,
  Shield,
  Upload,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Slider } from '@/components/ui/slider';
import { CAPABILITY_INFO } from '../constants';
import type { EditableManifest } from '../types';
import { CapabilityToggle } from './CapabilityToggle';
import { ManifestFlowDiagram } from './ManifestFlowDiagram';

interface ManifestPreviewCardProps {
  manifest: FunctionGenerationResult['manifest'];
  editableManifest: EditableManifest;
  manifestEditMode: boolean;
  manifestExpanded: boolean;
  onEditModeToggle: () => void;
  onExpandedToggle: () => void;
  onManifestChange: (updates: Partial<EditableManifest>) => void;
  onCapabilityToggle: (capability: string) => void;
}

export function ManifestPreviewCard({
  manifest,
  editableManifest,
  manifestEditMode,
  manifestExpanded,
  onEditModeToggle,
  onExpandedToggle,
  onManifestChange,
  onCapabilityToggle,
}: ManifestPreviewCardProps) {
  return (
    <Card className="border-border/50 shadow-sm">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <FileCode2 className="h-5 w-5" />
            Function Manifest
          </CardTitle>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={onEditModeToggle}>
              <Edit3 className="mr-2 h-4 w-4" />
              {manifestEditMode ? 'Done Editing' : 'Edit Manifest'}
            </Button>
            <Button variant="ghost" size="icon" onClick={onExpandedToggle}>
              {manifestExpanded ? (
                <Minimize2 className="h-4 w-4" />
              ) : (
                <Maximize2 className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>
      {manifestExpanded && (
        <CardContent>
          <div className="space-y-6">
            <div className="border rounded-lg bg-muted/30">
              <div className="px-4 py-2 border-b bg-muted/50 flex items-center gap-2">
                <Activity className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">Input/Output Flow</span>
              </div>
              <ManifestFlowDiagram manifest={manifest} />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <h4 className="font-semibold mb-2 flex items-center gap-2">
                  <Download className="h-4 w-4" />
                  Inputs
                </h4>
                {manifest.inputs.length > 0 ? (
                  <ul className="text-sm space-y-2">
                    {manifest.inputs.map((input, i) => (
                      <li key={i} className="bg-muted/50 rounded p-2">
                        <div className="flex items-center gap-2">
                          <code className="font-mono text-xs">{input.name}</code>
                          <Badge variant="outline" className="text-xs">
                            {input.type}
                          </Badge>
                          {input.required && (
                            <Badge variant="secondary" className="text-xs">
                              required
                            </Badge>
                          )}
                        </div>
                        <p className="text-muted-foreground text-xs mt-1">{input.description}</p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground italic">No inputs defined</p>
                )}
              </div>

              <div>
                <h4 className="font-semibold mb-2 flex items-center gap-2">
                  <Upload className="h-4 w-4" />
                  Outputs
                </h4>
                {manifest.outputs.length > 0 ? (
                  <ul className="text-sm space-y-2">
                    {manifest.outputs.map((output, i) => (
                      <li key={i} className="bg-muted/50 rounded p-2">
                        <div className="flex items-center gap-2">
                          <code className="font-mono text-xs">{output.name}</code>
                          <Badge variant="outline" className="text-xs">
                            {output.type}
                          </Badge>
                        </div>
                        <p className="text-muted-foreground text-xs mt-1">{output.description}</p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground italic">No outputs defined</p>
                )}
              </div>

              <div>
                <h4 className="font-semibold mb-2 flex items-center gap-2">
                  <Settings2 className="h-4 w-4" />
                  Configuration
                </h4>

                {manifestEditMode ? (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-xs">
                        <Label className="text-muted-foreground">Timeout</Label>
                        <span className="font-mono">{editableManifest.timeout_seconds}s</span>
                      </div>
                      <Slider
                        value={[editableManifest.timeout_seconds]}
                        onValueChange={([v]) =>
                          onManifestChange({ timeout_seconds: v })
                        }
                        min={1}
                        max={300}
                        step={1}
                      />
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-xs">
                        <Label className="text-muted-foreground">Memory</Label>
                        <span className="font-mono">{editableManifest.memory_mb} MB</span>
                      </div>
                      <Slider
                        value={[editableManifest.memory_mb]}
                        onValueChange={([v]) => onManifestChange({ memory_mb: v })}
                        min={128}
                        max={4096}
                        step={128}
                      />
                    </div>

                    <Separator />

                    <div className="flex justify-between text-sm">
                      <span className="text-muted-foreground">Runtime</span>
                      <span>{manifest.runtime}</span>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Timeout</span>
                      <span className="font-mono">{editableManifest.timeout_seconds}s</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Memory</span>
                      <span className="font-mono">{editableManifest.memory_mb} MB</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Runtime</span>
                      <span>{manifest.runtime}</span>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {manifest.capabilities && (
              <div className="border-t pt-4">
                <h4 className="font-semibold mb-3 flex items-center gap-2">
                  <Shield className="h-4 w-4" />
                  Capabilities
                  {manifestEditMode && (
                    <span className="text-xs font-normal text-muted-foreground">
                      (Click to toggle)
                    </span>
                  )}
                </h4>
                <div className="flex flex-wrap gap-2">
                  {Object.keys(CAPABILITY_INFO).map((cap) => {
                    const isEnabled = editableManifest.capabilities.includes(cap);
                    const isOriginal = manifest.capabilities.includes(cap);

                    if (!manifestEditMode && !isOriginal) return null;

                    return (
                      <CapabilityToggle
                        key={cap}
                        capability={cap}
                        enabled={manifestEditMode ? isEnabled : isOriginal}
                        onToggle={manifestEditMode ? onCapabilityToggle : () => {}}
                      />
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </CardContent>
      )}
    </Card>
  );
}
