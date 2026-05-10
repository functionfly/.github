import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EvolutionTimeline } from '../EvolutionTimeline';
import type { DNAMutation } from '@/types/dna';

const mockMutation: DNAMutation = {
  id: 'mut-1',
  function_id: 'func-1',
  function_type: 'registry',
  tenant_id: 'tenant-1',
  generation: 2,
  mutation_type: 'optimize_latency',
  status: 'proposed',
  trigger_reason: 'P99 latency above 500ms',
  original_code: null,
  mutated_code: null,
  original_hash: null,
  mutated_hash: null,
  diff: null,
  estimated_impact: {
    latency_improvement_pct: 40.0,
    memory_reduction_pct: 0,
    reliability_improvement_pct: 5.0,
  },
  actual_impact: null,
  confidence: 0.85,
  model_used: 'gpt-4',
  analysis_window_hours: 48,
  executions_analyzed: 10000,
  accepted_by: null,
  accepted_at: null,
  deployed_at: null,
  rolled_back_at: null,
  rejected_reason: null,
  created_at: '2026-05-01T12:00:00Z',
};

const mockMutations: DNAMutation[] = [
  mockMutation,
  { ...mockMutation, id: 'mut-2', status: 'accepted', generation: 1, mutation_type: 'fix_error_pattern' },
  { ...mockMutation, id: 'mut-3', status: 'rejected', generation: 3, mutation_type: 'reduce_memory', rejected_reason: 'Not needed' },
];

describe('EvolutionTimeline', () => {
  it('renders loading skeletons', () => {
    const { container } = render(<EvolutionTimeline mutations={[]} loading />);
    expect(container.querySelectorAll('.animate-pulse').length).toBe(3);
  });

  it('renders empty state', () => {
    render(<EvolutionTimeline mutations={[]} />);
    expect(screen.getByText('No evolution history yet')).toBeDefined();
  });

  it('renders mutation entries', () => {
    render(<EvolutionTimeline mutations={mockMutations} />);
    expect(screen.getByText('Latency Optimization')).toBeDefined();
    expect(screen.getByText('Error Pattern Fix')).toBeDefined();
    expect(screen.getByText('Memory Reduction')).toBeDefined();
  });

  it('renders generation labels', () => {
    render(<EvolutionTimeline mutations={mockMutations} />);
    const genLabels = screen.getAllByText(/Gen \d/);
    expect(genLabels.length).toBe(3);
  });

  it('renders status badges', () => {
    render(<EvolutionTimeline mutations={mockMutations} />);
    expect(screen.getByText('Proposed')).toBeDefined();
    expect(screen.getByText('Accepted')).toBeDefined();
    expect(screen.getByText('Rejected')).toBeDefined();
  });

  it('renders confidence percentages', () => {
    render(<EvolutionTimeline mutations={mockMutations} />);
    expect(screen.getByText('85%')).toBeDefined();
  });

  it('expands on click and shows trigger reason', () => {
    render(<EvolutionTimeline mutations={[mockMutation]} />);
    fireEvent.click(screen.getByText('Latency Optimization'));
    expect(screen.getByText('P99 latency above 500ms')).toBeDefined();
  });

  it('calls onViewDiff when button clicked', () => {
    const onViewDiff = vi.fn();
    const withDiff = { ...mockMutation, diff: '--- a\n+++ b' };
    render(<EvolutionTimeline mutations={[withDiff]} onViewDiff={onViewDiff} />);

    fireEvent.click(screen.getByText('Latency Optimization'));
    fireEvent.click(screen.getByText(/View Code Diff/));

    expect(onViewDiff).toHaveBeenCalledWith('mut-1');
  });
});
