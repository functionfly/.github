import {
  composerApi,
  type FunctionGenerationRequest,
  type FunctionGenerationResponse,
  type FunctionGenerationResult,
} from '@/api/composer';
import { functionsApi } from '@/api/functions';
import type { ModelSelection } from '@/api/aiModels';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import type { GenerationHistoryItem } from '../useGenerationHistory';
import type { EditableManifest, RefinementChunk, RefinementHistoryItem, StreamChunk } from '../types';

interface UseComposerGenerationOptions {
  description: string;
  runtime: string;
  constraints: string;
  selectedModel?: ModelSelection;
  addToHistory: (item: Omit<GenerationHistoryItem, 'id' | 'timestamp'>) => void;
  onRefinementComplete?: () => void;
}

export function useComposerGeneration({
  description,
  runtime,
  constraints,
  selectedModel,
  addToHistory,
  onRefinementComplete,
}: UseComposerGenerationOptions) {
  const queryClient = useQueryClient();

  const [generatedFunction, setGeneratedFunction] = useState<FunctionGenerationResponse | null>(
    null
  );
  const [streamingResult, setStreamingResult] = useState<
    Partial<FunctionGenerationResult> & { code: string }
  >({ code: '' });
  const [isStreaming, setIsStreaming] = useState(false);
  const [isRefining, setIsRefining] = useState(false);
  const [refinementHistory, setRefinementHistory] = useState<RefinementHistoryItem[]>([]);
  const [confidenceScore, setConfidenceScore] = useState<number | undefined>();
  const [editableManifest, setEditableManifest] = useState<EditableManifest>({
    timeout_seconds: 30,
    memory_mb: 256,
    capabilities: [],
  });

  const eventSourceRef = useRef<EventSource | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const cleanupEventSource = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
  }, []);

  const createFunctionMutation = useMutation({
    mutationFn: async () => {
      if (!streamingResult?.code && !generatedFunction?.result?.code) return null;
      const code = streamingResult?.code || generatedFunction?.result?.code || '';
      const name =
        generatedFunction?.result?.manifest?.name ||
        streamingResult?.manifest?.name ||
        'generated-function';
      return functionsApi.create({
        name,
        code,
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

  const refinementMutation = useMutation({
    mutationFn: async (request: string) => {
      if (!generatedFunction?.generation_id) throw new Error('No generation to refine');
      return composerApi.refineFunction({
        generation_id: generatedFunction.generation_id,
        modification_request: request,
        preserve_structure: true,
      });
    },
    onSuccess: (data) => {
      setGeneratedFunction(data);
      setStreamingResult(data.result || { code: '' });
      setIsRefining(false);
      toast.success('Function refined successfully!');
    },
    onError: (error: Error) => {
      setIsRefining(false);
      toast.error(`Refinement failed: ${error.message}`);
    },
  });

  const generateMutation = useMutation({
    mutationFn: (req: FunctionGenerationRequest) => composerApi.generateFunction(req),
    onSuccess: (data) => {
      setGeneratedFunction(data);
      setStreamingResult(data.result || { code: '' });
      setIsStreaming(false);
      setConfidenceScore(0.85);
      toast.success('Function generated successfully!');
    },
    onError: (error: Error) => {
      setIsStreaming(false);
      toast.error(`Generation failed: ${error.message}`);
    },
  });

  const handleGenerate = useCallback(() => {
    if (!description.trim()) {
      toast.error('Please describe what you want the function to do');
      return;
    }

    cleanupEventSource();
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });
    setIsStreaming(true);
    setRefinementHistory([]);
    setConfidenceScore(undefined);

    const request: FunctionGenerationRequest = {
      description,
      runtime,
      constraints: constraints || undefined,
      provider: selectedModel?.provider,
      model: selectedModel?.model_id,
    };

    const eventSource = composerApi.generateFunctionStream(request);
    eventSourceRef.current = eventSource;
    abortControllerRef.current = new AbortController();

    eventSource.onmessage = (event) => {
      try {
        const chunk: StreamChunk = JSON.parse(event.data);

        switch (chunk.type) {
          case 'chunk':
            setStreamingResult((prev) => ({
              ...prev,
              code: prev.code + (chunk.data || ''),
            }));
            if (chunk.confidence !== undefined) {
              setConfidenceScore(chunk.confidence);
            }
            break;

          case 'done':
            if (chunk.result) {
              const fullResponse: FunctionGenerationResponse = {
                success: true,
                result: chunk.result,
                generation_id: chunk.generation_id || '',
                latency_ms: chunk.latency_ms || 0,
                tokens_used: chunk.tokens_used || { prompt: 0, completion: 0, total: 0 },
              };
              setGeneratedFunction(fullResponse);
              setStreamingResult(chunk.result);
              setConfidenceScore(chunk.confidence || 0.85);
              setEditableManifest({
                timeout_seconds: chunk.result.manifest.timeout_seconds,
                memory_mb: chunk.result.manifest.memory_mb,
                capabilities: chunk.result.manifest.capabilities || [],
              });
              addToHistory({
                description,
                runtime,
                constraints: constraints || undefined,
                result: chunk.result,
                refinementHistory: [],
              });
              toast.success('Function generated successfully!');
            }
            setIsStreaming(false);
            cleanupEventSource();
            break;

          case 'error':
            toast.error(`Generation failed: ${chunk.error || 'Unknown error'}`);
            setIsStreaming(false);
            cleanupEventSource();
            break;

          default:
            console.warn('Unknown stream chunk type:', chunk.type);
        }
      } catch (error) {
        console.error('Failed to parse stream chunk:', error);
      }
    };

    eventSource.onerror = () => {
      if (!generatedFunction && streamingResult.code.length === 0) {
        toast.error('Connection failed. Please try again.');
        setIsStreaming(false);
      }
      cleanupEventSource();
    };

    abortControllerRef.current.signal.addEventListener('abort', () => {
      eventSource.close();
    });
  }, [
    description,
    runtime,
    constraints,
    selectedModel,
    cleanupEventSource,
    generatedFunction,
    streamingResult.code.length,
    addToHistory,
  ]);

  const handleStreamRefine = useCallback(
    (request: string) => {
      if (!generatedFunction?.generation_id) {
        toast.error('No generation to refine');
        return;
      }

      cleanupEventSource();
      setIsStreaming(true);
      setIsRefining(true);
      setStreamingResult((prev) => ({ ...prev, code: '' }));
      setRefinementHistory((prev) => [
        ...prev,
        { id: Date.now().toString(), request, timestamp: new Date() },
      ]);

      const eventSource = composerApi.refineFunctionStream({
        generation_id: generatedFunction.generation_id,
        modification_request: request,
        preserve_structure: true,
        provider: selectedModel?.provider,
        model: selectedModel?.model_id,
      });
      eventSourceRef.current = eventSource;
      abortControllerRef.current = new AbortController();

      eventSource.onmessage = (event) => {
        try {
          const chunk: RefinementChunk = JSON.parse(event.data);

          switch (chunk.type) {
            case 'chunk':
              setStreamingResult((prev) => ({
                ...prev,
                code: prev.code + (chunk.data || ''),
              }));
              break;

            case 'done':
              if (chunk.result) {
                const fullResponse: FunctionGenerationResponse = {
                  success: true,
                  result: chunk.result,
                  generation_id: chunk.refinement_id || chunk.generation_id || '',
                  latency_ms: chunk.latency_ms || 0,
                  tokens_used: chunk.tokens_used || { prompt: 0, completion: 0, total: 0 },
                };
                setGeneratedFunction(fullResponse);
                setStreamingResult(chunk.result);
                toast.success('Function refined successfully!');
                onRefinementComplete?.();
              }
              setIsStreaming(false);
              setIsRefining(false);
              cleanupEventSource();
              break;

            case 'error':
              toast.error(`Refinement failed: ${chunk.error || 'Unknown error'}`);
              setIsStreaming(false);
              setIsRefining(false);
              cleanupEventSource();
              break;

            default:
              console.warn('Unknown refinement chunk type:', chunk.type);
          }
        } catch (error) {
          console.error('Failed to parse refinement chunk:', error);
        }
      };

      eventSource.onerror = () => {
        toast.error('Refinement connection failed. Falling back to non-streaming...');
        refinementMutation.mutate(request);
        cleanupEventSource();
      };

      abortControllerRef.current.signal.addEventListener('abort', () => {
        eventSource.close();
      });
    },
    [generatedFunction?.generation_id, selectedModel, cleanupEventSource, refinementMutation, onRefinementComplete]
  );

  const handleFallbackGenerate = useCallback(() => {
    cleanupEventSource();
    setIsStreaming(true);
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });

    generateMutation.mutate({
      description,
      runtime,
      constraints: constraints || undefined,
      provider: selectedModel?.provider,
      model: selectedModel?.model_id,
    });
  }, [cleanupEventSource, description, runtime, constraints, selectedModel, generateMutation]);

  const handleCancel = useCallback(() => {
    cleanupEventSource();
    setIsStreaming(false);
    setIsRefining(false);
    toast.info('Generation cancelled');
  }, [cleanupEventSource]);

  const handleReset = useCallback(() => {
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });
    setRefinementHistory([]);
    setConfidenceScore(undefined);
    setEditableManifest({
      timeout_seconds: 30,
      memory_mb: 256,
      capabilities: [],
    });
    toast.info('Starting fresh generation');
  }, []);

  const handleCapabilityToggle = useCallback((capability: string) => {
    setEditableManifest((prev) => ({
      ...prev,
      capabilities: prev.capabilities.includes(capability)
        ? prev.capabilities.filter((c) => c !== capability)
        : [...prev.capabilities, capability],
    }));
  }, []);

  useEffect(() => {
    return () => {
      cleanupEventSource();
    };
  }, [cleanupEventSource]);

  const hasResult = Boolean(generatedFunction?.result || streamingResult.code.length > 0);
  const displayResult = generatedFunction?.result || streamingResult;
  const isGenerating = isStreaming || generateMutation.isPending;

  return {
    generatedFunction,
    setGeneratedFunction,
    streamingResult,
    setStreamingResult,
    isStreaming,
    isRefining,
    isGenerating,
    refinementHistory,
    setRefinementHistory,
    confidenceScore,
    editableManifest,
    setEditableManifest,
    hasResult,
    displayResult,
    createFunctionMutation,
    handleGenerate,
    handleStreamRefine,
    handleFallbackGenerate,
    handleCancel,
    handleReset,
    handleCapabilityToggle,
    loadGenerationState: (params: {
      generatedFunction: FunctionGenerationResponse;
      streamingResult: Partial<FunctionGenerationResult> & { code: string };
      refinementHistory: RefinementHistoryItem[];
      editableManifest: EditableManifest;
    }) => {
      setGeneratedFunction(params.generatedFunction);
      setStreamingResult(params.streamingResult);
      setRefinementHistory(params.refinementHistory);
      setEditableManifest(params.editableManifest);
    },
  };
}
