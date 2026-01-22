"use client";

import { useMemo } from "react";
import {
  Bar,
  BarChart,
  XAxis,
  YAxis,
  CartesianGrid,
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
  ChartLegend,
  ChartLegendContent,
} from "@/components/ui/chart";
import type { AnalyticsByGroup } from "@/lib/api/types";

interface GroupBreakdownChartProps {
  data: AnalyticsByGroup;
  title: string;
  description: string;
  displayCurrency?: string;
}

export function GroupBreakdownChart({
  data,
  title,
  description,
  displayCurrency = "USD",
}: GroupBreakdownChartProps) {
  const chartData = useMemo(() => {
    const items = data.items.map((item) => ({
      name:
        item.name.length > 12 ? item.name.substring(0, 12) + "..." : item.name,
      paid: item.paid_amount,
      unpaid: item.unpaid_amount,
    }));

    if (data.uncategorized && data.uncategorized.total_amount > 0) {
      items.push({
        name: "Other",
        paid: data.uncategorized.paid_amount,
        unpaid: data.uncategorized.unpaid_amount,
      });
    }

    // Sort by total and take top 5
    return items
      .sort((a, b) => b.paid + b.unpaid - (a.paid + a.unpaid))
      .slice(0, 5);
  }, [data]);

  const chartConfig = {
    paid: { label: "Paid", color: "hsl(var(--chart-2))" },
    unpaid: { label: "Unpaid", color: "hsl(var(--chart-1))" },
  };

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent className="flex h-[300px] items-center justify-center">
          <p className="text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[300px] w-full">
          <BarChart data={chartData} layout="vertical">
            <CartesianGrid
              strokeDasharray="3 3"
              horizontal={true}
              vertical={false}
            />
            <XAxis
              type="number"
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => {
                const symbol = displayCurrency === "EUR" ? "€" : displayCurrency === "GBP" ? "£" : displayCurrency === "JPY" ? "¥" : "$";
                return `${symbol}${value}`;
              }}
            />
            <YAxis
              type="category"
              dataKey="name"
              tickLine={false}
              axisLine={false}
              width={80}
            />
            <ChartTooltip content={<ChartTooltipContent />} />
            <ChartLegend content={<ChartLegendContent />} />
            <Bar
              dataKey="paid"
              stackId="a"
              fill="var(--color-paid)"
              radius={[0, 0, 0, 0]}
            />
            <Bar
              dataKey="unpaid"
              stackId="a"
              fill="var(--color-unpaid)"
              radius={[4, 4, 0, 0]}
            />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
