'use client';

import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Navbar } from '@/components/common/Navbar';
import { FileText, BookOpen, Users, FolderOpen, RefreshCw, Plus, Settings, ArrowLeft, ChevronDown } from 'lucide-react';
import { contentAdminApi } from '@/api/content';
import ChangelogManager from './components/ChangelogManager';
import BlogManager from './components/BlogManager';
import AuthorManager from './components/AuthorManager';
import CategoryManager from './components/CategoryManager';
import ErrorBoundary from './components/ErrorBoundary';

const AdminContentPage = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('changelog');
  const [syncLoading, setSyncLoading] = useState<{
    github: boolean;
  }>({
    github: false,
  });

  // Refs to access manager component methods
  const changelogManagerRef = useRef<{ openCreateDialog: () => void } | null>(null);
  const blogManagerRef = useRef<{ openCreateDialog: () => void } | null>(null);
  const authorManagerRef = useRef<{ openCreateDialog: () => void } | null>(null);
  const categoryManagerRef = useRef<{ openCreateDialog: () => void } | null>(null);

  const handleSyncGitHub = async () => {
    try {
      setSyncLoading(prev => ({ ...prev, github: true }));
      await contentAdminApi.syncGitHubReleases();
      // Refresh data or show success message
      window.location.reload();
    } catch (error) {
      console.error('Failed to sync GitHub releases:', error);
    } finally {
      setSyncLoading(prev => ({ ...prev, github: false }));
    }
  };

  const handleNewContent = (contentType: string) => {
    setActiveTab(contentType);
    // Use setTimeout to ensure tab change has completed before opening dialog
    setTimeout(() => {
      switch (contentType) {
        case 'changelog':
          changelogManagerRef.current?.openCreateDialog?.();
          break;
        case 'blog':
          blogManagerRef.current?.openCreateDialog?.();
          break;
        case 'authors':
          authorManagerRef.current?.openCreateDialog?.();
          break;
        case 'categories':
          categoryManagerRef.current?.openCreateDialog?.();
          break;
      }
    }, 0);
  };


  return (
    <div className="min-h-screen bg-background">
      {/* Navbar */}
      <Navbar variant="dashboard" />

      {/* Header */}
      <div className="border-b pt-16">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-between">
            <Button
              variant="ghost"
              onClick={() => navigate('/admin')}
              className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Dashboard
            </Button>
            <div className="flex-1 text-center flex items-center justify-center gap-3">
              <Settings className="h-8 w-8 text-primary" />
              <div>
                <h1 className="text-3xl font-bold text-text-primary">Content Management</h1>
                <p className="text-text-secondary">
                  Manage changelog entries, blog posts, authors, and categories
                </p>
              </div>
            </div>
            <div className="flex gap-2">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    New Content
                    <ChevronDown className="h-4 w-4 ml-2" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => handleNewContent('changelog')}>
                    <FileText className="h-4 w-4 mr-2" />
                    New Changelog Entry
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleNewContent('blog')}>
                    <BookOpen className="h-4 w-4 mr-2" />
                    New Blog Post
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleNewContent('authors')}>
                    <Users className="h-4 w-4 mr-2" />
                    New Author
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleNewContent('categories')}>
                    <FolderOpen className="h-4 w-4 mr-2" />
                    New Category
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <Button
                onClick={handleSyncGitHub}
                disabled={syncLoading.github}
                variant="outline"
              >
                <RefreshCw className={`h-4 w-4 mr-2 ${syncLoading.github ? 'animate-spin' : ''}`} />
                Sync GitHub
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="changelog" className="flex items-center gap-2">
              <FileText className="h-4 w-4" />
              Changelog
            </TabsTrigger>
            <TabsTrigger value="blog" className="flex items-center gap-2">
              <BookOpen className="h-4 w-4" />
              Blog Posts
            </TabsTrigger>
            <TabsTrigger value="authors" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              Authors
            </TabsTrigger>
            <TabsTrigger value="categories" className="flex items-center gap-2">
              <FolderOpen className="h-4 w-4" />
              Categories
            </TabsTrigger>
          </TabsList>

          <TabsContent value="changelog" className="space-y-6">
            <ErrorBoundary>
              <ChangelogManager ref={changelogManagerRef} />
            </ErrorBoundary>
          </TabsContent>

          <TabsContent value="blog" className="space-y-6">
            <ErrorBoundary>
              <BlogManager ref={blogManagerRef} />
            </ErrorBoundary>
          </TabsContent>

          <TabsContent value="authors" className="space-y-6">
            <ErrorBoundary>
              <AuthorManager ref={authorManagerRef} />
            </ErrorBoundary>
          </TabsContent>

          <TabsContent value="categories" className="space-y-6">
            <ErrorBoundary>
              <CategoryManager ref={categoryManagerRef} />
            </ErrorBoundary>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
};

export default AdminContentPage;