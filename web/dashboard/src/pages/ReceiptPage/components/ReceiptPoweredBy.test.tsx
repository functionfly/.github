// ReceiptPoweredBy component unit tests.
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ReceiptPoweredBy } from './ReceiptPoweredBy';

describe('ReceiptPoweredBy', () => {
  it('renders as a link with correct href', () => {
    render(<ReceiptPoweredBy />);
    const link = screen.getByTestId('receipt-powered-by');
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', expect.stringContaining('functionfly.com'));
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', expect.stringContaining('sponsored'));
  });

  it('has accessible aria-label', () => {
    render(<ReceiptPoweredBy />);
    expect(screen.getByLabelText(/powered by functionfly/i)).toBeInTheDocument();
  });
});
