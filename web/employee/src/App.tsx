import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navigate, Route, Routes } from 'react-router-dom';
import { Toaster } from 'sonner';
import { EmployeeLayout } from './components/layout/EmployeeLayout';
import { AuthGuard } from './components/auth/AuthGuard';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { ProfilePage } from './pages/ProfilePage';
import { ProjectsPage } from './pages/ProjectsPage';
import { ProjectDetailPage } from './pages/ProjectDetailPage';
import { TasksPage } from './pages/TasksPage';
import { LearningPage } from './pages/LearningPage';
import { KnowledgePage } from './pages/KnowledgePage';
import { KnowledgeArticlePage } from './pages/KnowledgeArticlePage';
import { CompensationPage } from './pages/CompensationPage';
import { OrgChartPage } from './pages/OrgChartPage';
import { TeamPage } from './pages/TeamPage';
import { AIAssistantPage } from './pages/AIAssistantPage';
import { PerformancePage } from './pages/PerformancePage';
import { TimeTrackingPage } from './pages/TimeTrackingPage';
import { AnalyticsPage } from './pages/AnalyticsPage';
import { InnovationPage } from './pages/InnovationPage';
import { MarketplacePage } from './pages/MarketplacePage';
import { CareerPage } from './pages/CareerPage';
import { MentorshipPage } from './pages/MentorshipPage';
import { DocumentsPage } from './pages/DocumentsPage';
import { MissionControlPage } from './pages/MissionControlPage';
import { TeamHealthPage } from './pages/TeamHealthPage';
import { SkillsGraphPage } from './pages/SkillsGraphPage';
import { ReputationPage } from './pages/ReputationPage';
import { BadgesPage } from './pages/BadgesPage';
import { LivingMemoryPage } from './pages/LivingMemoryPage';
import { IncidentsPage } from './pages/IncidentsPage';
import { LifecyclePage } from './pages/LifecyclePage';
import { FeatureFlagsPage } from './pages/FeatureFlagsPage';
import { DataClassificationPage } from './pages/DataClassificationPage';
import { CertificatesPage } from './pages/CertificatesPage';
import { EventsPage } from './pages/EventsPage';
import { EmailProvisioningPage } from './pages/EmailProvisioningPage';
import { DeviceManagementPage } from './pages/DeviceManagementPage';
import { SSOProvisioningPage } from './pages/SSOProvisioningPage';
import { WalletBadgePage } from './pages/WalletBadgePage';
import { NotificationSettingsPage } from './pages/NotificationSettingsPage';
import { FeedbackRoundsPage } from './pages/FeedbackRoundsPage';
import { GoalCascadePage } from './pages/GoalCascadePage';
import { SignaturesPage } from './pages/SignaturesPage';
import { CertificatePKIPage } from './pages/CertificatePKIPage';
import { OrgImportPage } from './pages/OrgImportPage';
import { PackageRegistryPage } from './pages/PackageRegistryPage';
import { PassportPage } from './pages/PassportPage';
import { IdentityWalletPage } from './pages/IdentityWalletPage';
import { CareerTimelinePage } from './pages/CareerTimelinePage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthGuard>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<EmployeeLayout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/profile/:id?" element={<ProfilePage />} />
            <Route path="/projects" element={<ProjectsPage />} />
            <Route path="/projects/:id" element={<ProjectDetailPage />} />
            <Route path="/tasks" element={<TasksPage />} />
            <Route path="/learning" element={<LearningPage />} />
            <Route path="/knowledge" element={<KnowledgePage />} />
            <Route path="/knowledge/:slug" element={<KnowledgeArticlePage />} />
            <Route path="/compensation" element={<CompensationPage />} />
            <Route path="/orgchart" element={<OrgChartPage />} />
            <Route path="/team" element={<TeamPage />} />
            <Route path="/ai-assistant" element={<AIAssistantPage />} />
            <Route path="/performance" element={<PerformancePage />} />
            <Route path="/time-tracking" element={<TimeTrackingPage />} />
            <Route path="/analytics" element={<AnalyticsPage />} />
            <Route path="/innovation" element={<InnovationPage />} />
            <Route path="/marketplace" element={<MarketplacePage />} />
            <Route path="/career" element={<CareerPage />} />
            <Route path="/mentorship" element={<MentorshipPage />} />
            <Route path="/documents" element={<DocumentsPage />} />
            <Route path="/mission-control" element={<MissionControlPage />} />
            <Route path="/team-health" element={<TeamHealthPage />} />
            <Route path="/skills-graph" element={<SkillsGraphPage />} />
            <Route path="/reputation" element={<ReputationPage />} />
            <Route path="/badges" element={<BadgesPage />} />
            <Route path="/memory" element={<LivingMemoryPage />} />
            <Route path="/incidents" element={<IncidentsPage />} />
            <Route path="/lifecycle" element={<LifecyclePage />} />
            <Route path="/feature-flags" element={<FeatureFlagsPage />} />
            <Route path="/data-classification" element={<DataClassificationPage />} />
            <Route path="/certificates" element={<CertificatesPage />} />
            <Route path="/events" element={<EventsPage />} />
            <Route path="/email" element={<EmailProvisioningPage />} />
            <Route path="/devices" element={<DeviceManagementPage />} />
            <Route path="/sso-provisioning" element={<SSOProvisioningPage />} />
            <Route path="/wallet-badge" element={<WalletBadgePage />} />
            <Route path="/notification-settings" element={<NotificationSettingsPage />} />
            <Route path="/feedback-rounds" element={<FeedbackRoundsPage />} />
            <Route path="/goal-cascade" element={<GoalCascadePage />} />
            <Route path="/signatures" element={<SignaturesPage />} />
            <Route path="/certificate-pki" element={<CertificatePKIPage />} />
            <Route path="/org-import" element={<OrgImportPage />} />
            <Route path="/packages" element={<PackageRegistryPage />} />
            <Route path="/passport" element={<PassportPage />} />
            <Route path="/wallet" element={<IdentityWalletPage />} />
            <Route path="/career-timeline" element={<CareerTimelinePage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </AuthGuard>
      <Toaster position="top-right" richColors />
    </QueryClientProvider>
  );
}
