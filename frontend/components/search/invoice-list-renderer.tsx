"use client";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";
import type { DisplayInvoicesResult } from "@/lib/search/types";
import type { InvoiceStatus } from "@/lib/api/types";
import { FileText, Building2, ExternalLink } from "lucide-react";

interface InvoiceListRendererProps {
  output: DisplayInvoicesResult;
  onAction?: (action: { type: string; path?: string; id?: number }) => void;
}

const statusVariants: Record<
  InvoiceStatus,
  "default" | "secondary" | "destructive"
> = {
  paid: "default",
  unpaid: "secondary",
  overdue: "destructive",
};

export function InvoiceListRenderer({
  output,
  onAction,
}: InvoiceListRendererProps) {
  const { invoices, total, query } = output;

  if (invoices.length === 0) {
    return (
      <Card className="my-3 max-w-md">
        <CardContent className="pt-4">
          <div className="text-center text-muted-foreground">
            <FileText className="h-6 w-6 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No invoices found{query ? ` for "${query}"` : ""}</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="my-3 max-w-md">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm flex items-center gap-2">
          <FileText className="h-3 w-3" />
          {total === invoices.length
            ? `${total} invoice${total !== 1 ? "s" : ""}`
            : `${invoices.length} of ${total} invoices`}
          {query && <span className="text-muted-foreground font-normal">for &quot;{query}&quot;</span>}
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-1.5">
          {invoices.map((invoice) => (
            <button
              key={invoice.id}
              onClick={() =>
                onAction?.({ type: "view_invoice", id: invoice.id })
              }
              className="w-full flex items-start gap-2 p-2 text-left hover:bg-accent rounded transition-colors border group"
            >
              <div className="flex-1 min-w-0 space-y-0.5">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-sm font-medium truncate flex items-center gap-1">
                    {invoice.title}
                    <ExternalLink className="h-2.5 w-2.5 opacity-0 group-hover:opacity-100" />
                  </span>
                  <div className="flex items-center gap-1.5 flex-shrink-0">
                    <span className="text-sm font-semibold">
                      {formatCurrency(invoice.amount, invoice.currency)}
                    </span>
                    <Badge variant={statusVariants[invoice.status]} className="text-xs px-1.5 py-0">
                      {invoice.status}
                    </Badge>
                  </div>
                </div>

                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  {invoice.company && (
                    <span className="flex items-center gap-0.5 truncate">
                      <Building2 className="h-2.5 w-2.5" />
                      {invoice.company.name}
                    </span>
                  )}
                  {invoice.category && (
                    <span className="flex items-center gap-0.5 truncate">
                      {invoice.category.color && (
                        <span
                          className="h-1.5 w-1.5 rounded-full"
                          style={{ backgroundColor: invoice.category.color }}
                        />
                      )}
                      {invoice.category.name}
                    </span>
                  )}
                </div>
              </div>
            </button>
          ))}
        </div>

        {total > invoices.length && (
          <button
            onClick={() =>
              onAction?.({
                type: "navigate",
                path: `/invoices?keyword=${encodeURIComponent(query || "")}`,
              })
            }
            className="w-full mt-2 p-1.5 text-xs text-primary hover:underline text-center"
          >
            View all {total} results
          </button>
        )}
      </CardContent>
    </Card>
  );
}
