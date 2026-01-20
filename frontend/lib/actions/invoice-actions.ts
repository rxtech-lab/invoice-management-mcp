"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import { createStreamableValue } from "@ai-sdk/rsc";
import { auth } from "@/auth";
import type {
  Invoice,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceStatus,
} from "@/lib/api/types";
import { getFileDownloadURLAction } from "./upload-actions";
import { runInvoiceAgent } from "../agent/invoice-agent";

// Define types locally to avoid bundling issues with server actions
export type ToolProgress = {
  status: "idle" | "calling" | "complete" | "error";
  toolName?: string;
  message: string;
  invoiceId?: number;
};

// Extended type for server action progress with invoice ID
export type AgentProgress = ToolProgress & {
  invoiceId?: number;
};

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

export async function createInvoiceWithAgentAction(fileKey: string) {
  const session = await auth();

  if (!session?.accessToken) {
    const errorStream = createStreamableValue<AgentProgress>({
      status: "error",
      message: "Authentication required",
    });
    errorStream.done();
    return { progress: errorStream.value };
  }

  const progressStream = createStreamableValue<AgentProgress>({
    status: "idle",
    message: "Initializing...",
  });

  // Run async to allow returning stream immediately
  (async () => {
    try {
      // Fetch download URL from file key
      progressStream.update({
        status: "calling",
        message: "Fetching file URL...",
      });

      const downloadResult = await getFileDownloadURLAction(fileKey);
      if (!downloadResult.success || !downloadResult.data) {
        progressStream.done({
          status: "error",
          message: downloadResult.error || "Failed to get file download URL",
        });
        return;
      }

      const fileUrl = downloadResult.data.download_url;

      // Run the shared invoice agent
      const result = await runInvoiceAgent({
        fileUrl,
        fileKey,
        accessToken: session.accessToken!,
        onProgress: (progress) => {
          progressStream.update(progress);
        },
      });

      // Send completion with invoice ID
      progressStream.done({
        status: "complete",
        message: "Invoice created successfully!",
        invoiceId: result.invoiceId ?? undefined,
      });

      revalidatePath("/invoices");
    } catch (error) {
      console.error("Agent error:", error);
      progressStream.done({
        status: "error",
        message: error instanceof Error ? error.message : "Failed to create invoice",
      });
    }
  })();

  return { progress: progressStream.value };
}
