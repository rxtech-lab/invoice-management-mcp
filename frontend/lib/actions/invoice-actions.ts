"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import { createMCPClient } from '@ai-sdk/mcp';
import { ToolLoopAgent } from 'ai';
import { auth } from "@/auth";
import type {
  Invoice,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceStatus,
} from "@/lib/api/types";
import { invoiceAgentPrompt } from "../prompt";
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

export async function createInvoiceWithAgentAction(fileUrl: string): Promise<{
  success: boolean;
  data?: Invoice;
  error?: string;
}> {
  try {
    const session = await auth();

    if (!session?.accessToken) {
      return {
        success: false,
        error: "Authentication required",
      };
    }

    const url = process.env.NEXT_PUBLIC_API_URL! + "/mcp"

    const httpClient = await createMCPClient({
      transport: {
        type: 'http',
        url: url,
        headers: {
          Authorization: `Bearer ${session.accessToken}`,
        },
      },
    });
    const tools = await httpClient.tools()

    const agent = new ToolLoopAgent({
      model: process.env.INVOICE_AGENT_MODEL!,
      instructions: invoiceAgentPrompt,
      tools: {
        ...tools,
      },
    });

    const result = await agent.generate({
      messages: [{
        role: "user",
        content: [
          {
            type: "text",
            text: `Create an invoice for the following file, make sure to use tools to add the invoice to the database. The original download link is ${fileUrl}, make sure add this link to the invoice`,
          },
          {
            type: 'file',
            data: fileUrl,
            mediaType: 'application/pdf',
            filename: 'invoice.pdf',
          },
        ]
      }]
    });

    console.log("Invoice created successfully", result);

    return {
      success: true,
    }

  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create invoice with agent",
    };
  }
  return {
    success: false,
    error: "Agent invoice creation not yet implemented",
  };
}
