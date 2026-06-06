// ReceiptInputOutput — pretty-printed input + output with copy & show-more.
import { Check, Copy } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { truncatedPretty } from "../lib/schema-render";

interface ReceiptInputOutputProps {
  input: unknown;
  output: unknown;
  inputLabel?: string;
  outputLabel?: string;
}

function CopyableBlock({ value, ariaLabel }: { value: unknown; ariaLabel: string }) {
  const { text, truncated } = truncatedPretty(value, 8 * 1024);
  const [copied, setCopied] = useState(false);
  const [expanded, setExpanded] = useState(!truncated);

  const display = expanded ? truncatedPretty(value, 1 << 20).text : text;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(display);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Some browsers block clipboard writes without user gesture; ignore.
    }
  };

  return (
    <div className="relative">
      <pre
        aria-label={ariaLabel}
        className="max-h-96 overflow-auto rounded-md border border-border/40 bg-muted/30 p-3 text-xs font-mono leading-relaxed"
      >
        {display}
      </pre>
      <div className="mt-2 flex items-center justify-between">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={copy}
          className="gap-1.5 text-xs"
          aria-label={`Copy ${ariaLabel} to clipboard`}
        >
          {copied ? <Check className="h-3.5 w-3.5" aria-hidden /> : <Copy className="h-3.5 w-3.5" aria-hidden />}
          {copied ? "Copied" : "Copy"}
        </Button>
        {truncated ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setExpanded(v => !v)}
            className="text-xs"
          >
            {expanded ? "Show less" : "Show more"}
          </Button>
        ) : null}
      </div>
    </div>
  );
}

export function ReceiptInputOutput({ input, output, inputLabel = "Input", outputLabel = "Output" }: ReceiptInputOutputProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Execution</CardTitle>
        <CardDescription>
          The exact JSON that flowed through the function for this run.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="output">
          <TabsList>
            <TabsTrigger value="output">{outputLabel}</TabsTrigger>
            <TabsTrigger value="input">{inputLabel}</TabsTrigger>
          </TabsList>
          <TabsContent value="output" className="pt-3">
            <CopyableBlock value={output} ariaLabel="function output" />
          </TabsContent>
          <TabsContent value="input" className="pt-3">
            <CopyableBlock value={input} ariaLabel="function input" />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}

export default ReceiptInputOutput;
