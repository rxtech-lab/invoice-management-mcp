"use server";

import { revalidatePath } from "next/cache";
import { apiClient } from "@/lib/api/client";
import type {
  InvoiceItem,
  CreateInvoiceItemRequest,
  UpdateInvoiceItemRequest,
} from "@/lib/api/types";

export async function addInvoiceItemAction(
  invoiceId: number,
  data: CreateInvoiceItemRequest
): Promise<{ success: boolean; data?: InvoiceItem; error?: string }> {
  try {
    const item = await apiClient<InvoiceItem>(`/api/invoices/${invoiceId}/items`, {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath(`/invoices/${invoiceId}`);
    return { success: true, data: item };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to add item",
    };
  }
}

export async function updateInvoiceItemAction(
  invoiceId: number,
  itemId: number,
  data: UpdateInvoiceItemRequest
): Promise<{ success: boolean; data?: InvoiceItem; error?: string }> {
  try {
    const item = await apiClient<InvoiceItem>(
      `/api/invoices/${invoiceId}/items/${itemId}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      }
    );
    revalidatePath(`/invoices/${invoiceId}`);
    return { success: true, data: item };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update item",
    };
  }
}

export async function deleteInvoiceItemAction(
  invoiceId: number,
  itemId: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/invoices/${invoiceId}/items/${itemId}`, {
      method: "DELETE",
    });
    revalidatePath(`/invoices/${invoiceId}`);
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete item",
    };
  }
}
