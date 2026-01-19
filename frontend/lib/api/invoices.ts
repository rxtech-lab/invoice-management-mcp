import { apiClient } from "./client";
import type {
  Invoice,
  PaginatedResponse,
  InvoiceListOptions,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceStatus,
} from "./types";

export async function getInvoices(
  options: InvoiceListOptions = {}
): Promise<PaginatedResponse<Invoice>> {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const query = params.toString();
  return apiClient<PaginatedResponse<Invoice>>(
    `/api/invoices${query ? `?${query}` : ""}`
  );
}

export async function getInvoice(id: number): Promise<Invoice> {
  return apiClient<Invoice>(`/api/invoices/${id}`);
}

export async function createInvoice(
  data: CreateInvoiceRequest
): Promise<Invoice> {
  return apiClient<Invoice>("/api/invoices", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateInvoice(
  id: number,
  data: UpdateInvoiceRequest
): Promise<Invoice> {
  return apiClient<Invoice>(`/api/invoices/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteInvoice(id: number): Promise<void> {
  return apiClient<void>(`/api/invoices/${id}`, { method: "DELETE" });
}

export async function updateInvoiceStatus(
  id: number,
  status: InvoiceStatus
): Promise<Invoice> {
  return apiClient<Invoice>(`/api/invoices/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}
