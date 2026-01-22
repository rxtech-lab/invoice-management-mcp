import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table/data-table";
import { tagColumns } from "@/components/data-table/columns/tag-columns";
import { getTags } from "@/lib/api/tags";

export default async function TagsPage() {
  const response = await getTags({ limit: 100 });
  const tags = response.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Tags</h1>
          <p className="text-muted-foreground">
            Organize your invoices with tags
          </p>
        </div>
        <Button asChild>
          <Link href="/tags/new">
            <Plus className="mr-2 h-4 w-4" />
            New Tag
          </Link>
        </Button>
      </div>
      <DataTable
        columns={tagColumns}
        data={tags}
        searchKey="name"
        searchPlaceholder="Search tags..."
      />
    </div>
  );
}
