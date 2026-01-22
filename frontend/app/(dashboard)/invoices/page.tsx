import { DataTable } from "@/components/data-table/data-table";
import { invoiceColumns } from "@/components/data-table/columns/invoice-columns";
import { getInvoices } from "@/lib/api/invoices";
import { getCategories } from "@/lib/api/categories";
import { getCompanies } from "@/lib/api/companies";
import { getReceivers } from "@/lib/api/receivers";
import { getTags } from "@/lib/api/tags";
import { NewInvoiceButton } from "@/components/invoices/new-invoice-button";
import { InvoiceFilters } from "@/components/invoices/invoice-filters";
import type { InvoiceStatus } from "@/lib/api/types";

interface Props {
  searchParams: Promise<{
    keyword?: string;
    category_id?: string;
    company_id?: string;
    receiver_id?: string;
    tag_ids?: string;
    status?: string;
  }>;
}

export default async function InvoicesPage({ searchParams }: Props) {
  const params = await searchParams;

  // Parse tag_ids from comma-separated string to number array
  const tagIds = params.tag_ids
    ? params.tag_ids.split(",").map(Number).filter(Boolean)
    : undefined;

  // Fetch filter options and invoices in parallel
  const [invoicesRes, categoriesRes, companiesRes, receiversRes, tagsRes] =
    await Promise.all([
      getInvoices({
        keyword: params.keyword,
        category_id: params.category_id
          ? parseInt(params.category_id)
          : undefined,
        company_id: params.company_id ? parseInt(params.company_id) : undefined,
        receiver_id: params.receiver_id
          ? parseInt(params.receiver_id)
          : undefined,
        tag_ids: tagIds,
        status: params.status as InvoiceStatus | undefined,
        limit: 100,
      }),
      getCategories({ limit: 100 }),
      getCompanies({ limit: 100 }),
      getReceivers({ limit: 100 }),
      getTags({ limit: 20 }), // Reduced limit - tags are now fetched server-side on search
    ]);

  const invoices = invoicesRes.data || [];
  const categories = categoriesRes.data || [];
  const companies = companiesRes.data || [];
  const receivers = receiversRes.data || [];
  const tags = tagsRes.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Invoices</h1>
          <p className="text-muted-foreground">
            Manage your invoices and track payments
          </p>
        </div>
        <NewInvoiceButton />
      </div>
      <InvoiceFilters
        categories={categories}
        companies={companies}
        receivers={receivers}
        tags={tags}
      />
      <DataTable columns={invoiceColumns} data={invoices} />
    </div>
  );
}
