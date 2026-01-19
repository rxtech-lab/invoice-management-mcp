import { apiClient } from "./client";
import type {
  InvoiceItem,
  CreateInvoiceItemRequest,
  UpdateInvoiceItemRequest,
} from "./types";

export async function addInvoiceItem(
  invoiceId: number,
  data: CreateInvoiceItemRequest
): Promise<InvoiceItem> {
  return apiClient<InvoiceItem>(`/api/invoices/${invoiceId}/items`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateInvoiceItem(
  invoiceId: number,
  itemId: number,
  data: UpdateInvoiceItemRequest
): Promise<InvoiceItem> {
  return apiClient<InvoiceItem>(`/api/invoices/${invoiceId}/items/${itemId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteInvoiceItem(
  invoiceId: number,
  itemId: number
): Promise<void> {
  return apiClient<void>(`/api/invoices/${invoiceId}/items/${itemId}`, {
    method: "DELETE",
  });
}
