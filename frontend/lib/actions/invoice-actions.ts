"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import type {
  Invoice,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceStatus,
} from "@/lib/api/types";

export async function createInvoiceAction(
  data: CreateInvoiceRequest
): Promise<{ success: boolean; data?: Invoice; error?: string }> {
  try {
    const invoice = await apiClient<Invoice>("/api/invoices", {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath("/invoices");
    return { success: true, data: invoice };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create invoice",
    };
  }
}

export async function updateInvoiceAction(
  id: number,
  data: UpdateInvoiceRequest
): Promise<{ success: boolean; data?: Invoice; error?: string }> {
  try {
    const invoice = await apiClient<Invoice>(`/api/invoices/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    revalidatePath("/invoices");
    revalidatePath(`/invoices/${id}`);
    return { success: true, data: invoice };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update invoice",
    };
  }
}

export async function deleteInvoiceAction(
  id: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/invoices/${id}`, { method: "DELETE" });
    revalidatePath("/invoices");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete invoice",
    };
  }
}

export async function updateInvoiceStatusAction(
  id: number,
  status: InvoiceStatus
): Promise<{ success: boolean; data?: Invoice; error?: string }> {
  try {
    const invoice = await apiClient<Invoice>(`/api/invoices/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
    revalidatePath("/invoices");
    revalidatePath(`/invoices/${id}`);
    return { success: true, data: invoice };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update invoice status",
    };
  }
}

export async function createInvoiceAndRedirect(data: CreateInvoiceRequest) {
  const result = await createInvoiceAction(data);
  if (result.success && result.data) {
    redirect(`/invoices/${result.data.id}`);
  }
  return result;
}

export async function deleteInvoiceAndRedirect(id: number) {
  const result = await deleteInvoiceAction(id);
  if (result.success) {
    redirect("/invoices");
  }
  return result;
}
