"use client";

import { useMemo } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  XAxis,
  YAxis,
  ResponsiveContainer,
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
import { Invoice } from "@/lib/api/types";
import { format, parseISO, startOfMonth, subMonths } from "date-fns";

interface ChartAreaInteractiveProps {
  invoices: Invoice[];
}

export function ChartAreaInteractive({ invoices }: ChartAreaInteractiveProps) {
  const chartData = useMemo(() => {
    // Get last 6 months
    const months: { month: Date; label: string }[] = [];
    for (let i = 5; i >= 0; i--) {
      const month = startOfMonth(subMonths(new Date(), i));
      months.push({
        month,
        label: format(month, "MMM yyyy"),
      });
    }

    // Aggregate invoices by month
    return months.map(({ month, label }) => {
      const monthInvoices = invoices.filter((inv) => {
        const invDate = parseISO(inv.created_at);
        return (
          invDate.getMonth() === month.getMonth() &&
          invDate.getFullYear() === month.getFullYear()
        );
      });

      const total = monthInvoices.reduce((sum, inv) => sum + inv.amount, 0);
      const paid = monthInvoices
        .filter((inv) => inv.status === "paid")
        .reduce((sum, inv) => sum + inv.amount, 0);

      return {
        month: label,
        total,
        paid,
      };
    });
  }, [invoices]);

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
      <CardHeader>
        <CardTitle>Invoice Trends</CardTitle>
        <CardDescription>
          Total and paid invoice amounts over the last 6 months
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[300px] w-full">
          <ResponsiveContainer width="100%" height="100%">
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
                dataKey="month"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(value) => `$${value}`}
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
                stackId="b"
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
