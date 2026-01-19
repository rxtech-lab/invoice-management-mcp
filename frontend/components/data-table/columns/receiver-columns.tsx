"use client";

import { ColumnDef } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Pencil, Trash, ArrowUpDown, User, Building2 } from "lucide-react";
import { Receiver } from "@/lib/api/types";
import { formatDate } from "@/lib/utils";
import Link from "next/link";
import { deleteReceiverAction } from "@/lib/actions/receiver-actions";
import { toast } from "sonner";

export const receiverColumns: ColumnDef<Receiver>[] = [
  {
    accessorKey: "name",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-4"
      >
        Name
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => {
      const receiver = row.original;
      return (
        <Link
          href={`/receivers/${receiver.id}`}
          className="flex items-center gap-2 font-medium hover:underline"
        >
          {receiver.is_organization ? (
            <Building2 className="h-4 w-4 text-muted-foreground" />
          ) : (
            <User className="h-4 w-4 text-muted-foreground" />
          )}
          {receiver.name}
        </Link>
      );
    },
  },
  {
    accessorKey: "is_organization",
    header: "Type",
    cell: ({ row }) => {
      const isOrganization = row.original.is_organization;
      return (
        <Badge variant={isOrganization ? "default" : "secondary"}>
          {isOrganization ? "Organization" : "Individual"}
        </Badge>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        className="-ml-4"
      >
        Created
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => formatDate(row.original.created_at),
  },
  {
    id: "actions",
    cell: ({ row }) => {
      const receiver = row.original;

      const handleDelete = async () => {
        if (!confirm("Are you sure you want to delete this receiver?")) return;
        const result = await deleteReceiverAction(receiver.id);
        if (result.success) {
          toast.success("Receiver deleted");
        } else {
          toast.error(result.error || "Failed to delete receiver");
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
              <Link href={`/receivers/${receiver.id}`}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </Link>
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
