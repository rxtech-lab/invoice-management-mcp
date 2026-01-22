"use client";

import { DataTable } from "@/components/data-table/data-table";
import { createReceiverColumns } from "@/components/data-table/columns/receiver-columns";
import type { Receiver } from "@/lib/api/types";

interface ReceiversTableProps {
  receivers: Receiver[];
}

export function ReceiversTable({ receivers }: ReceiversTableProps) {
  return (
    <DataTable
      columns={createReceiverColumns(receivers)}
      data={receivers}
      searchKey="name"
      searchPlaceholder="Search receivers..."
    />
  );
}
