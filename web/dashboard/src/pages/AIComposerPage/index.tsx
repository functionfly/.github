import { composerApi, type FunctionGenerationRequest, type FunctionGenerationResponse } from '@/api/composer';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { functionsApi } from '@/api/functions';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Sparkles, Code2, Wand2, Loader2, Save, Play, Copy, Check } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

const RUNTIMES = [
  { value: 'python', label: 'Python 3.11', icon: '🐍' },
  { value: 'nodejs', label: 'Node.js 20', icon: '🟢' },
  { value: 'go', label: 'Go 1.21', icon: '🐹' },
  { value: 'rust', label: 'Rust', icon: '🦀' },
  { value: 'deno', label: 'Deno', icon: '🦕' },
  { value: 'bun', label: 'Bun', icon: '🥯' },
];

const COMPLEXITY_COLORS = {
  simple: 'bg-green-500/20 text-green-700 dark:text-green-300',
  moderate: 'bg-yellow-500/20 text-yellow-700 dark:text-yellow-300',
  complex: 'bg-red-500/20 text-red-700 dark:text-red-300',
};

/**
 * AI Composer Page - Natural language to function code generation
 * Uses FlyMind AI to generate complete, production-ready functions
 */
export function AIComposerPage() {
  const queryClient = useQueryClient();
  const [description, setDescription] = useState('');
  const [runtime, setRuntime] = useState('python');
  const [constraints, setConstraints] = useState('');
  const [generatedFunction, setGeneratedFunction] = useState<FunctionGenerationResponse | null>(null);
  const [copied, setCopied] = useState(false);

  const generateMutation = useMutation({
    mutationFn: (req: FunctionGenerationRequest) => composerApi.generateFunction(req),
    onSuccess: (data) => {
      setGeneratedFunction(data);
      toast.success('Function generated successfully!');
    },
    onError: (error: Error) => {
      toast.error(`Generation failed: ${error.message}`);
    },
  });

  const createFunctionMutation = useMutation({
    mutationFn: async () => {
      if (!generatedFunction?.result) return null;
      const result = generatedFunction.result;
      return functionsApi.create({
        name: result.manifest.name,
        code: result.code,
        providers: ['functionfly'],
        region: 'auto',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success('Function saved to your workspace!');
    },
    onError: (error: Error) => {
      toast.error(`Failed to save function: ${error.message}`);
    },
  });

  const handleGenerate = () => {
    if (!description.trim()) {
      toast.error('Please describe what you want the function to do');
      return;
    }

    generateMutation.mutate({
      description,
      runtime,
      constraints: constraints || undefined,
    });
  };

  const handleCopy = () => {
    if (generatedFunction?.result?.code) {
      navigator.clipboard.writeText(generatedFunction.result.code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      toast.success('Code copied to clipboard');
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <div className="p-3 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-lg">
          <Sparkles className="h-8 w-8 text-white" />
        </div>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">AI Composer</h1>
          <p className="text-muted-foreground">
            Describe what you need. FlyMind AI generates production-ready functions.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input Panel */}
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
                onChange={(e) => setDescription(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Be specific about inputs, outputs, and any special requirements.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Runtime</Label>
                <Select value={runtime} onValueChange={setRuntime}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {RUNTIMES.map((r) => (
                      <SelectItem key={r.value} value={r.value}>
                        <span className="mr-2">{r.icon}</span>
                        {r.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="constraints">Constraints (Optional)</Label>
              <Input
                id="constraints"
                placeholder="e.g., Must handle errors gracefully, timeout after 5 seconds, no external dependencies..."
                value={constraints}
                onChange={(e) => setConstraints(e.target.value)}
              />
            </div>

            <Button
              onClick={handleGenerate}
              disabled={generateMutation.isPending || !description.trim()}
              className="w-full bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700"
            >
              {generateMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Generating with FlyMind AI...
                </>
              ) : (
                <>
                  <Sparkles className="mr-2 h-4 w-4" />
                  Generate Function
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        {/* Output Panel */}
        <Card className="border-border/50 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Code2 className="h-5 w-5" />
              Generated Code
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {generateMutation.isPending ? (
              <div className="flex flex-col items-center justify-center py-12 space-y-4">
                <div className="relative">
                  <div className="h-16 w-16 rounded-full border-4 border-violet-200 border-t-violet-500 animate-spin" />
                  <Sparkles className="h-6 w-6 text-violet-500 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2" />
                </div>
                <p className="text-muted-foreground animate-pulse">
                  FlyMind is crafting your function...
                </p>
              </div>
            ) : generatedFunction?.result ? (
              <div className="space-y-4">
                {/* Function Info */}
                <div className="flex items-center justify-between flex-wrap gap-2">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="font-mono">
                      {generatedFunction.result.runtime}
                    </Badge>
                    <Badge
                      className={
                        COMPLEXITY_COLORS[generatedFunction.result.estimated_complexity]
                      }
                    >
                      {generatedFunction.result.estimated_complexity}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" onClick={handleCopy}>
                            {copied ? (
                              <Check className="h-4 w-4 text-green-500" />
                            ) : (
                              <Copy className="h-4 w-4" />
                            )}
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Copy code</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </div>

                {/* Code Display */}
                <ScrollArea className="h-[300px] rounded-md border bg-muted/50">
                  <pre className="p-4 text-sm font-mono whitespace-pre-wrap">
                    {generatedFunction.result.code}
                  </pre>
                </ScrollArea>

                {/* Explanation */}
                <div className="bg-muted/50 rounded-lg p-4">
                  <h4 className="font-semibold mb-2">How it works</h4>
                  <p className="text-sm text-muted-foreground">
                    {generatedFunction.result.explanation}
                  </p>
                </div>

                {/* Suggested Tests */}
                {generatedFunction.result.suggested_tests.length > 0 && (
                  <div>
                    <h4 className="font-semibold mb-2">Suggested Tests</h4>
                    <ul className="text-sm space-y-1">
                      {generatedFunction.result.suggested_tests.map((test, i) => (
                        <li key={i} className="text-muted-foreground flex items-center gap-2">
                          <Play className="h-3 w-3" />
                          {test}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <Separator />

                {/* Actions */}
                <div className="flex gap-2">
                  <Button
                    onClick={() => createFunctionMutation.mutate()}
                    disabled={createFunctionMutation.isPending}
                    className="flex-1"
                  >
                    {createFunctionMutation.isPending ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Saving...
                      </>
                    ) : (
                      <>
                        <Save className="mr-2 h-4 w-4" />
                        Save to My Functions
                      </>
                    )}
                  </Button>
                  <Button variant="outline" onClick={handleCopy}>
                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <Code2 className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <p className="text-muted-foreground">
                  Your generated code will appear here
                </p>
                <p className="text-sm text-muted-foreground/60">
                  Describe what you need and click Generate
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Manifest Preview */}
      {generatedFunction?.result?.manifest && (
        <Card className="border-border/50 shadow-sm">
          <CardHeader>
            <CardTitle>Function Manifest</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div>
                <h4 className="font-semibold mb-2">Inputs</h4>
                {generatedFunction.result.manifest.inputs.length > 0 ? (
                  <ul className="text-sm space-y-2">
                    {generatedFunction.result.manifest.inputs.map((input: any, i: number) => (
                      <li key={i} className="bg-muted/50 rounded p-2">
                        <code className="font-mono text-xs">{input.name}</code>
                        <Badge variant="outline" className="ml-2 text-xs">
                          {input.type}
                        </Badge>
                        {input.required && (
                          <Badge variant="secondary" className="ml-2 text-xs">
                            required
                          </Badge>
                        )}
                        <p className="text-muted-foreground text-xs mt-1">
                          {input.description}
                        </p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">No inputs defined</p>
                )}
              </div>

              <div>
                <h4 className="font-semibold mb-2">Outputs</h4>
                {generatedFunction.result.manifest.outputs.length > 0 ? (
                  <ul className="text-sm space-y-2">
                    {generatedFunction.result.manifest.outputs.map((output: any, i: number) => (
                      <li key={i} className="bg-muted/50 rounded p-2">
                        <code className="font-mono text-xs">{output.name}</code>
                        <Badge variant="outline" className="ml-2 text-xs">
                          {output.type}
                        </Badge>
                        <p className="text-muted-foreground text-xs mt-1">
                          {output.description}
                        </p>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">No outputs defined</p>
                )}
              </div>

              <div>
                <h4 className="font-semibold mb-2">Configuration</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Timeout</span>
                    <span>{generatedFunction.result.manifest.timeout_seconds}s</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Memory</span>
                    <span>{generatedFunction.result.manifest.memory_mb} MB</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Runtime</span>
                    <span>{generatedFunction.result.manifest.runtime}</span>
                  </div>
                  {generatedFunction.result.manifest.capabilities.length > 0 && (
                    <div>
                      <span className="text-muted-foreground">Capabilities</span>
                      <div className="flex flex-wrap gap-1 mt-1">
                        {generatedFunction.result.manifest.capabilities.map((cap: string, i: number) => (
                          <Badge key={i} variant="outline" className="text-xs">
                            {cap}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
