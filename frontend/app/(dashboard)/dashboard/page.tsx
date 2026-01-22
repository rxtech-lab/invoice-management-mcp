import { Suspense } from "react";
import {
  getAnalyticsSummary,
  getAnalyticsByCategory,
  getAnalyticsByCompany,
  getAnalyticsByReceiver,
  getAnalyticsByTag,
} from "@/lib/api/analytics";
import { getInvoices } from "@/lib/api/invoices";
import { PeriodSelector } from "@/components/dashboard/period-selector";
import { DashboardContent } from "@/components/dashboard/dashboard-content";
import type { AnalyticsPeriod } from "@/lib/api/types";
import { Skeleton } from "@/components/ui/skeleton";

function DashboardContentFallback() {
  return (
    <div className="space-y-6">
      {/* Currency Picker placeholder */}
      <div className="flex justify-end">
        <Skeleton className="h-9 w-[180px]" />
      </div>
      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-[120px]" />
        ))}
      </div>
      {/* Trend Chart */}
      <Skeleton className="h-[400px]" />
      {/* Breakdown Charts */}
      <div className="grid gap-6 md:grid-cols-2">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-[400px]" />
        ))}
      </div>
    </div>
  );
}

interface DashboardPageProps {
  searchParams: Promise<{ period?: string }>;
}

function PeriodSelectorFallback() {
  return <Skeleton className="h-9 w-[150px]" />;
}

export default async function DashboardPage({
  searchParams,
}: DashboardPageProps) {
  const params = await searchParams;
  const period = (params.period as AnalyticsPeriod) || "1m";

  // Fetch all analytics data in parallel
  const [summary, byCategory, byCompany, byReceiver, byTag, invoicesResponse] = await Promise.all([
    getAnalyticsSummary(period),
    getAnalyticsByCategory(period),
    getAnalyticsByCompany(period),
    getAnalyticsByReceiver(period),
    getAnalyticsByTag(period),
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

      <Suspense fallback={<DashboardContentFallback />}>
        <DashboardContent
          summary={summary}
          byCategory={byCategory}
          byCompany={byCompany}
          byReceiver={byReceiver}
          byTag={byTag}
          invoices={invoices}
          period={period}
        />
      </Suspense>
    </div>
  );
}
