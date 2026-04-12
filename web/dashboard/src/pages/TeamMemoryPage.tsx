import '@/pages/TeamsPage/styles.css';
import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { Search, Plus, Filter, Brain, Shield, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useTeamMemories, useSearchMemories, useMemoryExtractions } from '@/hooks/use-team-memory';
import { MemoryCard } from '@/components/team-memory/MemoryCard';
import { CreateMemoryDialog } from '@/components/team-memory/CreateMemoryDialog';
import { ExtractionsPanel } from '@/components/team-memory/ExtractionsPanel';
import { EncryptionSetupDialog } from '@/components/team-memory/EncryptionSetupDialog';
import { type TeamMemory } from '@/services/team-memory.service';

export default function TeamMemoryPage() {
  const { teamId } = useParams<{ teamId: string }>();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEncryptionDialog, setShowEncryptionDialog] = useState(false);

  const { memories, isLoading } = useTeamMemories(teamId || '', {
    memory_type: selectedType === 'all' ? undefined : selectedType,
    limit: 50,
  });

  const searchMutation = useSearchMemories(teamId || '');
  const extractionsQuery = useMemoryExtractions(teamId || '', 'pending');

  const pendingCount = extractionsQuery.data?.length ?? 0;

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      await searchMutation.mutateAsync({ query: searchQuery });
    }
  };

  const typeFilters = [
    { value: 'all', label: 'All Memories' },
    { value: 'decision', label: 'Decisions', icon: CheckCircle },
    { value: 'preference', label: 'Preferences', icon: Brain },
    { value: 'process', label: 'Processes', icon: Filter },
    { value: 'client_context', label: 'Client Context', icon: Brain },
  ];

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Brain className="h-8 w-8 text-primary" />
            Team Memory
          </h1>
          <p className="text-muted-foreground mt-1">
            Shared brain for decisions, preferences, processes, and client context
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => setShowEncryptionDialog(true)}
          >
            <Shield className="h-4 w-4 mr-2" />
            Encryption
          </Button>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Memory
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <form onSubmit={handleSearch} className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search team memories..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Button type="submit" disabled={searchMutation.isPending}>
          {searchMutation.isPending ? 'Searching...' : 'Search'}
        </Button>
      </form>

      {/* Main Content Tabs */}
      <Tabs defaultValue="memories" className="w-full">
        <TabsList className="w-full justify-start">
          <TabsTrigger value="memories">
            Memories
            {memories.length > 0 && (
              <span className="ml-2 text-muted-foreground">
                ({memories.length})
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="extractions">
            Pending Review
            {pendingCount > 0 && (
              <span className="ml-2 bg-primary text-primary-foreground px-2 py-0.5 rounded-full text-xs">
                {pendingCount}
              </span>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="memories" className="space-y-4">
          {/* Type Filter Tabs */}
          <div className="flex gap-2 flex-wrap">
            {typeFilters.map((filter) => (
              <Button
                key={filter.value}
                variant={selectedType === filter.value ? 'default' : 'outline'}
                size="sm"
                onClick={() => setSelectedType(filter.value)}
              >
                {filter.icon && <filter.icon className="h-4 w-4 mr-2" />}
                {filter.label}
              </Button>
            ))}
          </div>

          {/* Memories Grid */}
          {isLoading ? (
            <div className="text-center py-12">Loading memories...</div>
          ) : memories.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Brain className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No memories yet</p>
              <p className="text-sm">
                Start building your team&apos;s shared brain by adding memories or
                completing conversations.
              </p>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {memories.map((memory: TeamMemory) => (
                <MemoryCard key={memory.id} memory={memory} teamId={teamId || ''} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="extractions">
          <ExtractionsPanel teamId={teamId || ''} />
        </TabsContent>
      </Tabs>

      {/* Dialogs */}
      <CreateMemoryDialog
        teamId={teamId || ''}
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
      />

      <EncryptionSetupDialog
        teamId={teamId || ''}
        open={showEncryptionDialog}
        onOpenChange={setShowEncryptionDialog}
      />
    </div>
  );
}
