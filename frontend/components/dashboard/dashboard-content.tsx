"use client";

import { useState } from "react";
import { CurrencyPicker } from "@/components/currency/currency-picker";
import { AnalyticsSummaryCards } from "@/components/dashboard/analytics-summary-cards";
import { CategoryBreakdownChart } from "@/components/dashboard/category-breakdown-chart";
import { TagBreakdownChart } from "@/components/dashboard/tag-breakdown-chart";
import { GroupBreakdownChart } from "@/components/dashboard/group-breakdown-chart";
import { SpendingTrendChart } from "@/components/dashboard/spending-trend-chart";
import { useExchangeRate } from "@/hooks/use-exchange-rate";
import type {
  AnalyticsSummary,
  AnalyticsByGroup,
  Invoice,
  AnalyticsPeriod,
} from "@/lib/api/types";
import type { CurrencyCode } from "@/lib/currency";

interface DashboardContentProps {
  summary: AnalyticsSummary;
  byCategory: AnalyticsByGroup;
  byCompany: AnalyticsByGroup;
  byReceiver: AnalyticsByGroup;
  byTag: AnalyticsByGroup;
  invoices: Invoice[];
  period: AnalyticsPeriod;
}

export function DashboardContent({
  summary,
  byCategory,
  byCompany,
  byReceiver,
  byTag,
  invoices,
  period,
}: DashboardContentProps) {
  const [displayCurrency, setDisplayCurrency] = useState<CurrencyCode | null>(
    null
  );

  // All amounts from backend are USD-normalized, so we convert from USD to display currency
  const { rate, isLoading: rateLoading, convert } = useExchangeRate({
    fromCurrency: "USD",
    toCurrency: displayCurrency,
  });

  // Convert summary amounts if a display currency is selected
  const convertedSummary: AnalyticsSummary = displayCurrency && rate
    ? {
        ...summary,
        total_amount: convert(summary.total_amount) ?? summary.total_amount,
        paid_amount: convert(summary.paid_amount) ?? summary.paid_amount,
        unpaid_amount: convert(summary.unpaid_amount) ?? summary.unpaid_amount,
        overdue_amount: convert(summary.overdue_amount) ?? summary.overdue_amount,
      }
    : summary;

  // Helper to convert analytics group data
  const convertGroupData = (data: AnalyticsByGroup): AnalyticsByGroup => {
    if (!displayCurrency || !rate) return data;
    return {
      ...data,
      items: data.items.map((item) => ({
        ...item,
        total_amount: convert(item.total_amount) ?? item.total_amount,
        paid_amount: convert(item.paid_amount) ?? item.paid_amount,
        unpaid_amount: convert(item.unpaid_amount) ?? item.unpaid_amount,
      })),
      uncategorized: data.uncategorized
        ? {
            ...data.uncategorized,
            total_amount: convert(data.uncategorized.total_amount) ?? data.uncategorized.total_amount,
            paid_amount: convert(data.uncategorized.paid_amount) ?? data.uncategorized.paid_amount,
            unpaid_amount: convert(data.uncategorized.unpaid_amount) ?? data.uncategorized.unpaid_amount,
          }
        : undefined,
    };
  };

  return (
    <div className="space-y-6">
      {/* Currency Picker */}
      <div className="flex justify-end">
        <CurrencyPicker
          value={displayCurrency}
          onChange={setDisplayCurrency}
          originalCurrency="USD"
          isLoading={rateLoading}
        />
      </div>

      {/* Summary Cards */}
      <AnalyticsSummaryCards
        summary={convertedSummary}
        displayCurrency={displayCurrency || "USD"}
      />

      {/* Spending Trend Chart */}
      <SpendingTrendChart
        invoices={invoices}
        defaultPeriod={period}
        displayCurrency={displayCurrency}
        exchangeRate={rate?.rate}
      />

      {/* Breakdown Charts Grid */}
      <div className="grid gap-6 md:grid-cols-2">
        <CategoryBreakdownChart
          data={convertGroupData(byCategory)}
          displayCurrency={displayCurrency || "USD"}
        />
        <TagBreakdownChart
          data={convertGroupData(byTag)}
          displayCurrency={displayCurrency || "USD"}
        />
        <GroupBreakdownChart
          data={convertGroupData(byCompany)}
          title="By Company"
          description="Invoice amounts grouped by company"
          displayCurrency={displayCurrency || "USD"}
        />
        <GroupBreakdownChart
          data={convertGroupData(byReceiver)}
          title="By Receiver"
          description="Invoice amounts grouped by receiver"
          displayCurrency={displayCurrency || "USD"}
        />
      </div>
    </div>
  );
}
