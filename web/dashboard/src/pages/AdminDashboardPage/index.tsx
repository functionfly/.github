import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Footer } from "@/pages/LandingPage/components/Footer";
import {
  Users,
  Building2,
  CreditCard,
  Shield,
  Settings,
  Mail,
  Calendar,
  FileText,
  MessageSquare,
  ExternalLink,
  Code,
  Database
} from "lucide-react";

const adminSections = [
  {
    title: "Tenants",
    description: "Manage multi-tenant organizations and their settings",
    path: "/admin/tenants",
    icon: <Building2 className="w-6 h-6 text-blue-500" />,
    color: "bg-blue-50 border-blue-200 hover:border-blue-300"
  },
  {
    title: "Users",
    description: "User management, roles, and permissions",
    path: "/admin/users",
    icon: <Users className="w-6 h-6 text-green-500" />,
    color: "bg-green-50 border-green-200 hover:border-green-300"
  },
  {
    title: "Billing",
    description: "Subscription management and billing operations",
    path: "/admin/billing",
    icon: <CreditCard className="w-6 h-6 text-purple-500" />,
    color: "bg-purple-50 border-purple-200 hover:border-purple-300"
  },
  {
    title: "Audit",
    description: "Security audit logs and system monitoring",
    path: "/admin/audit",
    icon: <Shield className="w-6 h-6 text-red-500" />,
    color: "bg-red-50 border-red-200 hover:border-red-300"
  },
  {
    title: "System",
    description: "System configuration and maintenance",
    path: "/admin/system",
    icon: <Settings className="w-6 h-6 text-gray-500" />,
    color: "bg-gray-50 border-gray-200 hover:border-gray-300"
  },
  {
    title: "Newsletter",
    description: "Newsletter management and subscriber lists",
    path: "/admin/newsletter",
    icon: <Mail className="w-6 h-6 text-indigo-500" />,
    color: "bg-indigo-50 border-indigo-200 hover:border-indigo-300"
  },
  {
    title: "Content Calendar",
    description: "Content planning and publication schedule",
    path: "/admin/content-calendar",
    icon: <Calendar className="w-6 h-6 text-orange-500" />,
    color: "bg-orange-50 border-orange-200 hover:border-orange-300"
  },
  {
    title: "Content",
    description: "Content management and publishing",
    path: "/admin/content",
    icon: <FileText className="w-6 h-6 text-teal-500" />,
    color: "bg-teal-50 border-teal-200 hover:border-teal-300"
  },
  {
    title: "Feedback",
    description: "User feedback and support tickets",
    path: "/admin/feedback",
    icon: <MessageSquare className="w-6 h-6 text-pink-500" />,
    color: "bg-pink-50 border-pink-200 hover:border-pink-300"
  },
  {
    title: "Functions",
    description: "Manage all functions across tenants",
    path: "/admin/functions",
    icon: <Code className="w-6 h-6 text-indigo-500" />,
    color: "bg-indigo-50 border-indigo-200 hover:border-indigo-300"
  },
  {
    title: "Registry",
    description: "Function registry management and moderation",
    path: "/admin/registry",
    icon: <Database className="w-6 h-6 text-cyan-500" />,
    color: "bg-cyan-50 border-cyan-200 hover:border-cyan-300"
  }
];

export function AdminDashboardPage() {
  const navigate = useNavigate();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="text-center lg:text-left">
        <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
          <span className="text-text-primary text-glow">Admin Dashboard</span>
        </h1>
        <p className="text-text-secondary text-lg">
          Manage and monitor your FunctionFly platform
        </p>
      </div>

      {/* Admin Sections Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {adminSections.map((section) => (
          <Card
            key={section.path}
            className={`glass-card glow hover-lift cursor-pointer transition-all duration-200 ${section.color}`}
            onClick={() => navigate(section.path)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/50">
                  {section.icon}
                </div>
                <div className="flex-1">
                  <CardTitle className="text-text-primary text-lg">
                    {section.title}
                  </CardTitle>
                </div>
                <ExternalLink className="w-4 h-4 text-text-muted opacity-50" />
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-text-secondary text-sm leading-relaxed">
                {section.description}
              </p>
              <Button
                variant="ghost"
                size="sm"
                className="mt-3 w-full justify-start p-0 h-auto font-normal text-text-primary hover:text-brand-500"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(section.path);
                }}
              >
                Manage {section.title.toLowerCase()} →
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Quick Stats or System Overview could go here */}
      <div className="mt-8 p-6 glass-card glow rounded-lg">
        <h3 className="text-lg font-semibold text-text-primary mb-2">System Overview</h3>
        <p className="text-text-secondary">
          Welcome to the admin dashboard. Use the cards above to access different administrative functions.
        </p>
      </div>

      {/* Footer */}
      <div className="mt-16">
        <Footer />
      </div>
    </div>
  );
}
