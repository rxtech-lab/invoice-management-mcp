"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import { createMCPClient } from '@ai-sdk/mcp';
import { streamText, stepCountIs } from 'ai';
import { createStreamableValue } from '@ai-sdk/rsc';
import { auth } from "@/auth";
import type {
  Invoice,
  CreateInvoiceRequest,
  UpdateInvoiceRequest,
  InvoiceStatus,
} from "@/lib/api/types";
import { invoiceAgentPrompt } from "../prompt";
import { getFileDownloadURLAction } from "./upload-actions";

export type ToolProgress = {
  status: 'idle' | 'calling' | 'complete' | 'error';
  toolName?: string;
  message: string;
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

function formatToolName(toolName: string): string {
  // Convert "create_invoice" to "Creating Invoice"
  const words = toolName.replace(/_/g, ' ').split(' ');
  if (words.length > 0) {
    // Add "ing" to first word (simple approximation)
    const firstWord = words[0];
    if (firstWord.endsWith('e')) {
      words[0] = firstWord.slice(0, -1) + 'ing';
    } else {
      words[0] = firstWord + 'ing';
    }
  }
  return words.map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
}

export async function createInvoiceWithAgentAction(fileKey: string) {
  const session = await auth();

  if (!session?.accessToken) {
    const errorStream = createStreamableValue<ToolProgress>({
      status: 'error',
      message: 'Authentication required',
    });
    errorStream.done();
    return { progress: errorStream.value };
  }

  const progressStream = createStreamableValue<ToolProgress>({
    status: 'idle',
    message: 'Initializing...',
  });

  // Run async to allow returning stream immediately
  (async () => {
    try {
      // Fetch download URL from file key
      progressStream.update({
        status: 'calling',
        message: 'Fetching file URL...',
      });

      const downloadResult = await getFileDownloadURLAction(fileKey);
      if (!downloadResult.success || !downloadResult.data) {
        progressStream.done({
          status: 'error',
          message: downloadResult.error || 'Failed to get file download URL',
        });
        return;
      }

      const fileUrl = downloadResult.data.download_url;

      const url = process.env.NEXT_PUBLIC_API_URL! + "/mcp";

      const httpClient = await createMCPClient({
        transport: {
          type: 'http',
          url: url,
          headers: {
            Authorization: `Bearer ${session.accessToken}`,
          },
        },
      });

      progressStream.update({
        status: 'calling',
        message: 'Connecting to AI agent...',
      });

      const tools = await httpClient.tools();

      const result = streamText({
        model: process.env.INVOICE_AGENT_MODEL! as Parameters<typeof streamText>[0]['model'],
        system: invoiceAgentPrompt,
        messages: [
          {
            role: 'user',
            content: [
              {
                type: 'text',
                text: `Create an invoice for the following file, make sure to use tools to add the invoice to the database. The original_download_link should be set to "${fileKey}" (this is the file key, not a URL)`,
              },
              {
                type: 'file',
                data: new URL(fileUrl),
                mediaType: 'application/pdf',
              },
            ],
          },
        ],
        tools: tools,
        stopWhen: stepCountIs(15),
        onChunk: ({ chunk }) => {
          if (chunk.type === 'tool-call') {
            progressStream.update({
              status: 'calling',
              toolName: chunk.toolName,
              message: `${formatToolName(chunk.toolName)}...`,
            });
          }
        },
      });

      // Wait for completion
      await result.text;

      progressStream.done({
        status: 'complete',
        message: 'Invoice created successfully!',
      });

      revalidatePath('/invoices');
    } catch (error) {
      console.error('Agent error:', error);
      progressStream.done({
        status: 'error',
        message: error instanceof Error ? error.message : 'Failed to create invoice',
      });
    }
  })();

  return { progress: progressStream.value };
}
