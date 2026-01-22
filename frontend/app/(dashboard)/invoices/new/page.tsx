import { InvoiceForm } from "@/components/forms/invoice-form";
import { getCategories } from "@/lib/api/categories";
import { getCompanies } from "@/lib/api/companies";
import { getReceivers } from "@/lib/api/receivers";
import { getTags } from "@/lib/api/tags";

export default async function NewInvoicePage() {
  const [categoriesResponse, companiesResponse, receiversResponse, tagsResponse] = await Promise.all([
    getCategories({ limit: 100 }),
    getCompanies({ limit: 100 }),
    getReceivers({ limit: 100 }),
    getTags({ limit: 100 }),
  ]);

  return (
    <div className="max-w-2xl">
      <InvoiceForm
        categories={categoriesResponse.data || []}
        companies={companiesResponse.data || []}
        receivers={receiversResponse.data || []}
        tags={tagsResponse.data || []}
      />
    </div>
  );
}
