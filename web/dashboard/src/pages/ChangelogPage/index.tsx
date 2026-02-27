'use client';

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FileText } from 'lucide-react';
import { Footer } from '@/pages/LandingPage/components/Footer';

// Hooks
import { useChangelogData } from './hooks/useChangelogData';
import { useChangelogFilters } from './hooks/useChangelogFilters';
import { useChangelogComparison } from './hooks/useChangelogComparison';
import { useWhatsNew } from './hooks/useWhatsNew';

// Components
import ChangelogHeader from './components/ChangelogHeader';
import LoadingState from './components/LoadingState';
import ErrorState from './components/ErrorState';
import SearchAndFilters from './components/SearchAndFilters';
import AllReleasesTab from './components/tabs/AllReleasesTab';
import CompareTab from './components/tabs/CompareTab';
import WhatsNewTab from './components/tabs/WhatsNewTab';

const ChangelogPage = () => {
  const [activeTab, setActiveTab] = useState('all');
  const [showFilters, setShowFilters] = useState(false);

  // Data fetching hook
  const { changelogEntries, loading, error } = useChangelogData();

  // Filters hook
  const {
    filteredEntries,
    searchTerm,
    setSearchTerm,
    releaseTypeFilter,
    setReleaseTypeFilter,
    categoryFilter,
    setCategoryFilter,
    dateFrom,
    setDateFrom,
    dateTo,
    setDateTo,
    clearFilters,
    hasActiveFilters
  } = useChangelogFilters(changelogEntries);

  // Comparison hook
  const {
    compareVersion1,
    setCompareVersion1,
    compareVersion2,
    setCompareVersion2,
    availableVersions,
    comparisonData,
    differences
  } = useChangelogComparison(changelogEntries);

  // What's new hook
  const {
    whatsNewVersion,
    setWhatsNewVersion,
    whatsNewData
  } = useWhatsNew(changelogEntries);

  if (loading) {
    return <LoadingState />;
  }

  if (error) {
    return <ErrorState error={error} />;
  }

  return (
    <div className="min-h-screen bg-gradient-radial relative overflow-hidden">
      {/* Background Animation Effects */}
      <div className="absolute inset-0 aurora-bg opacity-30"></div>
      <div className="absolute inset-0 gradient-shift-bg opacity-20"></div>

      <ChangelogHeader />

      <div className="container mx-auto px-4 py-8 pt-20 relative z-10">
        <SearchAndFilters
          searchTerm={searchTerm}
          setSearchTerm={setSearchTerm}
          releaseTypeFilter={releaseTypeFilter}
          setReleaseTypeFilter={setReleaseTypeFilter}
          categoryFilter={categoryFilter}
          setCategoryFilter={setCategoryFilter}
          dateFrom={dateFrom}
          setDateFrom={setDateFrom}
          dateTo={dateTo}
          setDateTo={setDateTo}
          showFilters={showFilters}
          setShowFilters={setShowFilters}
          hasActiveFilters={hasActiveFilters}
          clearFilters={clearFilters}
          filteredEntriesCount={filteredEntries.length}
          totalEntriesCount={changelogEntries.length}
        />

        <div className="max-w-6xl mx-auto">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-8">
            <TabsList className="grid w-full grid-cols-3 glass-card glow animate-float">
              <TabsTrigger value="all" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white transition-all duration-300">
                All Releases
              </TabsTrigger>
              <TabsTrigger value="compare" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white transition-all duration-300">
                Compare Versions
              </TabsTrigger>
              <TabsTrigger value="whatsnew" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white transition-all duration-300">
                What's New
              </TabsTrigger>
            </TabsList>

            <TabsContent value="all" className="space-y-8 animate-fade-in">
              <AllReleasesTab filteredEntries={filteredEntries} />
            </TabsContent>

            <TabsContent value="compare" className="space-y-6 animate-fade-in">
              <CompareTab
                compareVersion1={compareVersion1}
                setCompareVersion1={setCompareVersion1}
                compareVersion2={compareVersion2}
                setCompareVersion2={setCompareVersion2}
                availableVersions={availableVersions}
                comparisonData={comparisonData}
                differences={differences}
              />
            </TabsContent>

            <TabsContent value="whatsnew" className="space-y-6 animate-fade-in">
              <WhatsNewTab
                whatsNewVersion={whatsNewVersion}
                setWhatsNewVersion={setWhatsNewVersion}
                availableVersions={availableVersions}
                whatsNewData={whatsNewData}
              />
            </TabsContent>
          </Tabs>

          {/* Back to Home */}
          <div className="text-center mt-12">
            <Link to="/">
              <Button variant="outline" className="btn-secondary hover:glow-sm transition-all duration-300 hover:scale-105">
                <FileText className="h-4 w-4 mr-2 animate-pulse-glow" />
                Back to Home
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Footer */}
      <Footer />
    </div>
  );
};

export default ChangelogPage;