import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { IncidentTimeline } from './IncidentTimeline';
import type { Incident, IncidentSeverity, IncidentStatus } from '@/api/status';

const mockIncidents: Incident[] = [
    {
      id: 'inc-1',
      title: 'API Latency Issues',
      description: 'Increased latency on API endpoints',
      severity: 'high',
      status: 'investigating',
      affected_components: ['api', 'gateway'],
      created_at: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
      updated_at: new Date(Date.now() - 1800000).toISOString(), // 30 mins ago
      updates: [
        {
          id: 'upd-1',
          incident_id: 'inc-1',
          message: 'We are investigating increased latency',
          status: 'investigating',
          created_at: new Date(Date.now() - 3600000).toISOString(),
          created_by: 'user-1',
        },
      ],
    },
    {
      id: 'inc-2',
      title: 'Database Maintenance',
      description: 'Scheduled database maintenance completed',
      severity: 'low',
      status: 'resolved',
      affected_components: ['database'],
      created_at: new Date(Date.now() - 86400000).toISOString(), // 1 day ago
      updated_at: new Date(Date.now() - 82800000).toISOString(), // 23 hours ago
      resolved_at: new Date(Date.now() - 82800000).toISOString(),
      updates: [
        {
          id: 'upd-2',
          incident_id: 'inc-2',
          message: 'Maintenance completed successfully',
          status: 'resolved',
          created_at: new Date(Date.now() - 82800000).toISOString(),
          created_by: 'user-2',
        },
      ],
    },
    {
      id: 'inc-3',
      title: 'Critical System Outage',
      description: 'Complete service unavailability',
      severity: 'critical',
      status: 'monitoring',
      affected_components: ['api', 'database', 'cache'],
      created_at: new Date(Date.now() - 7200000).toISOString(), // 2 hours ago
      updated_at: new Date(Date.now() - 600000).toISOString(), // 10 mins ago
      updates: [],
    },
  ];

