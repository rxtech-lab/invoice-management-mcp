import { notFound } from "next/navigation";
import { InvoiceForm } from "@/components/forms/invoice-form";
import { InvoiceItemsTable } from "@/components/forms/invoice-items-table";
import { getInvoice } from "@/lib/api/invoices";
import { getCategories } from "@/lib/api/categories";
import { getCompanies } from "@/lib/api/companies";
import { getReceivers } from "@/lib/api/receivers";
import { getTags } from "@/lib/api/tags";

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

  let invoice;
  let categoriesResponse;
  let companiesResponse;
  let receiversResponse;
  let tagsResponse;

  try {
    [
      invoice,
      categoriesResponse,
      companiesResponse,
      receiversResponse,
      tagsResponse,
    ] = await Promise.all([
      getInvoice(invoiceId),
      getCategories({ limit: 100 }),
      getCompanies({ limit: 100 }),
      getReceivers({ limit: 100 }),
      getTags({ limit: 20 }), // Reduced limit - tags are now fetched server-side on search
    ]);
  } catch {
    notFound();
  }

  return (
    <div className="max-w-4xl space-y-6">
      <InvoiceForm
        invoice={invoice}
        categories={categoriesResponse.data || []}
        companies={companiesResponse.data || []}
        receivers={receiversResponse.data || []}
        tags={tagsResponse.data || []}
      />

      <InvoiceItemsTable
        invoiceId={invoice.id}
        items={invoice.items || []}
        currency={invoice.currency}
      />
    </div>
  );
}
