"use client";

import { Badge } from "@/components/ui/badge";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { InvoiceSearchResult } from "@/lib/search/types";
import type { InvoiceStatus } from "@/lib/api/types";
import { FileText, Building2, Tag, User } from "lucide-react";

interface InvoiceResultProps {
  invoice: InvoiceSearchResult;
  onSelect: () => void;
}

const statusVariants: Record<
  InvoiceStatus,
  "default" | "secondary" | "destructive"
> = {
  paid: "default",
  unpaid: "secondary",
  overdue: "destructive",
};

export function InvoiceResult({ invoice, onSelect }: InvoiceResultProps) {
  return (
    <button
      onClick={onSelect}
      className="w-full flex items-start gap-3 p-3 text-left hover:bg-accent rounded-lg transition-colors"
    >
      <div className="flex-shrink-0 mt-0.5">
        <FileText className="h-5 w-5 text-muted-foreground" />
      </div>
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex items-center justify-between gap-2">
          <span className="font-medium truncate">{invoice.title}</span>
          <Badge
            variant={statusVariants[invoice.status]}
            className="flex-shrink-0"
          >
            {invoice.status}
          </Badge>
        </div>

        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span className="font-semibold text-foreground">
            {formatCurrency(invoice.amount, invoice.currency)}
          </span>
          {invoice.due_date && (
            <span>Due: {formatDate(invoice.due_date)}</span>
          )}
        </div>

        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          {invoice.company && (
            <span className="flex items-center gap-1">
              <Building2 className="h-3 w-3" />
              {invoice.company.name}
            </span>
          )}
          {invoice.category && (
            <span className="flex items-center gap-1">
              {invoice.category.color && (
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ backgroundColor: invoice.category.color }}
                />
              )}
              <Tag className="h-3 w-3" />
              {invoice.category.name}
            </span>
          )}
          {invoice.receiver && (
            <span className="flex items-center gap-1">
              <User className="h-3 w-3" />
              {invoice.receiver.name}
            </span>
          )}
        </div>
      </div>
    </button>
  );
}