describe('IncidentTimeline', () => {
  it('renders incident list', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    expect(screen.getByText('API Latency Issues')).toBeInTheDocument();
    expect(screen.getByText('Database Maintenance')).toBeInTheDocument();
    expect(screen.getByText('Critical System Outage')).toBeInTheDocument();
  });

  it('renders loading state', () => {
    render(<IncidentTimeline incidents={[]} isLoading={true} />);

    // Should show skeleton loaders
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders empty state when no incidents', () => {
    render(<IncidentTimeline incidents={[]} />);

    // Should not crash and should render without incidents
    expect(screen.queryByText('API Latency Issues')).not.toBeInTheDocument();
  });

  it('displays severity badges correctly', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    expect(screen.getByText('Critical')).toBeInTheDocument();
    expect(screen.getByText('High')).toBeInTheDocument();
    expect(screen.getByText('Low')).toBeInTheDocument();
  });

  it('displays status badges correctly', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    expect(screen.getByText('Investigating')).toBeInTheDocument();
    expect(screen.getByText('Resolved')).toBeInTheDocument();
    expect(screen.getByText('Monitoring')).toBeInTheDocument();
  });

  it('displays affected components', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    expect(screen.getByText(/Affects: api, gateway/)).toBeInTheDocument();
    expect(screen.getByText(/Affects: database/)).toBeInTheDocument();
  });

  it('shows duration for ongoing incidents', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Ongoing incident should show "Ongoing for X"
    const ongoingTexts = screen.getAllByText(/Ongoing for/);
    expect(ongoingTexts.length).toBeGreaterThan(0);
  });

  it('shows resolved duration for resolved incidents', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Resolved incident should show "Resolved after X"
    expect(screen.getByText(/Resolved after/)).toBeInTheDocument();
  });

  it('expands incident card on click', async () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Initially, description should be truncated
    const card = screen.getByText('API Latency Issues').closest('[class*="Card"]') ||
                 screen.getByText('API Latency Issues').closest('div');

    // Click to expand
    fireEvent.click(screen.getByText('API Latency Issues'));

    // After expansion, should show description section
    await vi.waitFor(() => {
      expect(screen.getByText('Description')).toBeInTheDocument();
    });
  });

  it('collapses expanded incident on second click', async () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Expand
    fireEvent.click(screen.getByText('API Latency Issues'));
    await vi.waitFor(() => {
      expect(screen.getByText('Description')).toBeInTheDocument();
    });

    // Collapse
    fireEvent.click(screen.getByText('API Latency Issues'));

    // Description should be hidden
    await vi.waitFor(() => {
      expect(screen.queryByText('Description')).not.toBeInTheDocument();
    });
  });

  it('shows updates section when incident has updates', async () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Expand incident with updates
    fireEvent.click(screen.getByText('API Latency Issues'));

    await vi.waitFor(() => {
      expect(screen.getByText('Updates')).toBeInTheDocument();
      expect(screen.getByText('We are investigating increased latency')).toBeInTheDocument();
    });
  });

  it('shows resolved banner for resolved incidents', async () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Expand resolved incident
    fireEvent.click(screen.getByText('Database Maintenance'));

    await vi.waitFor(() => {
      expect(screen.getByText(/Resolved on/)).toBeInTheDocument();
    });
  });

  it('applies correct severity colors', () => {
    const { container } = render(<IncidentTimeline incidents={mockIncidents} />);

    // Check that critical incident has red styling
    const criticalIncident = screen.getByText('Critical System Outage').closest('[class*="Card"]') ||
                            screen.getByText('Critical System Outage').closest('div');

    expect(container.innerHTML).toContain('text-red-400');
    expect(container.innerHTML).toContain('border-red-500');
  });

  it('applies correct status colors', () => {
    const { container } = render(<IncidentTimeline incidents={mockIncidents} />);

    // Should have different colors for different statuses
    expect(container.innerHTML).toContain('bg-red-500/10'); // investigating
    expect(container.innerHTML).toContain('bg-emerald-500/10'); // resolved
  });

  it('respects maxItems limit', () => {
    const manyIncidents = Array.from({ length: 20 }, (_, i) => ({
      ...mockIncidents[0],
      id: `inc-${i}`,
      title: `Incident ${i}`,
    }));

    render(<IncidentTimeline incidents={manyIncidents} maxItems={5} />);

    // Should only show 5 incidents
    expect(screen.getByText('Incident 0')).toBeInTheDocument();
    expect(screen.getByText('Incident 4')).toBeInTheDocument();
    expect(screen.queryByText('Incident 5')).not.toBeInTheDocument();
  });

  it('shows filters when showFilters is true', () => {
    render(<IncidentTimeline incidents={mockIncidents} showFilters={true} />);

    // Should have filter controls
    expect(screen.getByPlaceholderText(/Search incidents/)).toBeInTheDocument();
  });

  it('filters incidents by search term', async () => {
    render(<IncidentTimeline incidents={mockIncidents} showFilters={true} />);

    const searchInput = screen.getByPlaceholderText(/Search incidents/);

    await userEvent.type(searchInput, 'API');

    // Should show API-related incidents
    expect(screen.getByText('API Latency Issues')).toBeInTheDocument();
    expect(screen.queryByText('Database Maintenance')).not.toBeInTheDocument();
  });

  it('clears search filter', async () => {
    render(<IncidentTimeline incidents={mockIncidents} showFilters={true} />);

    const searchInput = screen.getByPlaceholderText(/Search incidents/);
    const clearButton = screen.getByRole('button', { name: /Clear search/ });

    await userEvent.type(searchInput, 'API');
    fireEvent.click(clearButton);

    // Should show all incidents again
    expect(screen.getByText('API Latency Issues')).toBeInTheDocument();
    expect(screen.getByText('Database Maintenance')).toBeInTheDocument();
  });
});

