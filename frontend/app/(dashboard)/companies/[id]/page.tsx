import { notFound } from "next/navigation";
import { CompanyForm } from "@/components/forms/company-form";
import { getCompany } from "@/lib/api/companies";

interface CompanyDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function CompanyDetailPage({
  params,
}: CompanyDetailPageProps) {
  const { id } = await params;
  const companyId = parseInt(id, 10);

  if (isNaN(companyId)) {
    notFound();
  }

  try {
    const company = await getCompany(companyId);

    return (
      <div className="max-w-2xl">
        <CompanyForm company={company} />
      </div>
    );
  } catch {
    notFound();
  }
}
