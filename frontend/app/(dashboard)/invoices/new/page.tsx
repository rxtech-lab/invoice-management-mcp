import { InvoiceForm } from "@/components/forms/invoice-form";
import { getCategories } from "@/lib/api/categories";
import { getCompanies } from "@/lib/api/companies";

export default async function NewInvoicePage() {
  const [categoriesResponse, companiesResponse] = await Promise.all([
    getCategories({ limit: 100 }),
    getCompanies({ limit: 100 }),
  ]);

  return (
    <div className="max-w-2xl">
      <InvoiceForm
        categories={categoriesResponse.data || []}
        companies={companiesResponse.data || []}
      />
    </div>
  );
}
