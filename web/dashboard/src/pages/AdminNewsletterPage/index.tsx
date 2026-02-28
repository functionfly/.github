import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, MoreVertical, Mail, Users, TrendingUp, Calendar, Eye, Send, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { StatCard } from "@/components/common/StatCard";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthStore } from "@/stores/authStore";

// Types for newsletter data
interface NewsletterSubscriber {
  _id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  source: string;
  status: string;
  subscribedAt: string;
  unsubscribedAt?: string;
  tags?: string[];
  metadata?: any;
}

interface NewsletterCampaign {
  _id: string;
  title: string;
  subject: string;
  status: string;
  scheduledFor?: string;
  sentAt?: string;
  recipientCount?: number;
  openRate?: number;
  clickRate?: number;
  tags?: string[];
  createdBy?: { _id: string; name: string };
  metadata?: any;
}

interface NewsletterStats {
  title: string;
  value: number | string;
  change: { value: number; label: string };
  icon: React.ReactNode;
}

export function AdminNewsletterPage() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const [subscribers, setSubscribers] = useState<NewsletterSubscriber[]>([]);
  const [campaigns, setCampaigns] = useState<NewsletterCampaign[]>([]);
  const [stats, setStats] = useState<NewsletterStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTab, setActiveTab] = useState("subscribers");

  // Fetch newsletter data (mock data since Sanity was removed)
  useEffect(() => {
    const fetchNewsletterData = async () => {
      try {
        setLoading(true);

        // Mock subscriber data
        const subscribersData = [
          {
            _id: "1",
            email: "john.doe@example.com",
            firstName: "John",
            lastName: "Doe",
            source: "website",
            status: "subscribed",
            subscribedAt: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString(),
            tags: ["developer", "react"]
          },
          {
            _id: "2",
            email: "jane.smith@example.com",
            firstName: "Jane",
            lastName: "Smith",
            source: "blog",
            status: "subscribed",
            subscribedAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
            tags: ["designer", "ux"]
          },
          {
            _id: "3",
            email: "bob.johnson@example.com",
            firstName: "Bob",
            lastName: "Johnson",
            source: "newsletter",
            status: "unsubscribed",
            subscribedAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(),
            unsubscribedAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
            tags: ["manager"]
          }
        ];

        // Mock campaign data
        const campaignsData = [
          {
            _id: "1",
            title: "Welcome to FunctionFly",
            subject: "Get started with FunctionFly in 5 minutes",
            status: "sent",
            sentAt: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
            recipientCount: 1250,
            openRate: 42.5,
            clickRate: 8.3,
            tags: ["onboarding", "welcome"],
            createdBy: { _id: "admin1", name: "Admin" }
          },
          {
            _id: "2",
            title: "New Features Update",
            subject: "Check out our latest features and improvements",
            status: "scheduled",
            scheduledFor: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString(),
            recipientCount: 2100,
            tags: ["features", "update"],
            createdBy: { _id: "admin1", name: "Admin" }
          },
          {
            _id: "3",
            title: "Community Spotlight",
            subject: "Meet the developers building amazing things",
            status: "draft",
            tags: ["community", "spotlight"],
            createdBy: { _id: "admin2", name: "Editor" }
          }
        ];

        // Calculate stats
        const activeSubscribers = subscribersData.filter((s: NewsletterSubscriber) => s.status === 'subscribed').length;
        const activeCampaigns = campaignsData.filter((c: NewsletterCampaign) => c.status === 'sent' || c.status === 'scheduled').length;

        // Calculate average rates from sent campaigns
        const sentCampaigns = campaignsData.filter((c: NewsletterCampaign) => c.status === 'sent' && c.openRate && c.clickRate);
        const avgOpenRate = sentCampaigns.length > 0
          ? sentCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.openRate || 0), 0) / sentCampaigns.length
          : 0;
        const avgClickRate = sentCampaigns.length > 0
          ? sentCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.clickRate || 0), 0) / sentCampaigns.length
          : 0;

        // Calculate month-over-month changes
        const now = new Date();
        const lastMonth = new Date(now.getFullYear(), now.getMonth() - 1, now.getDate());
        const twoMonthsAgo = new Date(now.getFullYear(), now.getMonth() - 2, now.getDate());

        // Subscribers: Calculate growth based on subscription dates
        const currentMonthSubs = subscribersData.filter((s: NewsletterSubscriber) => {
          const subDate = new Date(s.subscribedAt);
          return subDate >= lastMonth;
        }).length;
        const previousMonthSubs = subscribersData.filter((s: NewsletterSubscriber) => {
          const subDate = new Date(s.subscribedAt);
          return subDate >= twoMonthsAgo && subDate < lastMonth;
        }).length;
        const subscriberChange = previousMonthSubs > 0 ? currentMonthSubs - previousMonthSubs : currentMonthSubs;

        // Campaigns: Calculate based on sent campaigns in each period
        const currentMonthCampaigns = campaignsData.filter((c: NewsletterCampaign) => {
          if (c.sentAt) {
            const sentDate = new Date(c.sentAt);
            return sentDate >= lastMonth;
          }
          return false;
        }).length;
        const previousMonthCampaigns = campaignsData.filter((c: NewsletterCampaign) => {
          if (c.sentAt) {
            const sentDate = new Date(c.sentAt);
            return sentDate >= twoMonthsAgo && sentDate < lastMonth;
          }
          return false;
        }).length;
        const campaignChange = previousMonthCampaigns > 0 ? currentMonthCampaigns - previousMonthCampaigns : currentMonthCampaigns;

        // Rates: Calculate based on recent vs older campaigns (last 30 days vs previous 30 days)
        const recentCampaigns = sentCampaigns.filter((c: NewsletterCampaign) => {
          if (c.sentAt) {
            const sentDate = new Date(c.sentAt);
            return sentDate >= lastMonth;
          }
          return false;
        });
        const olderCampaigns = sentCampaigns.filter((c: NewsletterCampaign) => {
          if (c.sentAt) {
            const sentDate = new Date(c.sentAt);
            return sentDate >= twoMonthsAgo && sentDate < lastMonth;
          }
          return false;
        });

        const recentAvgOpenRate = recentCampaigns.length > 0
          ? recentCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.openRate || 0), 0) / recentCampaigns.length
          : avgOpenRate;
        const olderAvgOpenRate = olderCampaigns.length > 0
          ? olderCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.openRate || 0), 0) / olderCampaigns.length
          : avgOpenRate;
        const openRateChange = recentAvgOpenRate - olderAvgOpenRate;

        const recentAvgClickRate = recentCampaigns.length > 0
          ? recentCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.clickRate || 0), 0) / recentCampaigns.length
          : avgClickRate;
        const olderAvgClickRate = olderCampaigns.length > 0
          ? olderCampaigns.reduce((sum: number, c: NewsletterCampaign) => sum + (c.clickRate || 0), 0) / olderCampaigns.length
          : avgClickRate;
        const clickRateChange = recentAvgClickRate - olderAvgClickRate;

        const calculatedStats: NewsletterStats[] = [
          {
            title: "Total Subscribers",
            value: activeSubscribers,
            change: { value: subscriberChange, label: "from last month" },
            icon: <Users className="w-5 h-5 text-brand-500" />,
          },
          {
            title: "Active Campaigns",
            value: activeCampaigns,
            change: { value: campaignChange, label: "from last month" },
            icon: <Mail className="w-5 h-5 text-status-online" />,
          },
          {
            title: "Avg. Open Rate",
            value: `${avgOpenRate.toFixed(1)}%`,
            change: { value: parseFloat(openRateChange.toFixed(1)), label: "from last month" },
            icon: <TrendingUp className="w-5 h-5 text-status-degraded" />,
          },
          {
            title: "Avg. Click Rate",
            value: `${avgClickRate.toFixed(1)}%`,
            change: { value: parseFloat(clickRateChange.toFixed(1)), label: "from last month" },
            icon: <Eye className="w-5 h-5 text-status-offline" />,
          },
        ];

        setSubscribers(subscribersData);
        setCampaigns(campaignsData);
        setStats(calculatedStats);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch newsletter data:', err);
        setError('Failed to load newsletter data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchNewsletterData();
  }, []);

  const filteredSubscribers = subscribers.filter(subscriber =>
    subscriber.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (subscriber.firstName && subscriber.firstName.toLowerCase().includes(searchQuery.toLowerCase())) ||
    (subscriber.lastName && subscriber.lastName.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const filteredCampaigns = campaigns.filter(campaign =>
    campaign.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    campaign.subject.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    const variants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
      subscribed: "default",
      unsubscribed: "secondary",
      bounced: "destructive",
      complained: "destructive",
      draft: "outline",
      scheduled: "secondary",
      sent: "default",
      failed: "destructive",
    };
    const variant = variants[status] || "outline";
    return <Badge variant={variant}>{status}</Badge>;
  };

  const getSourceLabel = (source: string) => {
    const labels = {
      footer: "Footer",
      blog_cta: "Blog CTA",
      landing_page: "Landing Page",
      case_study: "Case Study",
      manual: "Manual",
      other: "Other",
    };
    return labels[source as keyof typeof labels] || source;
  };

  // Ensure user is authenticated
  if (!user) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Newsletter Management</h1>
            <p className="text-red-600">You must be logged in to access this page.</p>
          </div>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Newsletter Management</h1>
            <p className="text-muted-foreground">Loading newsletter data...</p>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-8 bg-gray-200 rounded w-1/2"></div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Newsletter Management</h1>
            <p className="text-red-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
          <h1 className="text-3xl font-bold text-text-primary">Newsletter Management</h1>
          <p className="text-text-secondary">
            Manage subscribers and campaigns
          </p>
        </div>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          New Campaign
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
        <Input
          placeholder="Search..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10"
        />
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="subscribers">Subscribers</TabsTrigger>
          <TabsTrigger value="campaigns">Campaigns</TabsTrigger>
        </TabsList>

        <TabsContent value="subscribers" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Newsletter Subscribers</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {filteredSubscribers.map((subscriber) => (
                  <div key={subscriber._id} className="flex items-center justify-between p-4 border border-border-subtle rounded-lg hover:bg-bg-hover transition-colors">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <div>
                          <p className="font-medium">
                            {subscriber.firstName && subscriber.lastName
                              ? `${subscriber.firstName} ${subscriber.lastName}`
                              : subscriber.email}
                          </p>
                          <p className="text-sm text-muted-foreground">{subscriber.email}</p>
                        </div>
                        {getStatusBadge(subscriber.status)}
                        <Badge variant="outline">
                          {getSourceLabel(subscriber.source)}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        Subscribed {new Date(subscriber.subscribedAt).toLocaleDateString()}
                      </p>
                    </div>

                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem>
                          <Eye className="w-4 h-4 mr-2" />
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <Mail className="w-4 h-4 mr-2" />
                          Send Test Email
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="campaigns" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Email Campaigns</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {filteredCampaigns.map((campaign) => (
                  <div key={campaign._id} className="flex items-center justify-between p-4 border border-border-subtle rounded-lg hover:bg-bg-hover transition-colors">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <div>
                          <h3 className="font-medium">{campaign.title}</h3>
                          <p className="text-sm text-muted-foreground">{campaign.subject}</p>
                        </div>
                        {getStatusBadge(campaign.status)}
                      </div>

                      <div className="flex items-center space-x-4 mt-2 text-sm text-muted-foreground">
                        {campaign.status === 'sent' && campaign.sentAt && (
                          <span>Sent {new Date(campaign.sentAt).toLocaleDateString()}</span>
                        )}
                        {campaign.status === 'scheduled' && campaign.scheduledFor && (
                          <span>
                            <Calendar className="w-4 h-4 inline mr-1" />
                            Scheduled for {new Date(campaign.scheduledFor).toLocaleDateString()}
                          </span>
                        )}
                        {campaign.recipientCount && (
                          <span>{campaign.recipientCount.toLocaleString()} recipients</span>
                        )}
                        {campaign.openRate && (
                          <span>{campaign.openRate}% open rate</span>
                        )}
                        {campaign.clickRate && (
                          <span>{campaign.clickRate}% click rate</span>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      {campaign.status === 'draft' && (
                        <Button size="sm">
                          <Send className="w-4 h-4 mr-1" />
                          Send Now
                        </Button>
                      )}
                      {campaign.status === 'scheduled' && (
                        <Button size="sm" variant="outline">
                          <Calendar className="w-4 h-4 mr-1" />
                          Edit Schedule
                        </Button>
                      )}

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <Eye className="w-4 h-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Mail className="w-4 h-4 mr-2" />
                            Duplicate
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}