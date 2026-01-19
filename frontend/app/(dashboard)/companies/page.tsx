import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table/data-table";
import { companyColumns } from "@/components/data-table/columns/company-columns";
import { getCompanies } from "@/lib/api/companies";

export default async function CompaniesPage() {
  const response = await getCompanies({ limit: 100 });
  const companies = response.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Companies</h1>
          <p className="text-muted-foreground">
            Manage your vendors and clients
          </p>
        </div>
        <Button asChild>
          <Link href="/companies/new">
            <Plus className="mr-2 h-4 w-4" />
            New Company
          </Link>
        </Button>
      </div>
      <DataTable
        columns={companyColumns}
        data={companies}
        searchKey="name"
        searchPlaceholder="Search companies..."
      />
    </div>
  );
}
