import { SectionWrapper } from "./SectionWrapper";
import { ArchitectureDiagram } from "./ArchitectureDiagram";
import { CodeSnippetBlock } from "./CodeSnippetBlock";
import { FadeInOnScroll } from "./FadeInOnScroll";

const codeExample = `// StateFabric Event Logging
await stateFabric.log({
  eventType: "function_called",
  functionId: "agent-123",
  payload: { action: "process_data" },
  timestamp: Date.now()
});

// Deterministic Replay
const replay = await stateFabric.replay({
  fromTimestamp: startTime,
  toTimestamp: endTime,
  functionId: "agent-123"
});`;

export function ArchitectureVisualizationSection() {
  return (
    <SectionWrapper className="bg-bg-primary">
      <div className="max-w-6xl mx-auto">
        <FadeInOnScroll>
          <h2 className="text-3xl lg:text-4xl font-bold text-center text-slate-900 dark:text-white mb-4">
            How StateFabric Works
          </h2>
        </FadeInOnScroll>

        <FadeInOnScroll delay={0.2}>
          <p className="text-xl text-slate-700 dark:text-slate-300 text-center mb-12">
            A complete architecture for durable, replayable state management.
          </p>
        </FadeInOnScroll>

        <ArchitectureDiagram />

        <FadeInOnScroll delay={0.9}>
          <div className="mt-12 max-w-2xl mx-auto">
            <CodeSnippetBlock
              title="StateFabric API Example"
              code={codeExample}
              language="typescript"
            />
          </div>
        </FadeInOnScroll>
      </div>
    </SectionWrapper>
  );
}