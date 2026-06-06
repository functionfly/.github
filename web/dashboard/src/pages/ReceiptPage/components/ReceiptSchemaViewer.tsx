// ReceiptSchemaViewer — JSON-Schema-aware tree with tabs for input/output.
import { Braces, ChevronRight } from "lucide-react";
import { useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { flattenSchema, prettyJSON } from "../lib/schema-render";
import type { Receipt } from "../types";

interface ReceiptSchemaViewerProps {
  receipt: Receipt;
}

function SchemaTree({ schema }: { schema: unknown }) {
  if (!schema) {
    return (
      <p className="text-sm text-muted-foreground">
        No schema declared. The function accepts and returns arbitrary JSON.
      </p>
    );
  }
  const lines = flattenSchema(schema);
  if (lines.length === 0) {
    // Fall back to pretty JSON for non-standard schemas.
    return (
      <pre className="max-h-64 overflow-auto rounded-md bg-muted/40 p-3 text-xs font-mono">
        {prettyJSON(schema, 2)}
      </pre>
    );
  }
  return (
    <ul className="space-y-1 text-sm" data-testid="schema-tree">
      {lines.map((line, idx) => (
        <li
          key={`${line.key}-${idx}`}
          className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5"
          style={{ paddingLeft: line.depth * 16 }}
        >
          <span className="inline-flex items-center gap-1 font-mono text-xs">
            <ChevronRight className="h-3 w-3 text-muted-foreground/50" aria-hidden />
            <span className="text-foreground">{line.key}</span>
          </span>
          <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground">
            {line.type}
          </span>
          {line.required ? (
            <span className="rounded bg-rose-500/10 px-1.5 py-0.5 font-mono text-[10px] text-rose-400">
              required
            </span>
          ) : null}
          {line.description ? (
            <span className="text-xs text-muted-foreground">— {line.description}</span>
          ) : null}
        </li>
      ))}
    </ul>
  );
}

export function ReceiptSchemaViewer({ receipt }: ReceiptSchemaViewerProps) {
  const hasInput = !!receipt.function.input_schema;
  const hasOutput = !!receipt.function.output_schema;
  const [tab, setTab] = useState<string>(hasInput ? "input" : "output");

  if (!hasInput && !hasOutput) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Braces className="h-4 w-4" aria-hidden /> Schema
          </CardTitle>
          <CardDescription>
            The function author hasn&apos;t published a JSON Schema for this function.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-base">
          <Braces className="h-4 w-4" aria-hidden /> Schema
        </CardTitle>
        <CardDescription>
          Declared input and output types. The function runtime enforces these on the server.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList>
            {hasInput ? <TabsTrigger value="input">Input</TabsTrigger> : null}
            {hasOutput ? <TabsTrigger value="output">Output</TabsTrigger> : null}
          </TabsList>
          {hasInput ? (
            <TabsContent value="input" className="pt-3">
              <SchemaTree schema={receipt.function.input_schema} />
            </TabsContent>
          ) : null}
          {hasOutput ? (
            <TabsContent value="output" className="pt-3">
              <SchemaTree schema={receipt.function.output_schema} />
            </TabsContent>
          ) : null}
        </Tabs>
      </CardContent>
    </Card>
  );
}

export default ReceiptSchemaViewer;
