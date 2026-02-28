'use client';

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  MessageSquare,
  Search,
  Eye,
  CheckCircle,
  Clock,
  AlertTriangle,
  BarChart3,
  Download,
  FileText,
  TrendingUp,
  PieChart,
  ArrowLeft,
  LineChart,
  Activity
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import { StatCard } from '@/components/common/StatCard';
import { toast } from 'sonner';

interface Feedback {
  id: string;
  user_id?: string;
  user_email?: string;
  feedback_type: string;
  subject: string;
  message: string;
  priority: string;
  status: string;
  browser_info?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
  updated_at: string;
  attachments?: FeedbackAttachment[];
}

interface FeedbackAttachment {
  id: string;
  filename: string;
  content_type: string;
  size: number;
  s3_key: string;
}

interface FeedbackStats {
  total: number;
  status_breakdown: Record<string, number>;
  type_breakdown: Record<string, number>;
  trends?: {
    daily: Array<{ date: string; count: number }>;
    weekly: Array<{ week: string; count: number }>;
    monthly: Array<{ month: string; count: number }>;
  };
}

const mockFeedback: Feedback[] = [
  {
    id: 'fb-1',
    user_email: 'user1@example.com',
    feedback_type: 'bug',
    subject: 'App crashes on login',
    message: 'The application crashes when trying to log in with valid credentials. Steps to reproduce: 1. Open app 2. Enter email 3. Enter password 4. Click login 5. App crashes',
    priority: 'high',
    status: 'submitted',
    browser_info: 'Chrome 120.0.0.0',
    created_at: '2026-02-15T10:30:00Z',
    updated_at: '2026-02-15T10:30:00Z',
    attachments: [
      {
        id: 'att-1',
        filename: 'crash-log.txt',
        content_type: 'text/plain',
        size: 2048,
        s3_key: 'feedback/fb-1/crash-log.txt'
      }
    ]
  },
  {
    id: 'fb-2',
    user_id: 'user-123',
    feedback_type: 'feature',
    subject: 'Add dark mode support',
    message: 'It would be great to have a dark mode option for better user experience, especially when using the app at night. This would help reduce eye strain and battery usage on OLED displays.',
    priority: 'medium',
    status: 'in-review',
    browser_info: 'Firefox 121.0',
    created_at: '2026-02-14T15:45:00Z',
    updated_at: '2026-02-15T09:15:00Z'
  },
  {
    id: 'fb-3',
    feedback_type: 'improvement',
    subject: 'Faster loading times',
    message: 'The dashboard loads quite slowly, especially on slower connections. Could we implement some optimizations like code splitting or lazy loading?',
    priority: 'medium',
    status: 'resolved',
    browser_info: 'Safari 17.0',
    created_at: '2026-02-13T12:20:00Z',
    updated_at: '2026-02-15T08:30:00Z'
  }
];

const mockStats: FeedbackStats = {
  total: 247,
  status_breakdown: {
    submitted: 45,
    'in-review': 23,
    resolved: 156,
    closed: 23
  },
  type_breakdown: {
    bug: 89,
    feature: 67,
    improvement: 58,
    general: 33
  },
  trends: {
    daily: [
      { date: '2026-02-09', count: 12 },
      { date: '2026-02-10', count: 8 },
      { date: '2026-02-11', count: 15 },
      { date: '2026-02-12', count: 22 },
      { date: '2026-02-13', count: 18 },
      { date: '2026-02-14', count: 25 },
      { date: '2026-02-15', count: 31 }
    ],
    weekly: [
      { week: 'Week 1', count: 45 },
      { week: 'Week 2', count: 52 },
      { week: 'Week 3', count: 38 },
      { week: 'Week 4', count: 67 },
      { week: 'Week 5', count: 45 }
    ],
    monthly: [
      { month: 'Dec 2025', count: 145 },
      { month: 'Jan 2026', count: 178 },
      { month: 'Feb 2026', count: 247 }
    ]
  }
};

