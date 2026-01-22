import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AnalyticsSummary } from "@/lib/api/types";
import { formatCurrency } from "@/lib/utils";
import { DollarSign, FileText, AlertCircle, CheckCircle } from "lucide-react";

interface AnalyticsSummaryCardsProps {
  summary: AnalyticsSummary;
  displayCurrency?: string;
}

export function AnalyticsSummaryCards({ summary, displayCurrency = "USD" }: AnalyticsSummaryCardsProps) {
  const cards = [
    {
      title: "Total Invoices",
      value: summary.invoice_count.toString(),
      description: formatCurrency(summary.total_amount, displayCurrency),
      icon: FileText,
      className: "",
    },
    {
      title: "Paid",
      value: summary.paid_count.toString(),
      description: formatCurrency(summary.paid_amount, displayCurrency),
      icon: CheckCircle,
      className: "text-green-600",
    },
    {
      title: "Unpaid",
      value: summary.unpaid_count.toString(),
      description: formatCurrency(summary.unpaid_amount, displayCurrency),
      icon: DollarSign,
      className: "text-yellow-600",
    },
    {
      title: "Overdue",
      value: summary.overdue_count.toString(),
      description: formatCurrency(summary.overdue_amount, displayCurrency),
      icon: AlertCircle,
      className: "text-red-600",
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
            <card.icon
              className={`h-4 w-4 ${card.className || "text-muted-foreground"}`}
            />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{card.value}</div>
            <p className="text-xs text-muted-foreground">{card.description}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
