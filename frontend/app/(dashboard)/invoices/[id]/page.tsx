import { notFound } from "next/navigation";
import { InvoiceForm } from "@/components/forms/invoice-form";
import { InvoiceItemsTable } from "@/components/forms/invoice-items-table";
import { getInvoice } from "@/lib/api/invoices";
import { getCategories } from "@/lib/api/categories";
import { getCompanies } from "@/lib/api/companies";

interface InvoiceDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function InvoiceDetailPage({
  params,
}: InvoiceDetailPageProps) {
  const { id } = await params;
  const invoiceId = parseInt(id, 10);

  if (isNaN(invoiceId)) {
    notFound();
  }

  try {
    const [invoice, categoriesResponse, companiesResponse] = await Promise.all([
      getInvoice(invoiceId),
      getCategories({ limit: 100 }),
      getCompanies({ limit: 100 }),
    ]);

    return (
      <div className="space-y-6">
        <div className="max-w-2xl">
          <InvoiceForm
            invoice={invoice}
            categories={categoriesResponse.data || []}
            companies={companiesResponse.data || []}
          />
        </div>
        <InvoiceItemsTable
          invoiceId={invoice.id}
          items={invoice.items || []}
          currency={invoice.currency}
        />
      </div>
    );
  } catch {
    notFound();
  }
}
