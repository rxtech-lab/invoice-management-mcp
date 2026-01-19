import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table/data-table";
import { receiverColumns } from "@/components/data-table/columns/receiver-columns";
import { getReceivers } from "@/lib/api/receivers";

export default async function ReceiversPage() {
  const response = await getReceivers({ limit: 100 });
  const receivers = response.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Receivers</h1>
          <p className="text-muted-foreground">
            Manage invoice receivers (individuals and organizations)
          </p>
        </div>
        <Button asChild>
          <Link href="/receivers/new">
            <Plus className="mr-2 h-4 w-4" />
            New Receiver
          </Link>
        </Button>
      </div>
      <DataTable
        columns={receiverColumns}
        data={receivers}
        searchKey="name"
        searchPlaceholder="Search receivers..."
      />
    </div>
  );
}
