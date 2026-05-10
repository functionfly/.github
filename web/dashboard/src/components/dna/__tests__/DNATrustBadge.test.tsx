import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DNATrustBadge } from '../DNATrustBadge';

describe('DNATrustBadge', () => {
  it('renders micro variant', () => {
    render(
      <DNATrustBadge
        generation={3}
        fitnessScore={85}
        totalMutations={2}
        variant="micro"
      />
    );
    expect(screen.getByText('Gen 3')).toBeDefined();
  });

  it('renders compact variant with fitness dot', () => {
    render(
      <DNATrustBadge
        generation={5}
        fitnessScore={92}
        totalMutations={4}
        variant="compact"
      />
    );
    expect(screen.getByText('Gen 5')).toBeDefined();
    expect(screen.getByText('92')).toBeDefined();
    expect(screen.getByText('4 evolutions')).toBeDefined();
  });

  it('renders compact variant with singular evolution', () => {
    render(
      <DNATrustBadge
        generation={2}
        fitnessScore={70}
        totalMutations={1}
        variant="compact"
      />
    );
    expect(screen.getByText('1 evolution')).toBeDefined();
  });

  it('renders full variant with all stats', () => {
    render(
      <DNATrustBadge
        generation={10}
        fitnessScore={95}
        totalMutations={8}
        totalExecutions={50000}
        variant="full"
      />
    );
    expect(screen.getByText('Function DNA')).toBeDefined();
    expect(screen.getByText('10')).toBeDefined();
    expect(screen.getByText('95')).toBeDefined();
    expect(screen.getByText('8')).toBeDefined();
    expect(screen.getByText('Excellent')).toBeDefined();
  });

  it('shows improved since v1 when generation > 1', () => {
    render(
      <DNATrustBadge
        generation={3}
        fitnessScore={80}
        totalMutations={2}
        variant="full"
      />
    );
    expect(screen.getByText(/improved since v1/)).toBeDefined();
  });

  it('does not show improved since v1 when generation <= 1', () => {
    const { container } = render(
      <DNATrustBadge
        generation={1}
        fitnessScore={80}
        totalMutations={0}
        variant="full"
      />
    );
    expect(container.textContent).not.toContain('improved since v1');
  });

  it('applies custom className', () => {
    const { container } = render(
      <DNATrustBadge
        generation={1}
        fitnessScore={80}
        totalMutations={0}
        variant="micro"
        className="custom-class"
      />
    );
    expect(container.firstChild).toBeDefined();
  });
});
