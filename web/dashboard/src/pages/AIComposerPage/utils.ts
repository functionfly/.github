import { TOKEN_COST_USD } from './constants';

export function calculateCost(tokens_used?: {
  prompt: number;
  completion: number;
  total: number;
}): number {
  if (!tokens_used) return 0;
  const promptCost = (tokens_used.prompt / 1000) * TOKEN_COST_USD.prompt;
  const completionCost = (tokens_used.completion / 1000) * TOKEN_COST_USD.completion;
  return promptCost + completionCost;
}
