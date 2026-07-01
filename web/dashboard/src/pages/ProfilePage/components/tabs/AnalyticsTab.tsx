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

const CHART_ACCENT = 'var(--accent)';
const CHART_OK = 'var(--status-ok)';

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
          { key: "executions", name: "Executions", color: CHART_ACCENT },
          { key: "users", name: "Unique Users", color: CHART_OK },
        ]}
        title="Execution History"
        xAxisKey="date"
        height={300}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Popular Functions */}
        <BarChart
          data={popularFunctionsData}
          series={[{ key: "executions", name: "Executions", color: CHART_ACCENT }]}
          title="Popular Functions"
          xAxisKey="name"
          height={300}
          layout="horizontal"
        />

        {/* Geographic Distribution */}
        <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
              <Map className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
              Geographic Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.geographicDistribution.slice(0, 6).map((country) => (
                <div key={country.country} className="flex items-center gap-3">
                  <span className="text-sm w-24" style={{ color: 'var(--text-dim)' }}>{country.country}</span>
                  <div className="flex-1">
                    <div className="h-2 rounded-full overflow-hidden" style={{ background: 'var(--panel-edge)' }}>
                      <div
                        className="h-full rounded-full"
                        style={{ width: `${country.percentage}%`, background: 'var(--status-ok)' }}
                      />
                    </div>
                  </div>
                  <span className="text-sm font-medium font-mono tabular-nums w-16 text-right" style={{ color: 'var(--text)' }}>
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
        <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
              <Monitor className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
              Device Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.deviceStats.map((device) => (
                <div key={device.device} className="flex items-center justify-between">
                  <span className="text-sm" style={{ color: 'var(--text-dim)' }}>{device.device}</span>
                  <span className="text-sm font-medium font-mono tabular-nums" style={{ color: 'var(--text)' }}>{device.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
              <BarChart3 className="w-5 h-5" style={{ color: 'var(--status-ok)' }} />
              Browser Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {analytics.browserStats.map((browser) => (
                <div key={browser.browser} className="flex items-center justify-between">
                  <span className="text-sm" style={{ color: 'var(--text-dim)' }}>{browser.browser}</span>
                  <span className="text-sm font-medium font-mono tabular-nums" style={{ color: 'var(--text)' }}>{browser.percentage}%</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  );
}
