"use server";

import { apiClient } from "@/lib/api/client";
import type { Invoice, PaginatedResponse } from "@/lib/api/types";
import type { InvoiceSearchResult } from "./types";

interface KeywordSearchParams {
  query: string;
  limit?: number;
}

export async function keywordSearch(
  params: KeywordSearchParams
): Promise<InvoiceSearchResult[]> {
  const { query, limit = 10 } = params;

  if (!query.trim()) {
    return [];
  }

  try {
    const response = await apiClient<PaginatedResponse<Invoice>>(
      `/api/invoices?keyword=${encodeURIComponent(query)}&limit=${limit}`
    );

    return response.data.map((invoice) => ({
      id: invoice.id,
      title: invoice.title,
      description: invoice.description,
      amount: invoice.amount,
      currency: invoice.currency,
      status: invoice.status,
      company: invoice.company
        ? { id: invoice.company.id, name: invoice.company.name }
        : undefined,
      category: invoice.category
        ? {
            id: invoice.category.id,
            name: invoice.category.name,
            color: invoice.category.color,
          }
        : undefined,
      receiver: invoice.receiver
        ? { id: invoice.receiver.id, name: invoice.receiver.name }
        : undefined,
      due_date: invoice.due_date,
      created_at: invoice.created_at,
    }));
  } catch (error) {
    console.error("Keyword search failed:", error);
    return [];
  }
}
