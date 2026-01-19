import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table/data-table";
import { categoryColumns } from "@/components/data-table/columns/category-columns";
import { getCategories } from "@/lib/api/categories";

export default async function CategoriesPage() {
  const response = await getCategories({ limit: 100 });
  const categories = response.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Categories</h1>
          <p className="text-muted-foreground">
            Organize your invoices with categories
          </p>
        </div>
        <Button asChild>
          <Link href="/categories/new">
            <Plus className="mr-2 h-4 w-4" />
            New Category
          </Link>
        </Button>
      </div>
      <DataTable
        columns={categoryColumns}
        data={categories}
        searchKey="name"
        searchPlaceholder="Search categories..."
      />
    </div>
  );
}
