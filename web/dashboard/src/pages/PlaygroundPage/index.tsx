import { useEffect, useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { ErrorMessage } from '@/components/common/ErrorMessage';
import { useTranslation } from 'react-i18next';

import { usePlaygroundStore, FunctionInfo } from './store/playgroundStore';
import { usePlaygroundKeyboard } from './hooks/usePlaygroundKeyboard';
import { useResizablePanels } from './hooks/useResizablePanels';

import { PlaygroundHeader } from './components/PlaygroundHeader';
import { PlaygroundToolbar } from './components/PlaygroundToolbar';
import { PlaygroundInputPanel } from './components/PlaygroundInputPanel';
import { PlaygroundOutputPanel } from './components/PlaygroundOutputPanel';
import { PlaygroundSidebar } from './components/PlaygroundSidebar';
import { PlaygroundStatusBar } from './components/PlaygroundStatusBar';

export function PlaygroundPage() {
  const { t } = useTranslation();
  const { author, name } = useParams<{ author: string; name: string }>();
  const [searchParams] = useSearchParams();

  const {
    setFunctionInfo,
    setInputValue,
    setInputJson,
    sidebarOpen,
    executionHistory,
    loadFromHistory,
  } = usePlaygroundStore();

  // History navigation state
  const historyIndexRef = useRef(-1);
  const [, forceUpdate] = useState(0);

  // Fetch function info
  const { data: functionInfo, isLoading, error } = useQuery<FunctionInfo>({
    queryKey: ['function', author, name],
    queryFn: async () => {
      const response = await fetch(
        `/v1/registry/functions/${author}/${name}?expand=manifest`
      );
    if (response.status === 404) throw new Error(t('playground.functionNotFound'));
    if (!response.ok) throw new Error('Failed to fetch function');
      return response.json();
    },
    enabled: !!author && !!name,
  });

  // Sync function info into store
  useEffect(() => {
    if (functionInfo) {
      setFunctionInfo(functionInfo);
    }
  }, [functionInfo, setFunctionInfo]);

  // Parse input from URL params
  useEffect(() => {
    const inputParam = searchParams.get('input');
    if (inputParam) {
      try {
        let decoded: unknown;
        try {
          decoded = JSON.parse(atob(inputParam.replace(/-/g, '+').replace(/_/g, '/')));
        } catch {
          decoded = JSON.parse(decodeURIComponent(inputParam));
        }
        setInputValue(decoded);
        setInputJson(JSON.stringify(decoded, null, 2));
      } catch {
        // ignore invalid input param
      }
    }
  }, [searchParams, setInputValue, setInputJson]);

  // History navigation
  const handleNavigateHistory = (direction: 'prev' | 'next') => {
    const history = executionHistory;
    if (history.length === 0) return;

    if (direction === 'prev') {
      historyIndexRef.current = Math.min(
        historyIndexRef.current + 1,
        history.length - 1
      );
    } else {
      historyIndexRef.current = Math.max(historyIndexRef.current - 1, -1);
    }

    if (historyIndexRef.current >= 0) {
      loadFromHistory(history[historyIndexRef.current]);
    }
    forceUpdate((n) => n + 1);
  };

  // Keyboard shortcuts
  usePlaygroundKeyboard({
    author: author || '',
    name: name || '',
    onNavigateHistory: handleNavigateHistory,
  });

  // Resizable panels
  const { sizes, containerRef, handleMouseDown, resetToEqual } = useResizablePanels();

  // ─── Loading / Error states ────────────────────────────────────────────────

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <LoadingSpinner />
        </main>
        <Footer />
      </div>
    );
  }

  if (error) {
    const isNotFound = (error as Error).message === t('playground.functionNotFound');
    if (isNotFound) {
      return (
        <div className="min-h-screen flex flex-col bg-bg-primary">
          <Navbar variant="landing" />
          <main className="flex-1 pt-16 flex items-center justify-center">
            <div className="text-center">
              <h1 className="text-2xl font-bold mb-2">{t('playground.functionNotFound')}</h1>
              <p className="text-muted-foreground">
                {t('playground.functionNotFoundDescription', { author, name })}
              </p>
              <Link to="/registry">
                <Button variant="outline" className="mt-4">
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  {t('playground.backToRegistry')}
                </Button>
              </Link>
            </div>
          </main>
          <Footer />
        </div>
      );
    }
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <ErrorMessage error={error as Error} />
        </main>
        <Footer />
      </div>
    );
  }

  if (!functionInfo) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <div className="text-center">
            <h1 className="text-2xl font-bold mb-2">{t('playground.functionNotFound')}</h1>
            <p className="text-muted-foreground">
              {t('playground.functionNotFoundDescription', { author, name })}
            </p>
            <Link to="/registry">
              <Button variant="outline" className="mt-4">
                <ArrowLeft className="w-4 h-4 mr-2" />
                {t('playground.backToRegistry')}
              </Button>
            </Link>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  // ─── Main IDE layout ───────────────────────────────────────────────────────

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />

      <main className="flex-1 pt-16 flex flex-col overflow-hidden">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.3 }}
          className="flex flex-col h-[calc(100vh-4rem)]"
        >
          {/* Header */}
          <PlaygroundHeader functionInfo={functionInfo} />

          {/* Toolbar */}
          <PlaygroundToolbar author={functionInfo.author} name={functionInfo.name} />

          {/* Main panels area */}
          <div
            ref={containerRef}
            className="flex flex-1 overflow-hidden"
          >
            {/* Input panel */}
            <div
              className="flex flex-col overflow-hidden border-r border-border-subtle"
              style={{ width: `${sizes.input}%` }}
            >
              <PlaygroundInputPanel className="h-full" />
            </div>

            {/* Resize handle */}
            <div
              className="w-1 bg-border-subtle hover:bg-indigo-500/50 cursor-col-resize transition-colors shrink-0 relative group"
              onMouseDown={handleMouseDown}
              onDoubleClick={resetToEqual}
              title={t('playground.resizeTooltip')}
            >
              <div className="absolute inset-y-0 -left-1 -right-1 group-hover:bg-indigo-500/10" />
            </div>

            {/* Output panel */}
            <div
              className="flex flex-col overflow-hidden"
              style={{
                width: sidebarOpen
                  ? `calc(${sizes.output}% - 280px)`
                  : `${sizes.output}%`,
              }}
            >
              <PlaygroundOutputPanel className="h-full" />
            </div>

            {/* Sidebar */}
            {sidebarOpen && (
              <motion.div
                initial={{ width: 0, opacity: 0 }}
                animate={{ width: 280, opacity: 1 }}
                exit={{ width: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="shrink-0 overflow-hidden"
                style={{ width: 280 }}
              >
                <PlaygroundSidebar className="h-full w-full" />
              </motion.div>
            )}
          </div>

          {/* Status bar */}
          <PlaygroundStatusBar />
        </motion.div>
      </main>
    </div>
  );
}
