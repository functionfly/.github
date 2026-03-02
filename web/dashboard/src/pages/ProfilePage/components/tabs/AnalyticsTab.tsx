/**
 * Analytics Tab Component
 *
 * Displays detailed analytics charts and metrics for user's functions.
 */

import { format } from "date-fns";
import { motion } from "framer-motion";
import { Map, Monitor, BarChart3 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LineChart } from "@/components/common/LineChart";
import { BarChart } from "@/components/common/BarChart";
import { tabContentVariants } from "../../animations";
import type { ProfileAnalytics } from "@/types";

export interface AnalyticsTabProps {
  analytics: ProfileAnalytics;
}

export function AnalyticsTab({ analytics }: AnalyticsTabProps) {
  const executionChartData = analytics.executionHistory.map(h => ({
    date: format(new Date(h.date), "MMM d"),
    executions: h.executions,
    users: h.uniqueUsers,
  }));

  const popularFunctionsData = analytics.popularFunctions.map(f => ({
    name: f.name.length > 20 ? f.name.slice(0, 20) + "..." : f.name,
    executions: f.executions,
  }));

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Execution Chart */}
      <LineChart
        data={executionChartData}
        series={[
          { key: "executions", name: "Executions", color: "#6366f1" },
          { key: "users", name: "Unique Users", color: "#10b981" },
        ]}
        title="Execution History"
        xAxisKey="date"
        height={300}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Popular Functions */}
        <BarChart
          data={popularFunctionsData}
          series={[{ key: "executions", name: "Executions", color: "#6366f1" }]}
          title="Popular Functions"
          xAxisKey="name"
          height={300}
          layout="horizontal"
        />

        {/* Geographic Distribution */}
        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Map className="w-5 h-5 text-brand-500" />
              Geographic Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.geographicDistribution.slice(0, 6).map((country) => (
                <div key={country.country} className="flex items-center gap-3">
                  <span className="text-sm text-text-secondary w-24">{country.country}</span>
                  <div className="flex-1">
                    <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                      <div
                        className="h-full bg-brand-500 rounded-full"
                        style={{ width: `${country.percentage}%` }}
                      />
                    </div>
                  </div>
                  <span className="text-sm font-medium text-text-primary w-16 text-right">
                    {country.percentage}%
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Device & Browser Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Monitor className="w-5 h-5 text-brand-500" />
              Device Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.deviceStats.map((device) => (
                <div key={device.device} className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">{device.device}</span>
                  <span className="text-sm font-medium text-text-primary">{device.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-border-subtle">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-brand-500" />
              Browser Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.browserStats.map((browser) => (
                <div key={browser.browser} className="flex items-center justify-between">
                  <span className="text-sm text-text-secondary">{browser.browser}</span>
                  <span className="text-sm font-medium text-text-primary">{browser.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  );
}
