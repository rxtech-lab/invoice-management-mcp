"use client";

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  getFilteredRowModel,
  ColumnFiltersState,
} from "@tanstack/react-table";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ChevronLeft, ChevronRight, Search } from "lucide-react";

interface ServerPaginationProps {
  /** Total number of rows across all pages */
  total: number;
  /** Current page number (1-based) */
  currentPage: number;
  /** Number of rows per page */
  pageSize: number;
  /** Base path for pagination links (e.g. "/invoices") */
  basePath: string;
}

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  searchPlaceholder?: string;
  /** When provided, enables server-side pagination instead of client-side */
  serverPagination?: ServerPaginationProps;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  searchPlaceholder = "Search...",
  serverPagination,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const router = useRouter();
  const searchParams = useSearchParams();

  const isServerPaginated = !!serverPagination;
  const serverPageCount = serverPagination
    ? Math.max(1, Math.ceil(serverPagination.total / Math.max(1, serverPagination.pageSize)))
    : undefined;

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    ...(isServerPaginated ? {} : { getPaginationRowModel: getPaginationRowModel() }),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    state: {
      sorting,
      columnFilters,
      ...(isServerPaginated
        ? {
            pagination: {
              pageIndex: serverPagination!.currentPage - 1,
              pageSize: serverPagination!.pageSize,
            },
          }
        : {}),
    },
    ...(isServerPaginated
      ? {
          manualPagination: true,
          pageCount: serverPageCount,
        }
      : {
          initialState: {
            pagination: {
              pageSize: 10,
            },
          },
        }),
  });

  const navigateToPage = (page: number) => {
    if (!serverPagination) return;
    const params = new URLSearchParams(searchParams.toString());
    if (page <= 1) {
      params.delete("page");
    } else {
      params.set("page", String(page));
    }
    const query = params.toString();
    router.push(`${serverPagination.basePath}${query ? `?${query}` : ""}`);
  };

  const currentPage = serverPagination?.currentPage ?? table.getState().pagination.pageIndex + 1;
  const totalPages = serverPageCount ?? table.getPageCount();
  const canPreviousPage = isServerPaginated
    ? currentPage > 1
    : table.getCanPreviousPage();
  const canNextPage = isServerPaginated
    ? currentPage < totalPages
    : table.getCanNextPage();
  const totalRows = isServerPaginated
    ? serverPagination!.total
    : table.getFilteredRowModel().rows.length;

  return (
    <div className="space-y-4">
      {searchKey && (
        <div className="flex items-center gap-2">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={searchPlaceholder}
              value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
              onChange={(event) =>
                table.getColumn(searchKey)?.setFilterValue(event.target.value)
              }
              className="pl-9"
            />
          </div>
        </div>
      )}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between px-2">
        <div className="text-sm text-muted-foreground">
          {totalRows} row(s) total
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              isServerPaginated
                ? navigateToPage(currentPage - 1)
                : table.previousPage()
            }
            disabled={!canPreviousPage}
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          <div className="text-sm text-muted-foreground">
            Page {currentPage} of {totalPages}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              isServerPaginated
                ? navigateToPage(currentPage + 1)
                : table.nextPage()
            }
            disabled={!canNextPage}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
