import { getInvoices } from "@/lib/api/invoices";
import { SectionCards } from "@/components/dashboard/section-cards";
import { ChartAreaInteractive } from "@/components/dashboard/chart-area-interactive";

export default async function DashboardPage() {
  const invoicesResponse = await getInvoices({ limit: 1000 });
  const invoices = invoicesResponse.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">
          Overview of your invoice management
        </p>
      </div>
      <SectionCards invoices={invoices} />
      <ChartAreaInteractive invoices={invoices} />
    </div>
  );
}
