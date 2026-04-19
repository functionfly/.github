import type { Meta, StoryObj } from '@storybook/react';
import { ProviderSearchFilter } from '../components/ProviderSearchFilter';
import { useState } from 'react';

const meta = {
  title: 'Providers/ProviderSearchFilter',
  component: ProviderSearchFilter,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ProviderSearchFilter>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveTemplate = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'connected' | 'available' | 'degraded'>('all');
  const [sortBy, setSortBy] = useState<'name' | 'status' | 'recent' | 'regions'>('name');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  return (
    <ProviderSearchFilter
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      filterStatus={filterStatus}
      onFilterStatusChange={setFilterStatus}
      sortBy={sortBy}
      onSortChange={setSortBy}
      viewMode={viewMode}
      onViewModeChange={setViewMode}
      connectedCount={3}
      availableCount={2}
      degradedCount={1}
      totalCount={5}
    />
  );
};

export const Default: RenderStory = {
  render: InteractiveTemplate,
};

export const WithFilters: RenderStory = {
  render: () => {
    const [searchQuery, setSearchQuery] = useState('cloud');
    const [filterStatus, setFilterStatus] = useState<'all' | 'connected' | 'available' | 'degraded'>('connected');
    const [sortBy, setSortBy] = useState<'name' | 'status' | 'recent' | 'regions'>('status');
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('list');

    return (
      <ProviderSearchFilter
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        filterStatus={filterStatus}
        onFilterStatusChange={setFilterStatus}
        sortBy={sortBy}
        onSortChange={setSortBy}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        connectedCount={3}
        availableCount={2}
        degradedCount={1}
        totalCount={5}
      />
    );
  },
};

export const NoConnected: RenderStory = {
  render: () => {
    const [searchQuery, setSearchQuery] = useState('');
    const [filterStatus, setFilterStatus] = useState<'all' | 'connected' | 'available' | 'degraded'>('all');
    const [sortBy, setSortBy] = useState<'name' | 'status' | 'recent' | 'regions'>('name');
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

    return (
      <ProviderSearchFilter
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        filterStatus={filterStatus}
        onFilterStatusChange={setFilterStatus}
        sortBy={sortBy}
        onSortChange={setSortBy}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        connectedCount={0}
        availableCount={5}
        degradedCount={0}
        totalCount={5}
      />
    );
  },
};

export const AllConnected: RenderStory = {
  render: () => {
    const [searchQuery, setSearchQuery] = useState('');
    const [filterStatus, setFilterStatus] = useState<'all' | 'connected' | 'available' | 'degraded'>('all');
    const [sortBy, setSortBy] = useState<'name' | 'status' | 'recent' | 'regions'>('name');
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

    return (
      <ProviderSearchFilter
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        filterStatus={filterStatus}
        onFilterStatusChange={setFilterStatus}
        sortBy={sortBy}
        onSortChange={setSortBy}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        connectedCount={5}
        availableCount={0}
        degradedCount={0}
        totalCount={5}
      />
    );
  },
};
