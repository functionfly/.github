import { PageViewTracker } from '@/components/PageViewTracker';
import HistoryPage from '@/pages/HistoryPage';
import IncidentDetailPage from '@/pages/IncidentDetailPage';
import StatusPage from '@/pages/StatusPage';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Route, Routes } from 'react-router-dom';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000, // 30 seconds
      refetchInterval: 60000, // 1 minute
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <PageViewTracker />
        <Routes>
          <Route path="/" element={<StatusPage />} />
          <Route path="/incidents/:id" element={<IncidentDetailPage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;