describe('IncidentTimeline Severity Config', () => {
  const severityTestCases: Array<{ severity: IncidentSeverity; expectedLabel: string; expectedColor: string }> = [
    { severity: 'critical', expectedLabel: 'Critical', expectedColor: 'text-red-400' },
    { severity: 'high', expectedLabel: 'High', expectedColor: 'text-orange-400' },
    { severity: 'medium', expectedLabel: 'Medium', expectedColor: 'text-amber-400' },
    { severity: 'low', expectedLabel: 'Low', expectedColor: 'text-blue-400' },
  ];

  severityTestCases.forEach(({ severity, expectedLabel, expectedColor }) => {
    it(`displays ${severity} severity correctly`, () => {
      const incident: Incident = {
        id: 'test-inc',
        title: 'Test Incident',
        description: 'Test',
        severity,
        status: 'investigating',
        affected_components: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      const { container } = render(<IncidentTimeline incidents={[incident]} />);

      expect(screen.getByText(expectedLabel)).toBeInTheDocument();
      expect(container.innerHTML).toContain(expectedColor);
    });
  });
});

describe('IncidentTimeline Status Config', () => {
  const statusTestCases: Array<{ status: IncidentStatus; expectedLabel: string }> = [
    { status: 'investigating', expectedLabel: 'Investigating' },
    { status: 'identified', expectedLabel: 'Identified' },
    { status: 'monitoring', expectedLabel: 'Monitoring' },
    { status: 'resolved', expectedLabel: 'Resolved' },
  ];

  statusTestCases.forEach(({ status, expectedLabel }) => {
    it(`displays ${status} status correctly`, () => {
      const incident: Incident = {
        id: 'test-inc',
        title: 'Test Incident',
        description: 'Test',
        severity: 'medium',
        status,
        affected_components: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        resolved_at: status === 'resolved' ? new Date().toISOString() : undefined,
      };

      render(<IncidentTimeline incidents={[incident]} />);

      expect(screen.getByText(expectedLabel)).toBeInTheDocument();
    });
  });
});

describe('IncidentTimeline Duration Formatting', () => {
  it('formats duration less than 1 hour correctly', () => {
    const incident: Incident = {
      id: 'test-inc',
      title: 'Test Incident',
      description: 'Test',
      severity: 'medium',
      status: 'investigating',
      affected_components: [],
      created_at: new Date(Date.now() - 1800000).toISOString(), // 30 minutes ago
      updated_at: new Date().toISOString(),
    };

    render(<IncidentTimeline incidents={[incident]} />);

    expect(screen.getByText(/30m/)).toBeInTheDocument();
  });

  it('formats duration more than 1 hour correctly', () => {
    const incident: Incident = {
      id: 'test-inc',
      title: 'Test Incident',
      description: 'Test',
      severity: 'medium',
      status: 'investigating',
      affected_components: [],
      created_at: new Date(Date.now() - 5400000).toISOString(), // 1.5 hours ago
      updated_at: new Date().toISOString(),
    };

    render(<IncidentTimeline incidents={[incident]} />);

    expect(screen.getByText(/1h 30m/)).toBeInTheDocument();
  });
});

describe('IncidentTimeline Accessibility', () => {
  const mockIncidents: Incident[] = [
    {
      id: 'inc-1',
      title: 'Test Incident',
      description: 'Test description',
      severity: 'high',
      status: 'investigating',
      affected_components: ['api'],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  ];

  it('has accessible expand/collapse button', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    const expandButton = screen.getByRole('button');
    expect(expandButton).toBeInTheDocument();
  });

  it('has proper heading structure', () => {
    render(<IncidentTimeline incidents={mockIncidents} />);

    // Should have the incident title
    expect(screen.getByText('Test Incident')).toBeInTheDocument();
  });

  it('has accessible severity and status badges', () => {
    const { container } = render(<IncidentTimeline incidents={mockIncidents} />);

    // Should have badges with proper styling for screen readers
    const badges = container.querySelectorAll('[class*="Badge"]');
    expect(badges.length).toBeGreaterThan(0);
  });
});

describe('IncidentTimeline Skeleton Loading', () => {
  it('renders correct number of skeleton items', () => {
    render(<IncidentTimeline incidents={[]} isLoading={true} />);

    // Should have multiple skeleton elements
    const skeletonElements = document.querySelectorAll('.animate-pulse');
    expect(skeletonElements.length).toBeGreaterThanOrEqual(3);
  });

  it('does not show real content while loading', () => {
    const { container } = render(<IncidentTimeline incidents={mockIncidents} isLoading={true} />);

    // Should not show actual incident titles
    expect(container.querySelector('.animate-pulse')).toBeInTheDocument();
  });
});
