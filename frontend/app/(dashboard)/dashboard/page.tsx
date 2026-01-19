import { Suspense } from "react";
import {
  getAnalyticsSummary,
  getAnalyticsByCategory,
  getAnalyticsByCompany,
  getAnalyticsByReceiver,
} from "@/lib/api/analytics";
import { getInvoices } from "@/lib/api/invoices";
import { PeriodSelector } from "@/components/dashboard/period-selector";
import { AnalyticsSummaryCards } from "@/components/dashboard/analytics-summary-cards";
import { CategoryBreakdownChart } from "@/components/dashboard/category-breakdown-chart";
import { GroupBreakdownChart } from "@/components/dashboard/group-breakdown-chart";
import { SpendingTrendChart } from "@/components/dashboard/spending-trend-chart";
import type { AnalyticsPeriod } from "@/lib/api/types";
import { Skeleton } from "@/components/ui/skeleton";

interface DashboardPageProps {
  searchParams: Promise<{ period?: string }>;
}

function PeriodSelectorFallback() {
  return <Skeleton className="h-9 w-[150px]" />;
}

function SummaryCardsFallback() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {[1, 2, 3, 4].map((i) => (
        <Skeleton key={i} className="h-[120px]" />
      ))}
    </div>
  );
}

function ChartFallback() {
  return <Skeleton className="h-[400px]" />;
}

export default async function DashboardPage({
  searchParams,
}: DashboardPageProps) {
  const params = await searchParams;
  const period = (params.period as AnalyticsPeriod) || "1m";

  // Fetch all analytics data in parallel
  const [summary, byCategory, byCompany, byReceiver, invoicesResponse] = await Promise.all([
    getAnalyticsSummary(period),
    getAnalyticsByCategory(period),
    getAnalyticsByCompany(period),
    getAnalyticsByReceiver(period),
    getInvoices({ limit: 1000 }),
  ]);

  const invoices = invoicesResponse.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">
            Overview of your invoice management
          </p>
        </div>
        <Suspense fallback={<PeriodSelectorFallback />}>
          <PeriodSelector />
        </Suspense>
      </div>

      {/* Summary Cards */}
      <Suspense fallback={<SummaryCardsFallback />}>
        <AnalyticsSummaryCards summary={summary} />
      </Suspense>

      {/* Spending Trend Chart */}
      <Suspense fallback={<ChartFallback />}>
        <SpendingTrendChart invoices={invoices} defaultPeriod={period} />
      </Suspense>

      {/* Breakdown Charts Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Suspense fallback={<ChartFallback />}>
          <CategoryBreakdownChart data={byCategory} />
        </Suspense>
        <Suspense fallback={<ChartFallback />}>
          <GroupBreakdownChart
            data={byCompany}
            title="By Company"
            description="Invoice amounts grouped by company"
          />
        </Suspense>
        <Suspense fallback={<ChartFallback />}>
          <GroupBreakdownChart
            data={byReceiver}
            title="By Receiver"
            description="Invoice amounts grouped by receiver"
          />
        </Suspense>
      </div>
    </div>
  );
}
