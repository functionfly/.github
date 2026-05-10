import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DNAVariantDiff } from '../DNAVariantDiff';
import type { DNAMutation } from '@/types/dna';

const mockProposed: DNAMutation = {
  id: 'mut-1',
  function_id: 'func-1',
  function_type: 'registry',
  tenant_id: 'tenant-1',
  generation: 2,
  mutation_type: 'optimize_latency',
  status: 'proposed',
  trigger_reason: 'High latency detected',
  original_code: 'function slow() { sleep(1000); }',
  mutated_code: 'function fast() { /* optimized */ }',
  original_hash: 'sha256:aaa',
  mutated_hash: 'sha256:bbb',
  diff: '@@ -1 +1 @@\n-sleep(1000)\n+/* optimized */',
  estimated_impact: {
    latency_improvement_pct: 45.0,
    memory_reduction_pct: 10.0,
    reliability_improvement_pct: 5.0,
  },
  actual_impact: null,
  confidence: 0.92,
  model_used: 'gpt-4',
  analysis_window_hours: 48,
  executions_analyzed: 50000,
  accepted_by: null,
  accepted_at: null,
  deployed_at: null,
  rolled_back_at: null,
  rejected_reason: null,
  created_at: '2026-05-01T12:00:00Z',
};

const mockAccepted: DNAMutation = {
  ...mockProposed,
  id: 'mut-2',
  status: 'accepted',
  accepted_by: 'user-1',
  accepted_at: '2026-05-01T13:00:00Z',
};

describe('DNAVariantDiff', () => {
  it('renders mutation type label', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('Latency Optimization')).toBeDefined();
  });

  it('renders status badge', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('Proposed')).toBeDefined();
  });

  it('renders generation', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('Gen 2')).toBeDefined();
  });

  it('renders confidence', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('92%')).toBeDefined();
  });

  it('renders impact comparison', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('+45.0%')).toBeDefined();
    expect(screen.getByText('+10.0%')).toBeDefined();
    expect(screen.getByText('+5.0%')).toBeDefined();
  });

  it('renders trigger reason', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('High latency detected')).toBeDefined();
  });

  it('renders model and executions metadata', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('gpt-4', { exact: false })).toBeDefined();
    expect(screen.getByText('50,000', { exact: false })).toBeDefined();
  });

  it('renders diff panel', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    expect(screen.getByText('Unified Diff')).toBeDefined();
  });

  it('renders accept/reject buttons for proposed mutations', () => {
    render(<DNAVariantDiff mutation={mockProposed} onAccept={() => {}} onReject={() => {}} />);
    expect(screen.getByText('Accept & Deploy')).toBeDefined();
    expect(screen.getByText('Reject')).toBeDefined();
  });

  it('does not render accept/reject for non-proposed', () => {
    render(<DNAVariantDiff mutation={mockAccepted} onAccept={() => {}} onReject={() => {}} />);
    expect(screen.queryByText('Accept & Deploy')).toBeNull();
    expect(screen.queryByText('Reject')).toBeNull();
  });

  it('calls onAccept with canary percentage', () => {
    const onAccept = vi.fn();
    render(<DNAVariantDiff mutation={mockProposed} onAccept={onAccept} />);
    fireEvent.click(screen.getByText('Accept & Deploy'));
    expect(onAccept).toHaveBeenCalledWith(10);
  });

  it('shows reject form on reject click', () => {
    render(<DNAVariantDiff mutation={mockProposed} onReject={() => {}} />);
    fireEvent.click(screen.getByText('Reject'));
    expect(screen.getByPlaceholderText(/Reason for rejection/)).toBeDefined();
  });

  it('renders code panels in split view', () => {
    render(<DNAVariantDiff mutation={mockProposed} />);
    fireEvent.click(screen.getByText('Side by Side'));
    expect(screen.getByText('Original')).toBeDefined();
    expect(screen.getByText('Evolved')).toBeDefined();
  });
});
