import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DNAInsightsDashboard } from '../DNAInsightsDashboard';
import type { DNAInsights, EnterpriseInsights } from '@/types/dna';

const mockFunctionInsights: DNAInsights = {
  function_id: 'func-1',
  period: '30d',
  metrics: {
    total_executions: 50000,
    avg_latency_ms: 120,
    p50_latency_ms: 80,
    p95_latency_ms: 250,
    p99_latency_ms: 400,
    success_rate: 0.97,
    error_distribution: { timeout: 30, runtime: 10 },
    cold_start_rate: 0.05,
    avg_memory_peak_mb: 128,
  },
  mutation_outcomes: {
    total: 8,
    outcomes: { proposed: 3, accepted: 2, rejected: 1, deployed: 2, rolled_back: 0 },
  },
};

const mockEnterpriseInsights: EnterpriseInsights = {
  total_functions_analyzed: 42,
  total_mutations_proposed: 15,
  total_mutations_accepted: 8,
  avg_fitness_score: 78.5,
  avg_latency_improvement_pct: 22.3,
  total_cost_savings_usd: 1250.0,
  top_bottleneck_categories: [
    { category: 'timeout', count: 30 },
    { category: 'cold_start', count: 15 },
  ],
  evolution_leaderboard: [
    { function_id: 'func-abc', generation: 5, fitness_score: 95 },
    { function_id: 'func-def', generation: 3, fitness_score: 88 },
  ],
};

describe('DNAInsightsDashboard', () => {
  it('renders loading skeletons', () => {
    const { container } = render(<DNAInsightsDashboard loading />);
    expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0);
  });

  it('renders enterprise heading', () => {
    render(<DNAInsightsDashboard enterpriseInsights={mockEnterpriseInsights} />);
    expect(screen.getByText('Enterprise DNA Insights')).toBeDefined();
  });

  it('renders function heading', () => {
    render(<DNAInsightsDashboard functionInsights={mockFunctionInsights} />);
    expect(screen.getByText('Function DNA Insights')).toBeDefined();
  });

  it('renders enterprise stats', () => {
    render(<DNAInsightsDashboard enterpriseInsights={mockEnterpriseInsights} />);
    expect(screen.getByText('42')).toBeDefined();
    expect(screen.getByText('15')).toBeDefined();
    expect(screen.getByText('22.3%')).toBeDefined();
    expect(screen.getByText('$1250')).toBeDefined();
  });

  it('renders function metrics', () => {
    render(<DNAInsightsDashboard functionInsights={mockFunctionInsights} />);
    expect(screen.getByText('50.0K')).toBeDefined();
    expect(screen.getByText('80ms')).toBeDefined();
    expect(screen.getByText('97.0%')).toBeDefined();
    expect(screen.getByText('5.0%')).toBeDefined();
  });

  it('renders period selector', () => {
    render(<DNAInsightsDashboard onPeriodChange={() => {}} />);
    expect(screen.getByText('7 Days')).toBeDefined();
    expect(screen.getByText('30 Days')).toBeDefined();
    expect(screen.getByText('90 Days')).toBeDefined();
  });

  it('calls onPeriodChange when period selected', () => {
    const onChange = vi.fn();
    render(<DNAInsightsDashboard onPeriodChange={onChange} />);
    fireEvent.click(screen.getByText('7 Days'));
    expect(onChange).toHaveBeenCalledWith('7d');
  });

  it('renders enterprise leaderboard', () => {
    render(<DNAInsightsDashboard enterpriseInsights={mockEnterpriseInsights} />);
    expect(screen.getByText('Evolution Leaderboard')).toBeDefined();
    expect(screen.getByText('func-abc')).toBeDefined();
    expect(screen.getByText('Gen 5')).toBeDefined();
  });

  it('renders bottleneck chart', () => {
    render(<DNAInsightsDashboard enterpriseInsights={mockEnterpriseInsights} />);
    expect(screen.getByText('Top Bottleneck Categories')).toBeDefined();
    expect(screen.getByText('timeout')).toBeDefined();
    expect(screen.getByText('cold start')).toBeDefined();
  });

  it('renders mutation funnel', () => {
    render(<DNAInsightsDashboard functionInsights={mockFunctionInsights} />);
    expect(screen.getByText('Mutation Outcomes')).toBeDefined();
    expect(screen.getByText('Proposed')).toBeDefined();
    expect(screen.getByText('Accepted')).toBeDefined();
    expect(screen.getByText('Deployed')).toBeDefined();
    expect(screen.getByText('Rejected')).toBeDefined();
  });

  it('renders error distribution', () => {
    render(<DNAInsightsDashboard functionInsights={mockFunctionInsights} />);
    expect(screen.getByText('Error Distribution')).toBeDefined();
  });

  it('shows no errors message when empty', () => {
    const noErrors = {
      ...mockFunctionInsights,
      metrics: { ...mockFunctionInsights.metrics, error_distribution: {} },
    };
    render(<DNAInsightsDashboard functionInsights={noErrors} />);
    expect(screen.getByText('No errors detected')).toBeDefined();
  });
});
