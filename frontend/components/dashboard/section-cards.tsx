import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Invoice } from "@/lib/api/types";
import { formatCurrency } from "@/lib/utils";
import { DollarSign, FileText, AlertCircle, CheckCircle } from "lucide-react";

interface SectionCardsProps {
  invoices: Invoice[];
}

export function SectionCards({ invoices }: SectionCardsProps) {
  const totalAmount = invoices.reduce((sum, inv) => sum + inv.amount, 0);
  const paidAmount = invoices
    .filter((inv) => inv.status === "paid")
    .reduce((sum, inv) => sum + inv.amount, 0);
  const paidCount = invoices.filter((inv) => inv.status === "paid").length;
  const overdueCount = invoices.filter((inv) => inv.status === "overdue").length;

  const cards = [
    {
      title: "Total Invoices",
      value: invoices.length.toString(),
      description: "All time invoices",
      icon: FileText,
    },
    {
      title: "Total Amount",
      value: formatCurrency(totalAmount),
      description: "Sum of all invoices",
      icon: DollarSign,
    },
    {
      title: "Paid",
      value: paidCount.toString(),
      description: formatCurrency(paidAmount),
      icon: CheckCircle,
    },
    {
      title: "Overdue",
      value: overdueCount.toString(),
      description: "Requires attention",
      icon: AlertCircle,
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
            <card.icon className="h-4 w-4 text-muted-foreground" />
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
