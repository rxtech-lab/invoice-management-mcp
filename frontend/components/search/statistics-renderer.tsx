"use client";

import { useMemo } from "react";
import {
  Bar,
  BarChart,
  Pie,
  PieChart,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from "recharts";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { InvoiceStatistics, BreakdownItem } from "@/lib/search/types";
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  FileText,
  Calendar,
} from "lucide-react";

interface StatisticsRendererProps {
  output: InvoiceStatistics;
  onAction?: (action: { type: string; path?: string; id?: number }) => void;
}

// Color palette for charts
const COLORS = [
  "hsl(var(--chart-1))",
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
];

const STATUS_COLORS = {
  paid: "hsl(142.1 76.2% 36.3%)",
  unpaid: "hsl(var(--chart-4))",
  overdue: "hsl(var(--destructive))",
};

export function StatisticsRenderer({
  output,
  onAction,
}: StatisticsRendererProps) {
  const stats = output;

  // Normalize field names (backend may use different names)
  const invoiceCount = stats.invoice_count ?? stats.total_count ?? 0;
  const statusBreakdown = stats.by_status ?? stats.status_breakdown;
  const currency = stats.currency ?? "USD";

  // Determine chart type based on breakdown data
  const chartType = useMemo(() => {
    if (!stats.breakdown || stats.breakdown.length === 0) return "summary";
    // If breakdown has date field, it's time-based (bar chart)
    if (stats.breakdown[0]?.date) return "time";
    // Otherwise it's entity-based (pie chart)
    return "entity";
  }, [stats.breakdown]);

  // Format breakdown for time-based bar chart
  const timeChartData = useMemo(() => {
    if (chartType !== "time" || !stats.breakdown) return [];
    return stats.breakdown.map((item: BreakdownItem) => ({
      date: item.date,
      amount: item.amount,
      count: item.count,
      // Format date as short form (e.g., "Jan 14")
      label: item.date ? new Date(item.date).toLocaleDateString("en-US", { month: "short", day: "numeric" }) : "",
    }));
  }, [chartType, stats.breakdown]);

  // Format breakdown for entity-based pie chart
  const entityChartData = useMemo(() => {
    if (chartType !== "entity" || !stats.breakdown) return [];
    return stats.breakdown.map((item: BreakdownItem, index: number) => ({
      name: item.name || `Item ${index + 1}`,
      id: item.id,
      value: item.amount,
      count: item.count,
      fill: COLORS[index % COLORS.length],
    }));
  }, [chartType, stats.breakdown]);

  // Status breakdown pie chart data
  const statusChartData = useMemo(() => {
    if (!statusBreakdown) return [];
    return [
      {
        name: "Paid",
        value: statusBreakdown.paid?.amount ?? 0,
        count: statusBreakdown.paid?.count ?? 0,
        fill: STATUS_COLORS.paid,
      },
      {
        name: "Unpaid",
        value: statusBreakdown.unpaid?.amount ?? 0,
        count: statusBreakdown.unpaid?.count ?? 0,
        fill: STATUS_COLORS.unpaid,
      },
      {
        name: "Overdue",
        value: statusBreakdown.overdue?.amount ?? 0,
        count: statusBreakdown.overdue?.count ?? 0,
        fill: STATUS_COLORS.overdue,
      },
    ].filter((item) => item.value > 0 || item.count > 0);
  }, [statusBreakdown]);

  const chartConfig = {
    amount: { label: "Amount", color: "hsl(var(--chart-1))" },
    count: { label: "Count", color: "hsl(var(--chart-2))" },
  };

  return (
    <div className="space-y-3 my-3 max-w-md">
      {/* Summary Cards */}
      <div className="grid grid-cols-2 gap-2">
        <Card>
          <CardContent className="p-3">
            <div className="flex items-center gap-1">
              <DollarSign className="h-3 w-3 text-muted-foreground" />
              <span className="text-xs text-muted-foreground">Total</span>
            </div>
            <p className="text-base font-bold">
              {formatCurrency(stats.total_amount, currency)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-3">
            <div className="flex items-center gap-1">
              <FileText className="h-3 w-3 text-muted-foreground" />
              <span className="text-xs text-muted-foreground">Invoices</span>
            </div>
            <p className="text-base font-bold">{invoiceCount}</p>
          </CardContent>
        </Card>

        {stats.aggregations && (
          <>
            <Card>
              <CardContent className="p-3">
                <div className="flex items-center gap-1">
                  <TrendingUp className="h-3 w-3 text-green-500" />
                  <span className="text-xs text-muted-foreground">Max</span>
                </div>
                <p className="text-base font-bold">
                  {formatCurrency(stats.aggregations.max_amount)}
                </p>
                {stats.aggregations.max_invoice && (
                  <button
                    onClick={() =>
                      onAction?.({
                        type: "view_invoice",
                        id: stats.aggregations!.max_invoice!.id,
                      })
                    }
                    className="text-xs text-primary hover:underline truncate block"
                  >
                    {stats.aggregations.max_invoice.title}
                  </button>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-3">
                <div className="flex items-center gap-1">
                  <TrendingDown className="h-3 w-3 text-blue-500" />
                  <span className="text-xs text-muted-foreground">Avg</span>
                </div>
                <p className="text-base font-bold">
                  {formatCurrency(stats.aggregations.avg_amount)}
                </p>
              </CardContent>
            </Card>
          </>
        )}
      </div>

      {/* Period Info */}
      <div className="flex items-center gap-1 text-xs text-muted-foreground">
        <Calendar className="h-3 w-3" />
        <span>
          {stats.period}: {formatDate(stats.start_date)} - {formatDate(stats.end_date)}
        </span>
      </div>

      {/* Time-based Bar Chart */}
      {chartType === "time" && timeChartData.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Spending Over Time</CardTitle>
          </CardHeader>
          <CardContent>
            <ChartContainer config={chartConfig} className="h-[180px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={timeChartData} margin={{ top: 10, right: 10, left: 10, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} />
                  <XAxis
                    dataKey="label"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    fontSize={11}
                    interval={0}
                    angle={-45}
                    textAnchor="end"
                    height={50}
                  />
                  <YAxis
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    tickFormatter={(value) => `$${value}`}
                    fontSize={11}
                    width={50}
                  />
                  <Tooltip
                    cursor={{ fill: "hsl(var(--muted))", opacity: 0.3 }}
                    content={({ active, payload }) => {
                      if (!active || !payload?.length) return null;
                      const data = payload[0].payload;
                      return (
                        <div className="rounded-lg border bg-background p-2 shadow-md">
                          <p className="text-sm font-medium">{data.date}</p>
                          <p className="text-sm text-muted-foreground">
                            {formatCurrency(data.amount, currency)}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {data.count} invoice{data.count !== 1 ? "s" : ""}
                          </p>
                        </div>
                      );
                    }}
                  />
                  <Bar
                    dataKey="amount"
                    fill="hsl(var(--chart-1))"
                    radius={[4, 4, 0, 0]}
                    minPointSize={2}
                  />
                </BarChart>
              </ResponsiveContainer>
            </ChartContainer>
          </CardContent>
        </Card>
      )}

      {/* Entity-based Pie Chart */}
      {chartType === "entity" && entityChartData.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Spending Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[200px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={entityChartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={40}
                    outerRadius={70}
                    paddingAngle={2}
                    dataKey="value"
                    label={({ name, percent }) =>
                      `${name} (${(percent * 100).toFixed(0)}%)`
                    }
                    labelLine={false}
                  >
                    {entityChartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value: number) => formatCurrency(value)}
                  />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Status Breakdown Pie Chart */}
      {statusBreakdown && statusChartData.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Status Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[160px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={statusChartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={35}
                    outerRadius={60}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {statusChartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value: number, name: string) => [
                      formatCurrency(value),
                      name,
                    ]}
                  />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* Status Details */}
            <div className="grid grid-cols-3 gap-1 mt-2">
              {statusChartData.map((item) => (
                <button
                  key={item.name}
                  onClick={() =>
                    onAction?.({
                      type: "navigate",
                      path: `/invoices?status=${item.name.toLowerCase()}`,
                    })
                  }
                  className="p-1.5 rounded hover:bg-accent text-center"
                >
                  <div className="flex items-center justify-center gap-1 text-xs text-muted-foreground">
                    <span
                      className="h-1.5 w-1.5 rounded-full"
                      style={{ backgroundColor: item.fill }}
                    />
                    {item.name}
                  </div>
                  <p className="font-semibold text-xs">
                    {formatCurrency(item.value)}
                  </p>
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
