"use client";

import { useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  XAxis,
  YAxis,
} from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Invoice, AnalyticsPeriod } from "@/lib/api/types";
import type { CurrencyCode } from "@/lib/currency";
import {
  format,
  parseISO,
  subDays,
  subMonths,
  startOfMonth,
  eachDayOfInterval,
  eachMonthOfInterval,
  isSameDay,
  isSameMonth,
} from "date-fns";

interface SpendingTrendChartProps {
  invoices: Invoice[];
  defaultPeriod?: AnalyticsPeriod;
  displayCurrency?: CurrencyCode | null;
  exchangeRate?: number | null;
}

const periods: { value: AnalyticsPeriod; label: string }[] = [
  { value: "7d", label: "7 Days" },
  { value: "1m", label: "30 Days" },
  { value: "1y", label: "12 Months" },
];

export function SpendingTrendChart({
  invoices,
  defaultPeriod = "1m",
  displayCurrency,
  exchangeRate,
}: SpendingTrendChartProps) {
  const [period, setPeriod] = useState<AnalyticsPeriod>(defaultPeriod);

  // Helper to convert amount using exchange rate
  const convertAmount = (amount: number) => {
    if (displayCurrency && exchangeRate) {
      return amount * exchangeRate;
    }
    return amount;
  };

  const chartData = useMemo(() => {
    const now = new Date();

    // Helper to get invoice date (due_date with created_at fallback)
    const getInvoiceDate = (inv: Invoice) => parseISO(inv.due_date || inv.created_at);

    // Use target_amount (USD normalized) for calculations
    const getAmount = (inv: Invoice) => inv.target_amount || inv.amount;

    if (period === "7d") {
      // Last 7 days - daily breakdown
      const start = subDays(now, 6);
      const days = eachDayOfInterval({ start, end: now });

      return days.map((day) => {
        const dayInvoices = invoices.filter((inv) => {
          const invDate = getInvoiceDate(inv);
          return isSameDay(invDate, day);
        });

        const total = convertAmount(dayInvoices.reduce((sum, inv) => sum + getAmount(inv), 0));
        const paid = convertAmount(dayInvoices
          .filter((inv) => inv.status === "paid")
          .reduce((sum, inv) => sum + getAmount(inv), 0));

        return {
          date: format(day, "MMM d"),
          total,
          paid,
        };
      });
    } else if (period === "1m") {
      // Last month - daily breakdown (last 30 days)
      const start = subDays(now, 29);
      const days = eachDayOfInterval({ start, end: now });

      return days.map((day) => {
        const dayInvoices = invoices.filter((inv) => {
          const invDate = getInvoiceDate(inv);
          return isSameDay(invDate, day);
        });

        const total = convertAmount(dayInvoices.reduce((sum, inv) => sum + getAmount(inv), 0));
        const paid = convertAmount(dayInvoices
          .filter((inv) => inv.status === "paid")
          .reduce((sum, inv) => sum + getAmount(inv), 0));

        return {
          date: format(day, "MMM d"),
          total,
          paid,
        };
      });
    } else {
      // Last year - monthly breakdown
      const start = startOfMonth(subMonths(now, 11));
      const months = eachMonthOfInterval({ start, end: now });

      return months.map((month) => {
        const monthInvoices = invoices.filter((inv) => {
          const invDate = getInvoiceDate(inv);
          return isSameMonth(invDate, month);
        });

        const total = convertAmount(monthInvoices.reduce((sum, inv) => sum + getAmount(inv), 0));
        const paid = convertAmount(monthInvoices
          .filter((inv) => inv.status === "paid")
          .reduce((sum, inv) => sum + getAmount(inv), 0));

        return {
          date: format(month, "MMM yyyy"),
          total,
          paid,
        };
      });
    }
  }, [invoices, period, displayCurrency, exchangeRate]);

  const chartConfig = {
    total: {
      label: "Total",
      color: "hsl(var(--chart-1))",
    },
    paid: {
      label: "Paid",
      color: "hsl(var(--chart-2))",
    },
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle>Spending Trend</CardTitle>
          <CardDescription>
            Total and paid invoice amounts
          </CardDescription>
        </div>
        <Select value={period} onValueChange={(v) => setPeriod(v as AnalyticsPeriod)}>
          <SelectTrigger className="w-[120px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {periods.map((p) => (
              <SelectItem key={p.value} value={p.value}>
                {p.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[300px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillTotal" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-total)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-total)"
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient id="fillPaid" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--color-paid)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--color-paid)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              interval="preserveStartEnd"
              tick={{ fontSize: 12 }}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                const symbol = displayCurrency === "EUR" ? "€" : displayCurrency === "GBP" ? "£" : displayCurrency === "JPY" ? "¥" : "$";
                return `${symbol}${value}`;
              }}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent indicator="dot" />}
            />
            <Area
              dataKey="paid"
              type="monotone"
              fill="url(#fillPaid)"
              stroke="var(--color-paid)"
              stackId="a"
            />
            <Area
              dataKey="total"
              type="monotone"
              fill="url(#fillTotal)"
              stroke="var(--color-total)"
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
