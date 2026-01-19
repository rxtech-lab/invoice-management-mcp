import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table/data-table";
import { invoiceColumns } from "@/components/data-table/columns/invoice-columns";
import { getInvoices } from "@/lib/api/invoices";

export default async function InvoicesPage() {
  const response = await getInvoices({ limit: 100 });
  const invoices = response.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Invoices</h1>
          <p className="text-muted-foreground">
            Manage your invoices and track payments
          </p>
        </div>
        <Button asChild>
          <Link href="/invoices/new">
            <Plus className="mr-2 h-4 w-4" />
            New Invoice
          </Link>
        </Button>
      </div>
      <DataTable
        columns={invoiceColumns}
        data={invoices}
        searchKey="title"
        searchPlaceholder="Search invoices..."
      />
    </div>
  );
}