export function AdminFeedbackPage() {
  const navigate = useNavigate();
  const [feedback, setFeedback] = useState<Feedback[]>(mockFeedback);
  const [stats, setStats] = useState<FeedbackStats>(mockStats);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [priorityFilter, setPriorityFilter] = useState<string>('all');
  const [selectedFeedback, setSelectedFeedback] = useState<Feedback | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Filter feedback
  const filteredFeedback = feedback.filter(item => {
    const matchesSearch = item.subject.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         item.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (item.user_email && item.user_email.toLowerCase().includes(searchTerm.toLowerCase()));
    const matchesStatus = statusFilter === 'all' || item.status === statusFilter;
    const matchesType = typeFilter === 'all' || item.feedback_type === typeFilter;
    const matchesPriority = priorityFilter === 'all' || item.priority === priorityFilter;
    return matchesSearch && matchesStatus && matchesType && matchesPriority;
  });

  // Status color mapping
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'submitted': return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20';
      case 'in-review': return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20';
      case 'resolved': return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20';
      case 'closed': return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20';
      default: return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20';
    }
  };

  // Priority color mapping
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical': return 'text-red-600 dark:text-red-400';
      case 'high': return 'text-orange-600 dark:text-orange-400';
      case 'medium': return 'text-amber-600 dark:text-amber-400';
      case 'low': return 'text-emerald-600 dark:text-emerald-400';
      default: return 'text-text-secondary';
    }
  };

  // Type icon mapping
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'bug': return <AlertTriangle className="w-4 h-4 text-red-400" />;
      case 'feature': return <Activity className="w-4 h-4 text-blue-400" />;
      case 'improvement': return <TrendingUp className="w-4 h-4 text-amber-400" />;
      case 'general': return <MessageSquare className="w-4 h-4 text-text-muted" />;
      default: return <MessageSquare className="w-4 h-4 text-text-muted" />;
    }
  };

  const handleStatusUpdate = async (feedbackId: string, newStatus: string) => {
    setIsLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/feedback/${feedbackId}/status`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({ status: newStatus })
      });

      if (!response.ok) {
        throw new Error('Failed to update status');
      }

      setFeedback(prev => prev.map(item =>
        item.id === feedbackId
          ? { ...item, status: newStatus, updated_at: new Date().toISOString() }
          : item
      ));

      toast.success(`Feedback status updated to ${newStatus}`);
    } catch (error) {
      toast.error('Failed to update feedback status');
    } finally {
      setIsLoading(false);
    }
  };

  const handleExport = async (format: 'csv' | 'json') => {
    setIsLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/feedback/export?format=${format}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Export failed');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `feedback-export.${format}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);

      toast.success(`Feedback data exported as ${format.toUpperCase()}`);
    } catch (error) {
      toast.error('Failed to export feedback data');
    } finally {
      setIsLoading(false);
    }
  };

  const loadFeedback = async () => {
    setIsLoading(true);
    try {
      const [feedbackRes, analyticsRes] = await Promise.all([
        fetch('/api/v1/admin/feedback', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
          },
        }),
        fetch('/api/v1/admin/feedback/analytics', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
          },
        })
      ]);

      if (!feedbackRes.ok || !analyticsRes.ok) {
        throw new Error('Failed to fetch data');
      }

      const feedbackData = await feedbackRes.json();
      const analyticsData = await analyticsRes.json();

      setFeedback(feedbackData.feedback || []);
      setStats({
        total: analyticsData.stats.total || 0,
        status_breakdown: analyticsData.stats.status_breakdown || {},
        type_breakdown: analyticsData.stats.type_breakdown || {},
        trends: analyticsData.trends || {},
      });
    } catch (error) {
      console.error('Failed to load feedback data:', error);
      // Fallback to mock data if API fails
      setFeedback(mockFeedback);
      setStats(mockStats);
      toast.error('Failed to load feedback data, showing sample data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadFeedback();
  }, []);

  const statCards = [
    {
      title: 'Total Feedback',
      value: stats.total,
      change: { value: 15, label: 'from last month' },
      icon: <MessageSquare className="w-5 h-5 text-[#6366f1]" />,
      trend: 'up' as const,
    },
    {
      title: 'Open Items',
      value: (stats.status_breakdown.submitted || 0) + (stats.status_breakdown['in-review'] || 0),
      change: { value: -8, label: 'from last week' },
      icon: <Clock className="w-5 h-5 text-[#6366f1]" />,
      trend: 'down' as const,
    },
    {
      title: 'Resolution Rate',
      value: `${Math.round(((stats.status_breakdown.resolved || 0) + (stats.status_breakdown.closed || 0)) / stats.total * 100)}%`,
      change: { value: 12, label: 'from last month' },
      icon: <CheckCircle className="w-5 h-5 text-[#6366f1]" />,
      trend: 'up' as const,
    },
    {
      title: 'Avg Response Time',
      value: '2.3 days',
      change: { value: -5, label: 'from last month' },
      icon: <Activity className="w-5 h-5 text-[#6366f1]" />,
      trend: 'down' as const,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Feedback Management</h1>
          <p className="text-text-secondary">Monitor and manage user feedback submissions</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => handleExport('csv')}
            disabled={isLoading}
            className="border-border-subtle hover:bg-bg-hover"
          >
            <Download className="w-4 h-4 mr-2" />
            Export CSV
          </Button>
          <Button
            variant="outline"
            onClick={() => handleExport('json')}
            disabled={isLoading}
            className="text-text-secondary border-border-default hover:bg-bg-hover"
          >
            <FileText className="w-4 h-4 mr-2" />
            Export JSON
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Main Content */}
      <Tabs defaultValue="list" className="space-y-4">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="list">Feedback List</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
        </TabsList>

        <TabsContent value="list" className="space-y-4">
          {/* Filters */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col lg:flex-row gap-4">
                <div className="flex-1">
                  <div className="relative">
                    <Search className="absolute left-3 top-3 h-4 w-4 text-text-secondary" />
                    <Input
                      placeholder="Search feedback..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10"
                    />
                  </div>
                </div>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-full lg:w-40">
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="submitted">Submitted</SelectItem>
                    <SelectItem value="in-review">In Review</SelectItem>
                    <SelectItem value="resolved">Resolved</SelectItem>
                    <SelectItem value="closed">Closed</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger className="w-full lg:w-40">
                    <SelectValue placeholder="Type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Types</SelectItem>
                    <SelectItem value="bug">Bug</SelectItem>
                    <SelectItem value="feature">Feature</SelectItem>
                    <SelectItem value="improvement">Improvement</SelectItem>
                    <SelectItem value="general">General</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={priorityFilter} onValueChange={setPriorityFilter}>
                  <SelectTrigger className="w-full lg:w-40">
                    <SelectValue placeholder="Priority" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Priorities</SelectItem>
                    <SelectItem value="critical">Critical</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          {/* Feedback List */}
          <div className="space-y-4">
            {filteredFeedback.map((item) => (
              <Card key={item.id} className="hover:bg-white/5 transition-colors">
                <CardContent className="pt-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-2">
                        {getTypeIcon(item.feedback_type)}
                        <h3 className="font-semibold text-text-primary truncate">{item.subject}</h3>
                        <Badge className={`${getStatusColor(item.status)} text-xs`}>
                          {item.status}
                        </Badge>
                        <Badge variant="outline" className={`text-xs ${getPriorityColor(item.priority)}`}>
                          {item.priority}
                        </Badge>
                      </div>

                      <p className="text-text-secondary text-sm mb-3 line-clamp-2">
                        {item.message}
                      </p>

                      <div className="flex items-center justify-between text-xs text-text-muted">
                        <div className="flex items-center gap-4">
                          <span>{item.user_email || 'Anonymous'}</span>
                          <span>{new Date(item.created_at).toLocaleDateString()}</span>
                          {item.attachments && item.attachments.length > 0 && (
                            <span className="flex items-center gap-1">
                              <FileText className="w-3 h-3" />
                              {item.attachments.length} attachment{item.attachments.length !== 1 ? 's' : ''}
                            </span>
                          )}
                        </div>

                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSelectedFeedback(item)}
                            className="text-text-secondary hover:text-text-primary"
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Select
                            value={item.status}
                            onValueChange={(value) => handleStatusUpdate(item.id, value)}
                          >
                            <SelectTrigger className="w-32 h-8">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="submitted">Submitted</SelectItem>
                              <SelectItem value="in-review">In Review</SelectItem>
                              <SelectItem value="resolved">Resolved</SelectItem>
                              <SelectItem value="closed">Closed</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="analytics" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Status Breakdown */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <PieChart className="w-5 h-5" />
                  Status Distribution
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {Object.entries(stats.status_breakdown).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className={`w-3 h-3 rounded-full ${
                        status === 'submitted' ? 'bg-blue-400' :
                        status === 'in-review' ? 'bg-amber-400' :
                        status === 'resolved' ? 'bg-emerald-400' : 'bg-gray-400'
                      }`} />
                      <span className="text-sm capitalize text-text-primary">{status.replace('-', ' ')}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-text-secondary">{count}</span>
                      <Progress
                        value={(count / stats.total) * 100}
                        className="w-20 h-2"
                      />
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            {/* Type Breakdown */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="w-5 h-5" />
                  Feedback Types
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {Object.entries(stats.type_breakdown).map(([type, count]) => (
                  <div key={type} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {getTypeIcon(type)}
                      <span className="text-sm capitalize text-text-primary">{type}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-text-secondary">{count}</span>
                      <Progress
                        value={(count / stats.total) * 100}
                        className="w-20 h-2"
                      />
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            {/* Trends */}
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <LineChart className="w-5 h-5" />
                  Feedback Trends
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  <div>
                    <h4 className="text-sm font-medium text-text-primary mb-3">Daily Submissions (Last 7 Days)</h4>
                    <div className="flex items-end gap-2 h-32">
                      {stats.trends?.daily.map((day, index) => (
                        <div key={day.date} className="flex-1 flex flex-col items-center gap-1">
                          <div
                            className="w-full bg-[#6366f1] rounded-t"
                            style={{ height: `${(day.count / 35) * 100}%` }}
                          />
                          <span className="text-xs text-text-secondary">
                            {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div>
                    <h4 className="text-sm font-medium text-text-primary mb-3">Monthly Growth</h4>
                    <div className="flex items-end gap-4 h-24">
                      {stats.trends?.monthly.map((month, index) => (
                        <div key={month.month} className="flex-1 flex flex-col items-center gap-1">
                          <div
                            className="w-full bg-emerald-500 rounded-t"
                            style={{ height: `${(month.count / 200) * 100}%` }}
                          />
                          <span className="text-xs text-text-secondary">{month.month}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* Feedback Detail Dialog */}
      {selectedFeedback && (
        <Dialog open={!!selectedFeedback} onOpenChange={() => setSelectedFeedback(null)}>
          <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-text-primary">
                {getTypeIcon(selectedFeedback.feedback_type)}
                {selectedFeedback.subject}
              </DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <Badge className={getStatusColor(selectedFeedback.status)}>
                  {selectedFeedback.status}
                </Badge>
                <Badge variant="outline" className={getPriorityColor(selectedFeedback.priority)}>
                  {selectedFeedback.priority} priority
                </Badge>
                <span className="text-sm text-text-secondary">
                  {new Date(selectedFeedback.created_at).toLocaleString()}
                </span>
              </div>

              <div>
                <h4 className="font-medium text-text-primary mb-2">From</h4>
                <p className="text-text-secondary">{selectedFeedback.user_email || 'Anonymous'}</p>
              </div>

              <div>
                <h4 className="font-medium text-text-primary mb-2">Message</h4>
                <p className="text-text-secondary whitespace-pre-wrap">{selectedFeedback.message}</p>
              </div>

              {selectedFeedback.browser_info && (
                <div>
                  <h4 className="font-medium text-text-primary mb-2">Browser Info</h4>
                  <p className="text-text-secondary">{selectedFeedback.browser_info}</p>
                </div>
              )}

              {selectedFeedback.attachments && selectedFeedback.attachments.length > 0 && (
                <div>
                  <h4 className="font-medium text-text-primary mb-2">Attachments</h4>
                  <div className="space-y-2">
                    {selectedFeedback.attachments.map((attachment) => (
                      <div key={attachment.id} className="flex items-center gap-2 p-2 bg-white/5 rounded">
                        <FileText className="w-4 h-4 text-text-secondary" />
                        <span className="text-sm text-text-primary">{attachment.filename}</span>
                        <span className="text-xs text-text-secondary">
                          ({Math.round(attachment.size / 1024)}KB)
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
