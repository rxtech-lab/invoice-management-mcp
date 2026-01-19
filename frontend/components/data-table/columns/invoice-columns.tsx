"use client";

import { ColumnDef } from "@tanstack/react-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { MoreHorizontal, Pencil, Trash, Eye, ArrowUpDown } from "lucide-react";
import { Invoice, InvoiceStatus } from "@/lib/api/types";
import { formatCurrency, formatDate } from "@/lib/utils";
import Link from "next/link";
import { deleteInvoiceAction, updateInvoiceStatusAction } from "@/lib/actions/invoice-actions";
import { toast } from "sonner";

const statusVariants: Record<InvoiceStatus, "default" | "secondary" | "destructive"> = {
  paid: "default",
  unpaid: "secondary",
  overdue: "destructive",
};

export const invoiceColumns: ColumnDef<Invoice>[] = [
  {
    accessorKey: "title",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-4"
      >
        Title
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <Link
        href={`/invoices/${row.original.id}`}
        className="font-medium hover:underline"
      >
        {row.original.title}
      </Link>
    ),
  },
  {
    accessorKey: "company",
    header: "Company",
    cell: ({ row }) => row.original.company?.name || "-",
  },
  {
    accessorKey: "category",
    header: "Category",
    cell: ({ row }) => {
      const category = row.original.category;
      if (!category) return "-";
      return (
        <div className="flex items-center gap-2">
          {category.color && (
            <div
              className="h-3 w-3 rounded-full"
              style={{ backgroundColor: category.color }}
            />
          )}
          {category.name}
        </div>
      );
    },
  },
  {
    accessorKey: "amount",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-4"
      >
        Amount
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) =>
      formatCurrency(row.original.amount, row.original.currency),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const status = row.original.status;
      return (
        <Badge variant={statusVariants[status]}>
          {status.charAt(0).toUpperCase() + status.slice(1)}
        </Badge>
      );
    },
  },
  {
    accessorKey: "due_date",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-4"
      >
        Due Date
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => formatDate(row.original.due_date),
  },
  {
    id: "actions",
    cell: ({ row }) => {
      const invoice = row.original;

      const handleStatusChange = async (status: InvoiceStatus) => {
        const result = await updateInvoiceStatusAction(invoice.id, status);
        if (result.success) {
          toast.success(`Invoice marked as ${status}`);
        } else {
          toast.error(result.error || "Failed to update status");
        }
      };

      const handleDelete = async () => {
        if (!confirm("Are you sure you want to delete this invoice?")) return;
        const result = await deleteInvoiceAction(invoice.id);
        if (result.success) {
          toast.success("Invoice deleted");
        } else {
          toast.error(result.error || "Failed to delete invoice");
        }
      };

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem asChild>
              <Link href={`/invoices/${invoice.id}`}>
                <Eye className="mr-2 h-4 w-4" />
                View
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href={`/invoices/${invoice.id}`}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuLabel>Change Status</DropdownMenuLabel>
            <DropdownMenuItem onClick={() => handleStatusChange("paid")}>
              Mark as Paid
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleStatusChange("unpaid")}>
              Mark as Unpaid
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleStatusChange("overdue")}>
              Mark as Overdue
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleDelete}
              className="text-destructive"
            >
              <Trash className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];
