"use client";

import { useMemo } from "react";
import { Pie, PieChart, Cell } from "recharts";
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
import { formatCurrency } from "@/lib/utils";

interface CategoryBreakdownChartProps {
  data: AnalyticsByGroup;
  displayCurrency?: string;
}

const DEFAULT_COLORS = [
  "hsl(var(--chart-1))",
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
];

export function CategoryBreakdownChart({ data, displayCurrency = "USD" }: CategoryBreakdownChartProps) {
  const { chartData, chartConfig } = useMemo(() => {
    const items = data.items.map((item, index) => ({
      name: item.name,
      value: item.total_amount,
      fill: item.color || DEFAULT_COLORS[index % DEFAULT_COLORS.length],
    }));

    if (data.uncategorized && data.uncategorized.total_amount > 0) {
      items.push({
        name: "Uncategorized",
        value: data.uncategorized.total_amount,
        fill: "hsl(var(--muted))",
      });
    }

    const config = items.reduce(
      (acc, item) => {
        acc[item.name] = { label: item.name, color: item.fill };
        return acc;
      },
      {} as Record<string, { label: string; color: string }>
    );

    return { chartData: items, chartConfig: config };
  }, [data]);

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>By Category</CardTitle>
          <CardDescription>Invoice amounts grouped by category</CardDescription>
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
        <CardTitle>By Category</CardTitle>
        <CardDescription>Invoice amounts grouped by category</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="mx-auto h-[300px]">
          <PieChart>
            <ChartTooltip
              content={
                <ChartTooltipContent
                  formatter={(value) => formatCurrency(value as number, displayCurrency)}
                />
              }
            />
            <Pie
              data={chartData}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={100}
              paddingAngle={2}
            >
              {chartData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.fill} />
              ))}
            </Pie>
            <ChartLegend content={<ChartLegendContent nameKey="name" />} />
          </PieChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
