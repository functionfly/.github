// ReceiptStats component unit tests.
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ReceiptStats } from './ReceiptStats';

describe('ReceiptStats', () => {
  it('renders all 4 stat cards', () => {
    render(
      <ReceiptStats
        durationMs={142}
        cached={false}
        createdAt="2026-06-01T18:42:11Z"
        viewCount={42}
      />
    );
    expect(screen.getByText('142ms')).toBeVisible();
    expect(screen.getByText('Live')).toBeVisible();
    expect(screen.getByText('Execution time')).toBeVisible();
    expect(screen.getByText('Execution type')).toBeVisible();
    expect(screen.getByText('Public views')).toBeVisible();
  });

  it('shows Cached instead of Live when cached=true', () => {
    render(
      <ReceiptStats durationMs={5} cached={true} createdAt="2026-06-01T18:42:11Z" viewCount={0} />
    );
    expect(screen.getByText('Cached')).toBeVisible();
    expect(screen.queryByText('Live')).not.toBeInTheDocument();
  });

  it('formats view count with locale', () => {
    render(
      <ReceiptStats
        durationMs={10}
        cached={false}
        createdAt="2026-06-01T18:42:11Z"
        viewCount={1234}
      />
    );
    expect(screen.getByText('1,234')).toBeVisible();
  });

  it('defaults viewCount to 0', () => {
    render(<ReceiptStats durationMs={10} cached={false} createdAt="2026-06-01T18:42:11Z" />);
    expect(screen.getByText('0')).toBeVisible();
  });

  it('has correct data-testid', () => {
    render(<ReceiptStats durationMs={1} cached={false} createdAt="2026-06-01T18:42:11Z" />);
    expect(screen.getByTestId('receipt-stats')).toBeInTheDocument();
  });
});
