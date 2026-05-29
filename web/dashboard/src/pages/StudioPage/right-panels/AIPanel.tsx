import { AICommandPanel } from '@/components/ai/AICommandPanel';

export function AIPanel() {
  return (
    <div className="h-full min-h-[320px]">
      <AICommandPanel defaultView="chat" />
    </div>
  );
}
