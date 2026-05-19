/**
 * @functionfly/ui-ai
 * Intent Translator - Convert natural language to structured intents
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Languages, ArrowRight, Sparkles, CheckCircle2, AlertTriangle } from "lucide-react";

export interface Intent {
  action: string;
  entities: Record<string, string>;
  confidence: number;
  original: string;
}

export interface IntentTranslatorProps {
  input: string;
  parsedIntent?: Intent;
  onIntentConfirm?: (intent: Intent) => void;
  onIntentEdit?: (intent: Intent) => void;
  isProcessing?: boolean;
  className?: string;
}

export function IntentTranslator({
  input,
  parsedIntent,
  onIntentConfirm,
  onIntentEdit,
  isProcessing = false,
  className,
}: IntentTranslatorProps) {
  const [isEditing, setIsEditing] = React.useState(false);
  const [editedIntent, setEditedIntent] = React.useState<Intent | null>(null);

  React.useEffect(() => {
    if (parsedIntent) {
      setEditedIntent(parsedIntent);
      setIsEditing(false);
    }
  }, [parsedIntent]);

  const handleConfirm = () => {
    if (editedIntent) {
      onIntentConfirm?.(editedIntent);
    }
  };

  const handleEdit = () => {
    setEditedIntent(parsedIntent || {
      action: "",
      entities: {},
      confidence: 0,
      original: input,
    });
    setIsEditing(true);
  };

  const handleUpdateAction = (action: string) => {
    setEditedIntent(prev => prev ? { ...prev, action } : null);
  };

  const handleAddEntity = (key: string, value: string) => {
    setEditedIntent(prev => prev ? {
      ...prev,
      entities: { ...prev.entities, [key]: value }
    } : null);
  };

  const handleRemoveEntity = (key: string) => {
    setEditedIntent(prev => prev ? {
      ...prev,
      entities: Object.fromEntries(Object.entries(prev.entities).filter(([k]) => k !== key))
    } : null);
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle">
        <Languages className="size-4 text-brand-500" />
        <span className="text-sm font-medium text-text-primary">Intent Translation</span>
        {parsedIntent && (
          <Badge variant="brand" size="sm">
            {Math.round(parsedIntent.confidence * 100)}% confidence
          </Badge>
        )}
      </div>

      {/* Input Display */}
      <div className="px-4 py-3 border-b border-border-subtle bg-bg-tertiary/30">
        <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
          Original Input
        </label>
        <p className="text-sm text-text-primary mt-1">{input || "Waiting for input..."}</p>
      </div>

      {/* Processing State */}
      {isProcessing && (
        <div className="flex items-center justify-center py-8">
          <div className="flex items-center gap-3 text-brand-500">
            <Sparkles className="size-5 animate-pulse" />
            <span className="text-sm">Analyzing intent...</span>
          </div>
        </div>
      )}

      {/* Parsed Intent Display */}
      {!isProcessing && (parsedIntent || editedIntent) && (
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {editedIntent && (
            <>
              {/* Action */}
              <div className="space-y-2">
                <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                  Action
                </label>
                {isEditing ? (
                  <input
                    type="text"
                    value={editedIntent.action}
                    onChange={e => handleUpdateAction(e.target.value)}
                    className="w-full px-3 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
                    placeholder="e.g., create_workflow, deploy_function"
                  />
                ) : (
                  <div className="flex items-center gap-2">
                    <Badge variant="brand" size="sm">{editedIntent.action}</Badge>
                    <button
                      onClick={() => setIsEditing(true)}
                      className="text-xs text-text-muted hover:text-text-primary"
                    >
                      Edit
                    </button>
                  </div>
                )}
              </div>

              {/* Entities */}
              <div className="space-y-2">
                <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                  Entities
                </label>
                {Object.keys(editedIntent.entities).length === 0 ? (
                  <p className="text-xs text-text-muted italic">No entities extracted</p>
                ) : (
                  <div className="space-y-2">
                    {Object.entries(editedIntent.entities).map(([key, value]) => (
                      <div
                        key={key}
                        className="flex items-center gap-2 p-2 bg-bg-tertiary/50 rounded-lg"
                      >
                        <Badge variant="outline" size="sm" className="shrink-0">{key}</Badge>
                        {isEditing ? (
                          <>
                            <input
                              type="text"
                              value={value}
                              onChange={e => handleAddEntity(key, e.target.value)}
                              className="flex-1 px-2 py-1 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary"
                            />
                            <button
                              onClick={() => handleRemoveEntity(key)}
                              className="p-1 text-text-muted hover:text-error"
                            >
                              ×
                            </button>
                          </>
                        ) : (
                          <span className="text-sm text-text-primary flex-1">{value}</span>
                        )}
                      </div>
                    ))}
                  </div>
                )}

                {/* Add Entity */}
                {isEditing && (
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      placeholder="Entity name"
                      className="w-1/3 px-2 py-1 text-xs bg-bg-primary border border-border-subtle rounded text-text-primary"
                      onKeyDown={e => {
                        if (e.key === "Enter") {
                          const input = e.target as HTMLInputElement;
                          handleAddEntity(input.value, "");
                          input.value = "";
                        }
                      }}
                    />
                  </div>
                )}
              </div>

              {/* Confidence Indicator */}
              <div className="flex items-center gap-3">
                <div className="flex-1 h-2 rounded-full bg-bg-tertiary overflow-hidden">
                  <div
                    className={cn(
                      "h-full rounded-full transition-all",
                      editedIntent.confidence > 0.8 ? "bg-success" :
                      editedIntent.confidence > 0.5 ? "bg-warning" : "bg-error"
                    )}
                    style={{ width: `${editedIntent.confidence * 100}%` }}
                  />
                </div>
                <span className="text-xs text-text-muted">
                  {Math.round(editedIntent.confidence * 100)}%
                </span>
              </div>
            </>
          )}
        </div>
      )}

      {/* Empty State */}
      {!isProcessing && !parsedIntent && !editedIntent && (
        <div className="flex-1 flex flex-col items-center justify-center text-text-muted">
          <Languages className="size-12 mb-3 opacity-50" />
          <p className="text-sm">Enter text to translate to intent</p>
        </div>
      )}

      {/* Actions */}
      {(parsedIntent || editedIntent) && (
        <div className="flex items-center gap-2 px-4 py-3 border-t border-border-subtle">
          {isEditing ? (
            <>
              <button
                onClick={() => setIsEditing(false)}
                className="flex-1 px-3 py-2 text-sm bg-bg-tertiary text-text-secondary rounded-lg hover:bg-bg-hover transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirm}
                className="flex-1 px-3 py-2 text-sm bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
              >
                <CheckCircle2 className="size-4" />
                Confirm Intent
              </button>
            </>
          ) : (
            <>
              <button
                onClick={handleEdit}
                className="flex-1 px-3 py-2 text-sm bg-bg-tertiary text-text-secondary rounded-lg hover:bg-bg-hover transition-colors"
              >
                Edit
              </button>
              <button
                onClick={handleConfirm}
                className="flex-1 px-3 py-2 text-sm bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
              >
                <ArrowRight className="size-4" />
                Execute
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
