/**
 * Reusable Chart Components
 *
 * Pre-configured chart components using Recharts with FunctionFly design system.
 * All charts follow the same API patterns for consistency.
 *
 * @example
 * ```tsx
 * import { LineChart, BarChart, AreaChart, PieChart } from '@/components/ui';
 *
 * <LineChart
 *   data={data}
 *   series={[{ key: 'value', name: 'Revenue', color: '#6366f1' }]}
 *   title="Monthly Revenue"
 * />
 * ```
 */

export { LineChart, Sparkline } from './chart-line';
export type { LineChartProps, LineSeries, LineChartData } from './chart-line';

export { BarChart, SimpleBarChart } from './chart-bar';
export type { BarChartProps, BarSeries, BarChartData } from './chart-bar';

export { AreaChart, SparkAreaChart } from './chart-area';
export type { AreaChartProps, AreaSeries, AreaChartData } from './chart-area';

export { PieChart } from './chart-pie';
export type { PieChartProps, PieChartData } from './chart-pie';
