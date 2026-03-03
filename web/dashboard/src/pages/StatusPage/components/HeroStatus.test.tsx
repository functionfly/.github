import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { HeroStatus } from './HeroStatus';
import type { PlatformStatus } from '@/api/status';

describe('HeroStatus', () => {
  const mockRefresh = vi.fn();

  const baseStatus: PlatformStatus = {
    status: 'operational',
    message: 'All systems operational',
    timestamp: new Date().toISOString(),
    components: [
      { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
      { id: 'database', name: 'Database', category: 'infrastructure', status: 'operational', latency_ms: 12, uptime_percent: 99.99, last_checked: new Date().toISOString() },
    ],
  };

  it('renders operational status correctly', () => {
    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText('All Systems Operational')).toBeInTheDocument();
    expect(screen.getByText('All systems operational')).toBeInTheDocument();
  });

  it('renders degraded status correctly', () => {
    const degradedStatus: PlatformStatus = {
      ...baseStatus,
      status: 'degraded',
      message: 'Some services are experiencing issues',
    };

    render(
      <HeroStatus
        status={degradedStatus}
        lastUpdated={degradedStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText('Degraded Performance')).toBeInTheDocument();
    expect(screen.getByText('Some services are experiencing issues')).toBeInTheDocument();
  });

  it('renders major outage status correctly', () => {
    const outageStatus: PlatformStatus = {
      ...baseStatus,
      status: 'major_outage',
      message: 'Multiple services are currently unavailable',
    };

    render(
      <HeroStatus
        status={outageStatus}
        lastUpdated={outageStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText('Major Service Outage')).toBeInTheDocument();
    expect(screen.getByText('Multiple services are currently unavailable')).toBeInTheDocument();
  });

  it('renders with null status using defaults', () => {
    render(
      <HeroStatus
        status={null}
        lastUpdated={null}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    // Should default to operational
    expect(screen.getByText('All Systems Operational')).toBeInTheDocument();
  });

  it('displays loading state', () => {
    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={true}
        onRefresh={mockRefresh}
      />
    );

    // Refresh button should be disabled during loading
    const refreshButton = screen.getByLabelText('Refresh status');
    expect(refreshButton).toBeDisabled();
  });

  it('calls onRefresh when refresh button is clicked', () => {
    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    const refreshButton = screen.getByLabelText('Refresh status');
    fireEvent.click(refreshButton);

    expect(mockRefresh).toHaveBeenCalledTimes(1);
  });

  it('displays last updated timestamp', () => {
    const timestamp = new Date().toISOString();

    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText(/Last updated:/)).toBeInTheDocument();
  });

  it('displays "Never" when lastUpdated is null', () => {
    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={null}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText(/Last updated: Never/)).toBeInTheDocument();
  });

  it('displays component summary with counts', () => {
    const statusWithComponents: PlatformStatus = {
      ...baseStatus,
      components: [
        { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
        { id: 'db', name: 'Database', category: 'infrastructure', status: 'operational', latency_ms: 12, uptime_percent: 99.99, last_checked: new Date().toISOString() },
        { id: 'cache', name: 'Cache', category: 'infrastructure', status: 'degraded', latency_ms: 120, uptime_percent: 95.0, last_checked: new Date().toISOString() },
      ],
    };

    render(
      <HeroStatus
        status={statusWithComponents}
        lastUpdated={statusWithComponents.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText(/2 Operational/)).toBeInTheDocument();
    expect(screen.getByText(/1 Degraded/)).toBeInTheDocument();
  });

  it('displays down count when components are down', () => {
    const statusWithDown: PlatformStatus = {
      ...baseStatus,
      components: [
        { id: 'api', name: 'API', category: 'core', status: 'operational', latency_ms: 45, uptime_percent: 99.9, last_checked: new Date().toISOString() },
        { id: 'db', name: 'Database', category: 'infrastructure', status: 'down', latency_ms: 0, uptime_percent: 0, last_checked: new Date().toISOString() },
      ],
    };

    render(
      <HeroStatus
        status={statusWithDown}
        lastUpdated={statusWithDown.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByText(/1 Operational/)).toBeInTheDocument();
    expect(screen.getByText(/1 Down/)).toBeInTheDocument();
  });

  it('does not display component summary when no components', () => {
    const statusWithoutComponents: PlatformStatus = {
      ...baseStatus,
      components: [],
    };

    render(
      <HeroStatus
        status={statusWithoutComponents}
        lastUpdated={statusWithoutComponents.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.queryByText(/Operational/)).not.toBeInTheDocument();
  });

  it('has correct aria-label', () => {
    render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByLabelText('Platform Status')).toBeInTheDocument();
  });

  it('applies correct gradient for operational status', () => {
    const { container } = render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    const section = container.querySelector('section');
    expect(section).toHaveClass('from-emerald-500/20');
  });

  it('applies correct gradient for degraded status', () => {
    const degradedStatus: PlatformStatus = {
      ...baseStatus,
      status: 'degraded',
    };

    const { container } = render(
      <HeroStatus
        status={degradedStatus}
        lastUpdated={degradedStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    const section = container.querySelector('section');
    expect(section).toHaveClass('from-amber-500/20');
  });

  it('applies correct gradient for major outage status', () => {
    const outageStatus: PlatformStatus = {
      ...baseStatus,
      status: 'major_outage',
    };

    const { container } = render(
      <HeroStatus
        status={outageStatus}
        lastUpdated={outageStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    const section = container.querySelector('section');
    expect(section).toHaveClass('from-red-500/20');
  });

  it('shows refresh button only when onRefresh is provided', () => {
    const { rerender } = render(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByLabelText('Refresh status')).toBeInTheDocument();

    rerender(
      <HeroStatus
        status={baseStatus}
        lastUpdated={baseStatus.timestamp}
        isLoading={false}
      />
    );

    expect(screen.queryByLabelText('Refresh status')).not.toBeInTheDocument();
  });
});

describe('HeroStatus Pulse Animation', () => {
  it('shows pulse animation for operational status', () => {
    const status: PlatformStatus = {
      status: 'operational',
      message: 'All systems operational',
      timestamp: new Date().toISOString(),
      components: [],
    };

    const { container } = render(
      <HeroStatus
        status={status}
        lastUpdated={status.timestamp}
        isLoading={false}
      />
    );

    // Should have animated background pulse
    const pulse = container.querySelector('.animate-pulse, [class*="animate-"]');
    expect(pulse).toBeInTheDocument();
  });

  it('shows pulse ring for non-operational statuses', () => {
    const degradedStatus: PlatformStatus = {
      status: 'degraded',
      message: 'Issues detected',
      timestamp: new Date().toISOString(),
      components: [],
    };

    const { container } = render(
      <HeroStatus
        status={degradedStatus}
        lastUpdated={degradedStatus.timestamp}
        isLoading={false}
      />
    );

    // Should have pulse ring animation
    const pulseRing = container.querySelector('[class*="animate-"]');
    expect(pulseRing).toBeInTheDocument();
  });
});

describe('HeroStatus Timestamp Formatting', () => {
  it('formats time correctly', () => {
    const now = new Date();
    const timestamp = now.toISOString();

    const status: PlatformStatus = {
      status: 'operational',
      message: 'All systems operational',
      timestamp,
      components: [],
    };

    render(
      <HeroStatus
        status={status}
        lastUpdated={timestamp}
        isLoading={false}
      />
    );

    // Should display formatted time (e.g., "Last updated: 2:30:45 PM")
    expect(screen.getByText(/Last updated:/)).toBeInTheDocument();
  });
});
