/**
 * MCP Center - Registry Tab
 * Main MCP function registry management
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Zap, Plus, Rocket } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { SummaryCards, FunctionTable } from '../components';
import { useMCPFunctions } from '../hooks';
import type { MCPFunctionFilter, MCPFunctionSort } from '../types';

export function RegistryTab() {
  const [filter, setFilter] = useState<MCPFunctionFilter>('all');
  const [sort, setSort] = useState<MCPFunctionSort>('name');
  const [search, setSearch] = useState('');

  const { functions, isLoading, stats } = useMCPFunctions(filter, sort, search);

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <SummaryCards stats={stats} isLoading={isLoading} />

      {/* Registry Table */}
      <Card className="mcp-panel mcp-panel-glow">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="mcp-table-title">
                <Zap className="h-5 w-5" />
                MCP Function Registry
              </CardTitle>
              <CardDescription>
                Manage which functions are available via Model Context Protocol
              </CardDescription>
            </div>
            <Button asChild>
              <Link to="/functions/new">
                <Plus className="h-4 w-4 mr-2" />
                New Function
              </Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <FunctionTable
            functions={functions}
            isLoading={isLoading}
            filter={filter}
            sort={sort}
            search={search}
            onFilterChange={setFilter}
            onSortChange={setSort}
            onSearchChange={setSearch}
          />
        </CardContent>
      </Card>

      {/* Getting Started */}
      {stats.total === 0 && !isLoading && (
        <Card className="mcp-panel">
          <CardContent className="py-12">
            <div className="mcp-empty-state">
              <Rocket className="h-12 w-12" />
              <h3 className="mcp-empty-state-title">No MCP-enabled functions yet</h3>
              <p className="mcp-empty-state-text">
                MCP allows AI agents like Claude Desktop, Cursor, and VS Code to call your
                functions. Enable MCP on any function to make it available to these clients.
              </p>
              <Button asChild>
                <Link to="/functions">Browse Functions</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
