import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DNAHelix, fitnessColor, fitnessLabel, formatNumber } from '../DNAHelix';
import type { DNAProfile } from '@/types/dna';

const mockProfile: DNAProfile = {
  id: '1',
  function_id: 'func-1',
  function_type: 'registry',
  tenant_id: 'tenant-1',
  generation: 3,
  fitness_score: 82,
  total_executions: 15000,
  total_mutations: 2,
  avg_latency_ms: 120.5,
  p99_latency_ms: 350.0,
  success_rate: 0.98,
  error_distribution: { timeout: 5 },
  input_patterns: [
    { shape: 'object', hash: 'abc', frequency: 0.7, count: 70 },
    { shape: 'array', hash: 'def', frequency: 0.3, count: 30 },
  ],
  bottleneck_signature: [
    { type: 'cold_start', severity: 'medium', frequency: 0.15 },
  ],
  dna_hash: 'sha256:abcdef1234567890',
  evolution_enabled: true,
  last_analyzed_at: '2026-05-01T00:00:00Z',
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
};

describe('DNAHelix helpers', () => {
  it('fitnessColor returns correct classes', () => {
    expect(fitnessColor(90)).toBe('text-success');
    expect(fitnessColor(70)).toBe('text-velocity-500');
    expect(fitnessColor(50)).toBe('text-warning');
    expect(fitnessColor(20)).toBe('text-error');
  });

  it('fitnessLabel returns correct labels', () => {
    expect(fitnessLabel(90)).toBe('Excellent');
    expect(fitnessLabel(70)).toBe('Healthy');
    expect(fitnessLabel(50)).toBe('Needs Work');
    expect(fitnessLabel(20)).toBe('Critical');
  });

  it('formatNumber formats correctly', () => {
    expect(formatNumber(500)).toBe('500');
    expect(formatNumber(1500)).toBe('1.5K');
    expect(formatNumber(2500000)).toBe('2.5M');
  });
});

describe('DNAHelix', () => {
  it('renders generation label', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText('Generation 3')).toBeDefined();
  });

  it('renders fitness score', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText(/82\/100/)).toBeDefined();
    expect(screen.getByText(/Healthy/)).toBeDefined();
  });

  it('renders stat cards', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText('15.0K')).toBeDefined();
    expect(screen.getByText('121ms')).toBeDefined();
    expect(screen.getByText('98.0%')).toBeDefined();
    expect(screen.getByText('2')).toBeDefined();
  });

  it('renders DNA hash', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText('sha256:abcdef1234567890')).toBeDefined();
  });

  it('renders bottleneck tags', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText('cold_start')).toBeDefined();
    expect(screen.getByText(/15%/)).toBeDefined();
  });

  it('renders evolution toggle button when callback provided', () => {
    const onToggle = vi.fn();
    render(<DNAHelix profile={mockProfile} onToggleEvolution={onToggle} />);
    expect(screen.getByText('Evolving')).toBeDefined();
  });

  it('renders analyze button when callback provided', () => {
    const onAnalyze = vi.fn();
    render(<DNAHelix profile={mockProfile} onTriggerAnalysis={onAnalyze} />);
    expect(screen.getByText('Analyze Now')).toBeDefined();
  });

  it('shows Paused when evolution is disabled', () => {
    const disabled = { ...mockProfile, evolution_enabled: false };
    render(<DNAHelix profile={disabled} onToggleEvolution={() => {}} />);
    expect(screen.getByText('Paused')).toBeDefined();
  });

  it('renders P99 sublabel', () => {
    render(<DNAHelix profile={mockProfile} />);
    expect(screen.getByText('P99: 350ms')).toBeDefined();
  });
});
