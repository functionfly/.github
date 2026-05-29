import type { ModelSelection } from '@/api/aiModels';
import { ModelPicker } from '@/components/ai/ModelPicker';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { RotateCcw, Sparkles, Wand2 } from 'lucide-react';
import { RUNTIMES } from '../constants';

interface ComposerInputPanelProps {
  description: string;
  runtime: string;
  constraints: string;
  selectedModel: ModelSelection | undefined;
  isGenerating: boolean;
  isRefining: boolean;
  hasResult: boolean;
  onDescriptionChange: (value: string) => void;
  onRuntimeChange: (value: string) => void;
  onConstraintsChange: (value: string) => void;
  onModelChange: (model: ModelSelection | undefined) => void;
  onGenerate: () => void;
  onCancel: () => void;
  onFallbackGenerate: () => void;
  onReset: () => void;
}

export function ComposerInputPanel({
  description,
  runtime,
  constraints,
  selectedModel,
  isGenerating,
  isRefining,
  hasResult,
  onDescriptionChange,
  onRuntimeChange,
  onConstraintsChange,
  onModelChange,
  onGenerate,
  onCancel,
  onFallbackGenerate,
  onReset,
}: ComposerInputPanelProps) {
  return (
    <Card className="border-border/50 shadow-sm">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Wand2 className="h-5 w-5" />
          Describe Your Function
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="description">What should this function do?</Label>
          <Textarea
            id="description"
            placeholder="e.g., A function that takes a URL, fetches the webpage content, extracts all image URLs, and returns them as a JSON array..."
            className="min-h-[150px] resize-none"
            value={description}
            onChange={(e) => onDescriptionChange(e.target.value)}
            disabled={isGenerating}
          />
          <p className="text-xs text-muted-foreground">
            Be specific about inputs, outputs, and any special requirements.
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>Runtime</Label>
            <Select value={runtime} onValueChange={onRuntimeChange} disabled={isGenerating}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {RUNTIMES.map((r) => {
                  const RuntimeIcon = r.icon;
                  return (
                    <SelectItem key={r.value} value={r.value}>
                      <span className="mr-2 inline-flex items-center justify-center">
                        <RuntimeIcon className="w-5 h-5" />
                      </span>
                      {r.label}
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Model</Label>
            <ModelPicker
              feature="composer"
              value={selectedModel}
              onChange={onModelChange}
              capability="code"
              compact
            />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="constraints">Constraints (Optional)</Label>
          <Input
            id="constraints"
            placeholder="e.g., Must handle errors gracefully, timeout after 5 seconds, no external dependencies..."
            value={constraints}
            onChange={(e) => onConstraintsChange(e.target.value)}
            disabled={isGenerating}
          />
        </div>

        {isGenerating ? (
          <div className="flex gap-2">
            <Button variant="outline" onClick={onCancel} className="flex-1">
              Cancel Generation
            </Button>
            {!isRefining && (
              <Button variant="secondary" onClick={onFallbackGenerate} className="flex-1">
                Use Non-Streaming
              </Button>
            )}
          </div>
        ) : (
          <div className="flex gap-2">
            <Button
              onClick={onGenerate}
              disabled={!description.trim()}
              className="flex-1 bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700"
            >
              <Sparkles className="mr-2 h-4 w-4" />
              Generate Function
            </Button>
            {hasResult && (
              <Button variant="outline" onClick={onReset} className="shrink-0">
                <RotateCcw className="h-4 w-4" />
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
