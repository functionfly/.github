import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, Activity, CreditCard, ExternalLink } from 'lucide-react';
import { Link } from 'react-router-dom';
import {
  ResponsiveContainer,
  RadialBarChart,
  RadialBar,
  PolarAngleAxis,
} from 'recharts';
import { formatCostUsd } from '@/api/usageAnalytics';
import { COLORS } from '../constants';
import { ROUTES } from '@/lib/constants';

interface ForecastTabProps {
  // Loading states
  forecastLoading: boolean;
  spendCapLoading: boolean;
  billingLoading: boolean;

  // Data
  forecast: {
    predicted_monthly_cost_usd?: number;
    confidence?: number;
  } | null | undefined;
  spendCap: {
    spend_cap_usd?: number;
  } | null | undefined;

  // Navigation
  username?: string;

  // Actions
  openBillingPortal: () => void;
}

export function ForecastTab({
  forecastLoading,
  spendCapLoading,
  billingLoading,
  forecast,
  spendCap,
  username,
  openBillingPortal,
}: ForecastTabProps) {
  const predictedCost = forecast?.predicted_monthly_cost_usd ?? 0;
  const capAmount = spendCap?.spend_cap_usd ?? 0;
  const utilizationPercent = capAmount > 0 ? Math.min((predictedCost / capAmount) * 100, 100) : 0;

  // Determine gauge color based on utilization
  const gaugeColor =
    predictedCost > capAmount && capAmount > 0
      ? COLORS.error
      : predictedCost > capAmount * 0.8 && capAmount > 0
        ? COLORS.cached
        : COLORS.success;

  return (
    <div className="space-y-4">
      {/* Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Projected Monthly Cost
            </CardTitle>
          </CardHeader>
          <CardContent>
            {forecastLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-xl font-semibold text-text-primary">
                {formatCostUsd(predictedCost)}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Confidence</CardTitle>
          </CardHeader>
          <CardContent>
            {forecastLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-xl font-semibold text-text-primary">
                {((forecast?.confidence ?? 0) * 100).toFixed(0)}%
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Spend Cap</CardTitle>
          </CardHeader>
          <CardContent>
            {spendCapLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-xl font-semibold text-text-primary">
                {capAmount > 0 ? formatCostUsd(capAmount) : 'Not set'}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Forecast vs Cap
            </CardTitle>
          </CardHeader>
          <CardContent>
            {forecastLoading || spendCapLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : predictedCost && capAmount > 0 ? (
              <p
                className={`text-xl font-semibold ${
                  predictedCost > capAmount ? 'text-red-500' : 'text-emerald-500'
                }`}
              >
                {predictedCost > capAmount ? 'Over Budget' : 'Within Budget'}
              </p>
            ) : (
              <p className="text-xl font-semibold text-text-muted">N/A</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Budget Gauge Chart */}
      {capAmount > 0 && (
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Budget Utilization Gauge
            </CardTitle>
            <CardDescription>
              Visual comparison of predicted cost against your spend cap
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[200px] flex items-center justify-center">
              {forecastLoading || spendCapLoading ? (
                <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <RadialBarChart
                    innerRadius="60%"
                    outerRadius="90%"
                    data={[
                      {
                        name: 'Predicted Cost',
                        value: utilizationPercent,
                        fill: gaugeColor,
                      },
                    ]}
                    startAngle={180}
                    endAngle={0}
                  >
                    <PolarAngleAxis type="number" domain={[0, 100]} angleAxisId={0} tick={false} />
                    <RadialBar background dataKey="value" cornerRadius={10} fill={gaugeColor} />
                    <text
                      x="50%"
                      y="60%"
                      textAnchor="middle"
                      dominantBaseline="middle"
                      className="fill-text-primary text-lg font-semibold"
                    >
                      {utilizationPercent.toFixed(0)}%
                    </text>
                    <text
                      x="50%"
                      y="75%"
                      textAnchor="middle"
                      dominantBaseline="middle"
                      className="fill-text-muted text-xs"
                    >
                      of ${capAmount} cap
                    </text>
                  </RadialBarChart>
                </ResponsiveContainer>
              )}
            </div>
            <div className="flex justify-center gap-6 mt-4">
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS.success }}
                />
                <span className="text-sm text-text-secondary">On Track (&lt;80%)</span>
              </div>
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS.cached }}
                />
                <span className="text-sm text-text-secondary">Caution (80-100%)</span>
              </div>
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: COLORS.error }}
                />
                <span className="text-sm text-text-secondary">Over Budget (&gt;100%)</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Forecast Explanation */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">About Forecasting</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-text-secondary">
            Cost forecasts are based on your historical usage patterns and current spending trends.
            The prediction includes all cost categories: execution fees, compute time, platform
            fees, and data transfer. Set a spend cap to receive alerts when your predicted
            monthly cost approaches your budget.
          </p>
          <div className="flex flex-wrap gap-3 mt-4">
            <Button
              variant="outline"
              size="sm"
              onClick={openBillingPortal}
              disabled={billingLoading}
            >
              <CreditCard className="h-4 w-4 mr-2" />
              Manage Spend Cap
            </Button>
            <Button variant="ghost" size="sm" asChild>
              <Link to={ROUTES.SETTINGS}>
                Settings <ExternalLink className="h-4 w-4 ml-2" />
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
