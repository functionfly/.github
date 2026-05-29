import { Badge } from '@/components/ui/badge';

type Props = {
  modelUsed?: { provider: string; model_id: string };
  costUsd?: number;
};

export function ModelUsedBadge({ modelUsed, costUsd }: Props) {
  if (!modelUsed) return null;
  const label = `${modelUsed.provider}/${modelUsed.model_id}`;
  return (
    <Badge variant="outline" className="text-xs">
      Generated with {label}
      {typeof costUsd === 'number' ? ` · ~$${costUsd.toFixed(4)}` : ''}
    </Badge>
  );
}
